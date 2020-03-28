package msg

import (
    "fmt"
    "time"
    "errors"
    
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
    
    mt "github.com/mjackson001/mtzone/x/microtick/types"
    "github.com/mjackson001/mtzone/x/microtick/keeper"
)

type TxPickTrade struct {
    Buyer mt.MicrotickAccount
    Id mt.MicrotickId
    TradeType mt.MicrotickTradeType
}

func NewTxPickTrade(buyer sdk.AccAddress, id mt.MicrotickId, tradeType mt.MicrotickTradeType) TxPickTrade {
    return TxPickTrade {
        Buyer: buyer,
        Id: id,
        TradeType: tradeType,
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
    if msg.Buyer.Empty() {
        return errors.New(fmt.Sprintf("Invalid address: %s", msg.Buyer.String()))
    }
    return nil
}

func (msg TxPickTrade) GetSignBytes() []byte {
    return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg TxPickTrade) GetSigners() []sdk.AccAddress {
    return []sdk.AccAddress{ msg.Buyer }
}

// Handler

func HandleTxPickTrade(ctx sdk.Context, mtKeeper keeper.Keeper, msg TxPickTrade) (*sdk.Result, error) {
    params := mtKeeper.GetParams(ctx)
    
    quote, err := mtKeeper.GetActiveQuote(ctx, msg.Id)
    if err != nil {
        return nil, errors.New("No such quote")
    }
    
    // Step 1 - Obtain the strike spot price and create trade struct
    market, err2 := mtKeeper.GetDataMarket(ctx, quote.Market)
    if err2 != nil {
        return nil, errors.New("Error fetching market")
    }
    
    if quote.Provider.Equals(msg.Buyer) {
        return nil, errors.New("Cannot buy owned quote")
    }
    
    commission := mt.NewMicrotickCoinFromDec(params.CommissionTradeFixed)
    settleIncentive := mt.NewMicrotickCoinFromDec(params.SettleIncentive)
    now := ctx.BlockHeader().Time
    trade := keeper.NewDataActiveTrade(now, quote.Market, quote.Duration, msg.TradeType,
        msg.Buyer, market.Consensus, commission, settleIncentive)
        
    matcher := keeper.NewMatcher(trade, nil)
        
    // Step 2 - Compute premium and cost
    var premium mt.MicrotickPremium
    if msg.TradeType == mt.MicrotickCall {
        premium = quote.PremiumAsCall(market.Consensus)
    }
    if msg.TradeType == mt.MicrotickPut {
        premium = quote.PremiumAsPut(market.Consensus)
    }
        
    cost := mt.NewMicrotickCoinFromDec(premium.Amount.Mul(quote.Quantity.Amount))
    
    matcher.TotalQuantity = quote.Quantity.Amount
    matcher.TotalCost = cost
    
    matcher.FillInfo = append(matcher.FillInfo, keeper.QuoteFillInfo {
        Quote: quote,
        BoughtQuantity: quote.Quantity.Amount,
        Cost: cost,
        FinalFill: true,
    })
    
    if matcher.HasQuantity() {
        
        // Step 3 - Deduct premium from buyer account and add it to provider account
        // We do this first because if the funds aren't there we abort
        //fmt.Printf("TotalCost: %s\n", matcher.TotalCost.String())
        //fmt.Printf("Commission: %s\n", trade.Commission.String())
        //fmt.Printf("Settle Incentive: %s\n", settleIncentive.String())
        total := matcher.TotalCost.Add(trade.Commission).Add(settleIncentive)
        err2 = mtKeeper.WithdrawMicrotickCoin(ctx, msg.Buyer, total)
        if err2 != nil {
            return nil, errors.New("Insufficient funds")
        }
        //fmt.Printf("Trade Commission: %s\n", trade.Commission.String())
        //fmt.Printf("Settle Incentive: %s\n", settleIncentive.String())
        mtKeeper.PoolCommission(ctx, msg.Buyer, trade.Commission)
    
        // Step 4 - Finalize trade 
        matcher.Trade.Id = mtKeeper.GetNextActiveTradeId(ctx)
        
        err2 = matcher.AssignCounterparties(ctx, mtKeeper, &market)
        if err2 != nil {
            return nil, errors.New("Error assigning counterparties")
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
        data := PickTradeData {
            Market: quote.Market,
            Duration: mt.MicrotickDurationNameFromDur(quote.Duration),
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
            sdk.NewAttribute("mtm.MarketTick", quote.Market),
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
            
        return &sdk.Result {
            Data: bz,
            Events: events,
        }, nil
        
    } else {
       
        // No liquidity available
        return nil, errors.New("No liquidity available")
        
    }
}