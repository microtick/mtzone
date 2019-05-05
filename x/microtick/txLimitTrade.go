package microtick

import (
    "fmt"
    
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
)

type TxLimitTrade struct {
    Market MicrotickMarket
    Duration MicrotickDuration
    Buyer sdk.AccAddress
    TradeType MicrotickTradeType
    Limit MicrotickPremium
}

func NewTxLimitTrade(market MicrotickMarket, dur MicrotickDuration, buyer sdk.AccAddress,
    tradeType MicrotickTradeType, limit MicrotickPremium) TxLimitTrade {
        
    return TxLimitTrade {
        Market: market,
        Duration: dur,
        Buyer: buyer,
        TradeType: tradeType,
        
        Limit: limit,
    }
}

type LimitTradeData struct {
    Originator string `json:"originator"`
    Trade DataActiveTrade `json:"trade"`
    Consensus MicrotickSpot `json:"consensus"`
}

func (msg TxLimitTrade) Route() string { return "microtick" }

func (msg TxLimitTrade) Type() string { return "trade_limit" }

func (msg TxLimitTrade) ValidateBasic() sdk.Error {
    if len(msg.Market) == 0 {
        return sdk.ErrInternal("Unknown market")
    }
    if msg.Buyer.Empty() {
        return sdk.ErrInvalidAddress(msg.Buyer.String())
    }
    return nil
}

func (msg TxLimitTrade) GetSignBytes() []byte {
    return sdk.MustSortJSON(msgCdc.MustMarshalJSON(msg))
}

func (msg TxLimitTrade) GetSigners() []sdk.AccAddress {
    return []sdk.AccAddress{ msg.Buyer }
}

// Handler

func handleTxLimitTrade(ctx sdk.Context, keeper Keeper, msg TxLimitTrade) sdk.Result {
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
    trade := NewDataActiveTrade(msg.Market, msg.Duration, msg.TradeType,
        msg.Buyer, market.Consensus)
        
    matcher := NewMatcher(trade, func (id MicrotickId) DataActiveQuote {
        quote, err := keeper.GetActiveQuote(ctx, id)
        if err != nil {
            // This function should always be called with an active quote
            panic("Invalid quote ID")
        }
        return quote
    })
        
    // Step 2 - Compute premium for quantity requested
    market.MatchByLimit(&matcher, msg.Limit)
    
    if matcher.hasQuantity() {
        
        // Step 3 - Deduct premium from buyer account and add it to provider account
        // We do this first because if the funds aren't there we abort
        keeper.WithdrawMicrotickCoin(ctx, msg.Buyer, NewMicrotickCoinFromDec(matcher.TotalCost))
    
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
        keeper.SetActiveTrade(ctx, matcher.Trade)
    
        tags := sdk.NewTags(
            "mtm.NewTrade", fmt.Sprintf("%d", matcher.Trade.Id),
            fmt.Sprintf("trade.%d", matcher.Trade.Id), "create",
            fmt.Sprintf("acct.%s", msg.Buyer), "trade.long",
            "mtm.MarketTick", msg.Market,
        )
        
        for i := 0; i < len(matcher.FillInfo); i++ {
            thisFill := matcher.FillInfo[i]
            
            tags = tags.AppendTag(fmt.Sprintf("acct.%s", thisFill.Quote.Provider), "trade.short")
            
            quoteKey := fmt.Sprintf("quote.%d", thisFill.Quote.Id)
            if thisFill.FinalFill {
                tags = tags.AppendTag(quoteKey, "final")
            } else {
                // should never get here, but in case logic changes for filling limit order
                tags = tags.AppendTag(quoteKey, "match")
            }
        }
        
        // Data
        data := LimitTradeData {
            Originator: "limitTrade",
            Consensus: market.Consensus,
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
