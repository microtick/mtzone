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

type TxMarketTrade struct {
    Market mt.MicrotickMarket
    Duration mt.MicrotickDurationName
    Taker mt.MicrotickAccount
    OrderType mt.MicrotickOrderType
    Quantity mt.MicrotickQuantity
}

func NewTxMarketTrade(market mt.MicrotickMarket, dur mt.MicrotickDurationName, taker sdk.AccAddress,
    orderType mt.MicrotickOrderType, quantity mt.MicrotickQuantity) TxMarketTrade {
        
    return TxMarketTrade {
        Market: market,
        Duration: dur,
        Taker: taker,
        OrderType: orderType,
        Quantity: quantity,
    }
}

type MarketTradeData struct {
    Market mt.MicrotickMarket `json:"market"`
    Duration mt.MicrotickDurationName `json:"duration"`
    Trade keeper.DataActiveTrade `json:"trade"`
    Consensus mt.MicrotickSpot `json:"consensus"`
    Time time.Time `json:"time"`
}

func (msg TxMarketTrade) Route() string { return "microtick" }

func (msg TxMarketTrade) Type() string { return "trade_market" }

func (msg TxMarketTrade) ValidateBasic() error {
    if msg.OrderType != mt.MicrotickOrderBuyCall &&
        msg.OrderType != mt.MicrotickOrderSellCall &&
        msg.OrderType != mt.MicrotickOrderBuyPut &&
        msg.OrderType != mt.MicrotickOrderSellPut &&
        msg.OrderType != mt.MicrotickOrderBuySyn &&
        msg.OrderType != mt.MicrotickOrderSellSyn {
        return sdkerrors.Wrap(mt.ErrInvalidOrderType, msg.OrderType)
    }
    if msg.Market == "" {
        return sdkerrors.Wrap(mt.ErrMissingParam, "market")
    }
    if msg.Taker.Empty() {
        return sdkerrors.Wrap(mt.ErrInvalidAddress, msg.Taker.String())
    }
    return nil
}

func (msg TxMarketTrade) GetSignBytes() []byte {
    return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg TxMarketTrade) GetSigners() []sdk.AccAddress {
    return []sdk.AccAddress{ msg.Taker }
}

// Handler

func HandleTxMarketTrade(ctx sdk.Context, mtKeeper keeper.Keeper, params mt.Params,
    msg TxMarketTrade) (*sdk.Result, error) {
     
    if !mtKeeper.HasDataMarket(ctx, msg.Market) {
        return nil, sdkerrors.Wrap(mt.ErrInvalidMarket, msg.Market)
    }
    
    if !mtKeeper.ValidDurationName(ctx, msg.Duration) {
        return nil, sdkerrors.Wrapf(mt.ErrInvalidDuration, "%s", msg.Duration)
    }
    
    // Step 1 - Obtain the strike spot price and create trade struct
    market, err := mtKeeper.GetDataMarket(ctx, msg.Market)
    if err != nil {
        return nil, mt.ErrInvalidMarket
    }
    commission := mt.NewMicrotickCoinFromDec(params.CommissionTradeFixed)
    settleIncentive := mt.NewMicrotickCoinFromDec(params.SettleIncentive)
    now := ctx.BlockHeader().Time
    seconds := mtKeeper.DurationFromName(ctx, msg.Duration)
    trade := keeper.NewDataActiveTrade(now, msg.Market, msg.Duration, seconds,
        msg.OrderType, msg.Taker, msg.Quantity, market.Consensus, commission, settleIncentive)
        
    matcher := keeper.NewMatcher(trade, func (id mt.MicrotickId) keeper.DataActiveQuote {
        quote, err := mtKeeper.GetActiveQuote(ctx, id)
        if err != nil {
            // This function should always be called with an active quote
            panic("Invalid quote ID")
        }
        return quote
    })
        
    // Step 2 - Compute premium for quantity requested
    if msg.OrderType == mt.MicrotickOrderBuyCall || msg.OrderType == mt.MicrotickOrderSellCall ||
        msg.OrderType == mt.MicrotickOrderBuyPut || msg.OrderType == mt.MicrotickOrderSellPut {
            
        err = matcher.MatchByQuantity(&market, msg.OrderType, msg.Quantity)
        if err != nil {
            return nil, err
        }
    } else {
        syntheticBook := mtKeeper.GetSyntheticBook(ctx, &market, msg.Duration, &msg.Taker)
        err = matcher.MatchSynthetic(&syntheticBook, &market, msg.Quantity)
        if err != nil {
            return nil, err
        }
    }
    
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
            return nil, err
        }
        
        accountStatus := mtKeeper.GetAccountStatus(ctx, msg.Taker)
        accountStatus.SettleBacking = accountStatus.SettleBacking.Add(settleIncentive)
        accountStatus.NumTrades++
        mtKeeper.SetAccountStatus(ctx, msg.Taker, accountStatus)
        
        // Save
        mtKeeper.CommitTradeId(ctx, matcher.Trade.Id)
        mtKeeper.SetDataMarket(ctx, market)
        mtKeeper.SetActiveTrade(ctx, matcher.Trade)
        
        // Data
        data := MarketTradeData {
            Market: msg.Market,
            Duration: msg.Duration,
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
            sdk.NewAttribute("mtm.MarketTick", msg.Market),
        ), sdk.NewEvent(
            sdk.EventTypeMessage,
            sdk.NewAttribute("commission", commission.String()),
            sdk.NewAttribute("reward", reward.String()),
        ))
        
        for _, leg := range trade.Legs {
            var maker mt.MicrotickAccount
            if leg.Long.Equals(msg.Taker) {
                maker = leg.Short
            } else {
                maker = leg.Long
            }
            quoteKey := fmt.Sprintf("quote.%d", leg.Quoted.Id)
            matchType := "event.match"
            if leg.Quoted.Final {
                matchType = "event.final"
            }
            events = append(events, sdk.NewEvent(
                sdk.EventTypeMessage,
                sdk.NewAttribute(fmt.Sprintf("acct.%s", maker), "trade.start"),
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
