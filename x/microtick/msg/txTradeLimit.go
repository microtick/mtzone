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

type TxLimitTrade struct {
    Market mt.MicrotickMarket
    Duration mt.MicrotickDuration
    Buyer mt.MicrotickAccount
    TradeType mt.MicrotickTradeTypeName
    Limit mt.MicrotickPremium
    MaxCost mt.MicrotickCoin
}

func NewTxLimitTrade(market mt.MicrotickMarket, dur mt.MicrotickDuration, buyer sdk.AccAddress,
    tradeType mt.MicrotickTradeType, limit mt.MicrotickPremium, maxCost mt.MicrotickCoin) TxLimitTrade {
        
    return TxLimitTrade {
        Market: market,
        Duration: dur,
        Buyer: buyer,
        TradeType: tradeType,
        
        Limit: limit,
        MaxCost: maxCost,
    }
}

type LimitTradeData struct {
    Market mt.MicrotickMarket `json:"market"`
    Duration mt.MicrotickDurationName `json:"duration"`
    Trade keeper.DataActiveTrade `json:"trade"`
    Consensus mt.MicrotickSpot `json:"consensus"`
    Time time.Time `json:"time"`
}

func (msg TxLimitTrade) Route() string { return "microtick" }

func (msg TxLimitTrade) Type() string { return "trade_limit" }

func (msg TxLimitTrade) ValidateBasic() error {
    if len(msg.Market) == 0 {
        return sdkerrors.Wrap(mt.ErrInvalidMarket, msg.Market)
    }
    if msg.Buyer.Empty() {
        return sdkerrors.Wrap(mt.ErrInvalidAddress, msg.Buyer.String())
    }
    return nil
}

func (msg TxLimitTrade) GetSignBytes() []byte {
    return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg TxLimitTrade) GetSigners() []sdk.AccAddress {
    return []sdk.AccAddress{ msg.Buyer }
}

// Handler

func HandleTxLimitTrade(ctx sdk.Context, mtKeeper keeper.Keeper, msg TxLimitTrade) (*sdk.Result, error) {
    params := mtKeeper.GetParams(ctx)
     
    if !mtKeeper.HasDataMarket(ctx, msg.Market) {
        return nil, sdkerrors.Wrap(mt.ErrInvalidMarket, msg.Market)
    }
    
    if !mt.ValidMicrotickDuration(msg.Duration) {
        return nil, sdkerrors.Wrapf(mt.ErrInvalidDuration, "%d", msg.Duration)
    }
    
    // Step 1 - Obtain the strike spot price and create trade struct
    market, err2 := mtKeeper.GetDataMarket(ctx, msg.Market)
    if err2 != nil {
        return nil, sdkerrors.Wrap(mt.ErrInvalidMarket, msg.Market)
    }
    commission := mt.NewMicrotickCoinFromDec(params.CommissionTradeFixed)
    settleIncentive := mt.NewMicrotickCoinFromDec(params.SettleIncentive)
    now := ctx.BlockHeader().Time
    trade := keeper.NewDataActiveTrade(now, msg.Market, msg.Duration, msg.TradeType,
        msg.Buyer, market.Consensus, commission, settleIncentive)
        
    matcher := keeper.NewMatcher(trade, func (id mt.MicrotickId) keeper.DataActiveQuote {
        quote, err := mtKeeper.GetActiveQuote(ctx, id)
        if err != nil {
            // This function should always be called with an active quote
            panic("Invalid quote ID")
        }
        return quote
    })
        
    // Step 2 - Compute premium for quantity requested
    market.MatchByLimit(&matcher, msg.Limit, msg.MaxCost)
    
    if matcher.HasQuantity() {
        
        // Step 3 - Deduct premium from buyer account and add it to provider account
        // We do this first because if the funds aren't there we abort
        total := matcher.TotalCost.Add(trade.Commission).Add(settleIncentive)
        err2 = mtKeeper.WithdrawMicrotickCoin(ctx, msg.Buyer, total)
        if err2 != nil {
            return nil, mt.ErrInsufficientFunds
        }
        //fmt.Printf("Trade Commission: %s\n", trade.Commission.String())
        //fmt.Printf("Settle Incentive: %s\n", settleIncentive.String())
        mtKeeper.PoolCommission(ctx, msg.Buyer, trade.Commission)
    
        // Step 4 - Finalize trade 
        matcher.Trade.Id = mtKeeper.GetNextActiveTradeId(ctx)
        
        err2 = matcher.AssignCounterparties(ctx, mtKeeper, &market)
        if err2 != nil {
            return nil, sdkerrors.Wrap(mt.ErrTradeMatch, "counterparty assignment")
        }
        
        // Update the account status for the buyer
        accountStatus := mtKeeper.GetAccountStatus(ctx, msg.Buyer)
        accountStatus.ActiveTrades.Insert(keeper.NewListItem(matcher.Trade.Id, sdk.NewDec(matcher.Trade.Expiration.UnixNano())))
        accountStatus.SettleBacking = accountStatus.SettleBacking.Add(settleIncentive)
        accountStatus.NumTrades++
        
        // Commit changes
        mtKeeper.SetAccountStatus(ctx, msg.Buyer, accountStatus)
        mtKeeper.SetDataMarket(ctx, market)
        
        mtKeeper.SetActiveTrade(ctx, matcher.Trade)
    
        // Data
        data := LimitTradeData {
            Market: msg.Market,
            Duration: mt.MicrotickDurationNameFromDur(msg.Duration),
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
            sdk.NewAttribute(fmt.Sprintf("acct.%s", msg.Buyer), "trade.long"),
            sdk.NewAttribute("mtm.MarketTick", msg.Market),
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
                sdk.NewAttribute(fmt.Sprintf("acct.%s", thisFill.Quote.Provider), "trade.short"),
                sdk.NewAttribute(quoteKey, matchType),
            ))
        }
        
        ctx.EventManager().EmitEvents(events)
            
        return &sdk.Result {
            Data: bz,
            Events: ctx.EventManager().ABCIEvents(),
        }, nil
        
    } else {
        
        // No liquidity available
        return nil, sdkerrors.Wrap(mt.ErrTradeMatch, "no liquidity available")
        
    }
}
