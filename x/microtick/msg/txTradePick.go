package msg

import (
    "fmt"
    "time"
    
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    
    mt "github.com/mjackson001/mtzone/x/microtick/types"
    "github.com/mjackson001/mtzone/x/microtick/keeper"
)

type TxPickTrade struct {
    Taker mt.MicrotickAccount
    Id mt.MicrotickId
    OrderType mt.MicrotickOrderType
}

func NewTxPickTrade(taker sdk.AccAddress, id mt.MicrotickId, orderType mt.MicrotickOrderType) TxPickTrade {
    return TxPickTrade {
        Taker: taker,
        Id: id,
        OrderType: orderType,
    }
}

type PickTradeData struct {
    Market mt.MicrotickMarket `json:"market"`
    Duration mt.MicrotickDurationName `json:"duration"`
    Trade keeper.DataActiveTrade `json:"trade"`
    Consensus mt.MicrotickSpot `json:"consensus"`
    Time time.Time `json:"time"`
}

func (msg TxPickTrade) Route() string { return "microtick" }

func (msg TxPickTrade) Type() string { return "trade_pick" }

func (msg TxPickTrade) ValidateBasic() error {
    if msg.OrderType != mt.MicrotickOrderBuyCall &&
        msg.OrderType != mt.MicrotickOrderSellCall &&
        msg.OrderType != mt.MicrotickOrderBuyPut &&
        msg.OrderType != mt.MicrotickOrderSellPut {
        return sdkerrors.Wrap(mt.ErrInvalidOrderType, msg.OrderType)
    }
    if msg.Taker.Empty() {
        return sdkerrors.Wrap(mt.ErrInvalidAddress, msg.Taker.String())
    }
    return nil
}

func (msg TxPickTrade) GetSignBytes() []byte {
    return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg TxPickTrade) GetSigners() []sdk.AccAddress {
    return []sdk.AccAddress{ msg.Taker }
}

// Handler

func HandleTxPickTrade(ctx sdk.Context, mtKeeper keeper.Keeper, params mt.Params,
    msg TxPickTrade) (*sdk.Result, error) {
    
    quote, err := mtKeeper.GetActiveQuote(ctx, msg.Id)
    if err != nil {
        return nil, sdkerrors.Wrapf(mt.ErrInvalidQuote, "%d", msg.Id)
    }
    
    // Step 1 - Obtain the strike spot price and create trade struct
    market, err := mtKeeper.GetDataMarket(ctx, quote.Market)
    if err != nil {
        return nil, sdkerrors.Wrap(mt.ErrInvalidMarket, quote.Market)
    }
    
    if quote.Provider.Equals(msg.Taker) {
        return nil, sdkerrors.Wrap(mt.ErrTradeMatch, "already owner")
    }
    
    commission := mt.NewMicrotickCoinFromDec(params.CommissionTradeFixed)
    settleIncentive := mt.NewMicrotickCoinFromDec(params.SettleIncentive)
    now := ctx.BlockHeader().Time
    durName := mtKeeper.NameFromDuration(ctx, quote.Duration)
    trade := keeper.NewDataActiveTrade(now, quote.Market, durName, mtKeeper.DurationFromName(ctx, durName),
        msg.OrderType, msg.Taker, market.Consensus, commission, settleIncentive)
        
    matcher := keeper.NewMatcher(trade, nil)
    
    // Step 2 - Compute premium and cost
    matcher.MatchQuote(quote)
    
    if matcher.HasQuantity {
        
        // Deduct commission and settle incentive
        total := trade.Commission.Add(settleIncentive)
        err = mtKeeper.WithdrawMicrotickCoin(ctx, msg.Taker, total)
        if err != nil {
            return nil, mt.ErrInsufficientFunds
        }
        reward, err := mtKeeper.PoolCommission(ctx, msg.Taker, trade.Commission)
        if err != nil {
            return nil, err
        }
    
        // Step 4 - Finalize trade 
        matcher.Trade.Id = mtKeeper.GetNextActiveTradeId(ctx)
        
        err = matcher.AssignCounterparties(ctx, mtKeeper, &market)
        if err != nil {
            return nil, sdkerrors.Wrap(mt.ErrTradeMatch, "counterparty assignment")
        }
        
        accountStatus := mtKeeper.GetAccountStatus(ctx, msg.Taker)
        accountStatus.SettleBacking = accountStatus.SettleBacking.Add(settleIncentive)
        accountStatus.NumTrades++
        mtKeeper.SetAccountStatus(ctx, msg.Taker, accountStatus)
        
        // Save
        mtKeeper.SetDataMarket(ctx, market)
        mtKeeper.SetActiveTrade(ctx, matcher.Trade)
        
        // Data
        data := PickTradeData {
            Market: quote.Market,
            Duration: quote.DurationName,
            Consensus: market.Consensus,
            Time: now,
            Trade: matcher.Trade,
        }
        bz, _ := codec.MarshalJSONIndent(ModuleCdc, data)
        
        var events []sdk.Event
        events = append(events, sdk.NewEvent(
            sdk.EventTypeMessage,
            sdk.NewAttribute(sdk.AttributeKeyModule, mt.ModuleKey),
        ), sdk.NewEvent(
            sdk.EventTypeMessage,
            sdk.NewAttribute("mtm.NewTrade", fmt.Sprintf("%d", matcher.Trade.Id)),
            sdk.NewAttribute(fmt.Sprintf("trade.%d", matcher.Trade.Id), "event.create"),
            sdk.NewAttribute(fmt.Sprintf("acct.%s", msg.Taker), "trade.start"),
            sdk.NewAttribute("mtm.MarketTick", quote.Market),
        ), sdk.NewEvent(
            sdk.EventTypeMessage,
            sdk.NewAttribute("commission", commission.String()),
            sdk.NewAttribute("reward", reward.String()),
        ))
        
        for i := 0; i < len(matcher.FillInfo); i++ {
            thisFill := matcher.FillInfo[i]
            
            quoteKey := fmt.Sprintf("quote.%d", thisFill.Quote.Id)
            matchType := "event.match"
            if thisFill.FinalFill {
                matchType = "event.final"
            }
            
            events = append(events, sdk.NewEvent(
                sdk.EventTypeMessage,
                sdk.NewAttribute(fmt.Sprintf("acct.%s", thisFill.Quote.Provider), "trade.start"),
                sdk.NewAttribute(quoteKey, matchType),
            ))
        }
        
        ctx.EventManager().EmitEvents(events)
            
        return &sdk.Result {
            Data: bz,
            Events: ctx.EventManager().ABCIEvents(),
        }, nil
        
    }
       
    // No liquidity available
    return nil, sdkerrors.Wrap(mt.ErrTradeMatch, "no liquidity available")
}
