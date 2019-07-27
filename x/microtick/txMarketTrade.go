package microtick

import (
    "fmt"
    "time"
    
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
)

type TxMarketTrade struct {
    Market MicrotickMarket
    Duration MicrotickDuration
    Buyer MicrotickAccount
    TradeType MicrotickTradeType
    Quantity MicrotickQuantity
}

func NewTxMarketTrade(market MicrotickMarket, dur MicrotickDuration, buyer sdk.AccAddress,
    tradeType MicrotickTradeType, quantity MicrotickQuantity) TxMarketTrade {
        
    return TxMarketTrade {
        Market: market,
        Duration: dur,
        Buyer: buyer,
        TradeType: tradeType,
        Quantity: quantity,
    }
}

type MarketTradeData struct {
    Originator string `json:"originator"`
    Trade DataActiveTrade `json:"trade"`
    Consensus MicrotickSpot `json:"consensus"`
    Time time.Time `json:"time"`
}

func (msg TxMarketTrade) Route() string { return "microtick" }

func (msg TxMarketTrade) Type() string { return "trade_market" }

func (msg TxMarketTrade) ValidateBasic() sdk.Error {
    if len(msg.Market) == 0 {
        return sdk.ErrInternal("Unknown market")
    }
    if msg.Buyer.Empty() {
        return sdk.ErrInvalidAddress(msg.Buyer.String())
    }
    return nil
}

func (msg TxMarketTrade) GetSignBytes() []byte {
    return sdk.MustSortJSON(msgCdc.MustMarshalJSON(msg))
}

func (msg TxMarketTrade) GetSigners() []sdk.AccAddress {
    return []sdk.AccAddress{ msg.Buyer }
}

// Handler

func handleTxMarketTrade(ctx sdk.Context, keeper Keeper, msg TxMarketTrade) sdk.Result {
    params := keeper.GetParams(ctx)
     
    if !keeper.HasDataMarket(ctx, msg.Market) {
        return sdk.ErrInternal("No such market: " + msg.Market).Result()
    }
    
    if !ValidMicrotickDuration(msg.Duration) {
        return sdk.ErrInternal(fmt.Sprintf("Invalid duration: %d", msg.Duration)).Result()
    }
    
    // Step 1 - Obtain the strike spot price and create trade struct
    market, err2 := keeper.GetDataMarket(ctx, msg.Market)
    if err2 != nil {
        return sdk.ErrInternal("Error fetching market").Result()
    }
    commission := NewMicrotickCoinFromDec(params.CommissionTradeFixed)
    settleIncentive := NewMicrotickCoinFromDec(params.SettleIncentive)
    now := ctx.BlockHeader().Time
    trade := NewDataActiveTrade(now, msg.Market, msg.Duration, msg.TradeType,
        msg.Buyer, market.Consensus, commission, settleIncentive)
        
    matcher := NewMatcher(trade, func (id MicrotickId) DataActiveQuote {
        quote, err := keeper.GetActiveQuote(ctx, id)
        if err != nil {
            // This function should always be called with an active quote
            panic("Invalid quote ID")
        }
        return quote
    })
        
    // Step 2 - Compute premium for quantity requested
    market.MatchByQuantity(&matcher, msg.Quantity)
    
    if matcher.hasQuantity() {
        
        // Step 3 - Deduct premium from buyer account and add it to provider account
        // We do this first because if the funds aren't there we abort
        total := NewMicrotickCoinFromDec(matcher.TotalCost.Add(trade.Commission.Amount).Add(settleIncentive.Amount))
        keeper.WithdrawMicrotickCoin(ctx, msg.Buyer, total)
        fmt.Printf("Trade Commission: %s\n", trade.Commission.String())
        fmt.Printf("Settle Incentive: %s\n", settleIncentive.String())
        keeper.PoolCommission(ctx, trade.Commission)
    
        // Step 4 - Finalize trade 
        matcher.Trade.Id = keeper.GetNextActiveTradeId(ctx)
        
        matcher.AssignCounterparties(ctx, keeper, &market)
        
        // Update the account status for the buyer
        accountStatus := keeper.GetAccountStatus(ctx, msg.Buyer)
        accountStatus.ActiveTrades.Insert(NewListItem(matcher.Trade.Id, sdk.NewDec(matcher.Trade.Expiration.UnixNano())))
        accountStatus.NumTrades++
        
        // Commit changes
        keeper.SetAccountStatus(ctx, msg.Buyer, accountStatus)
        keeper.SetDataMarket(ctx, market)
        
        matcher.Trade.Balance = keeper.GetTotalBalance(ctx, msg.Buyer)
        keeper.SetActiveTrade(ctx, matcher.Trade)
    
        tags := sdk.NewTags(
            "mtm.NewTrade", fmt.Sprintf("%d", matcher.Trade.Id),
            fmt.Sprintf("trade.%d", matcher.Trade.Id), "event.create",
            fmt.Sprintf("acct.%s", msg.Buyer), "trade.long",
            "mtm.MarketTick", msg.Market,
        )
        
        for i := 0; i < len(matcher.FillInfo); i++ {
            thisFill := matcher.FillInfo[i]
            
            tags = tags.AppendTag(fmt.Sprintf("acct.%s", thisFill.Quote.Provider), "trade.short")
            
            quoteKey := fmt.Sprintf("quote.%d", thisFill.Quote.Id)
            if thisFill.FinalFill {
                tags = tags.AppendTag(quoteKey, "event.final")
            } else {
                tags = tags.AppendTag(quoteKey, "event.match")
            }
        }
        
        // Data
        data := MarketTradeData {
            Originator: "marketTrade",
            Consensus: market.Consensus,
            Time: now,
            Trade: matcher.Trade,
        }
        bz, _ := codec.MarshalJSONIndent(keeper.cdc, data)
            
        return sdk.Result {
            Data: bz,
            Tags: tags,
        }
        
    } else {
       
        // No liquidity available
        return sdk.ErrInternal("No liquidity available").Result()
        
    }
}
