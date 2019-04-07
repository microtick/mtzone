package microtick

import (
    "fmt"
    "encoding/json"
    
    sdk "github.com/cosmos/cosmos-sdk/types"
)

type TxTrade struct {
    Market MicrotickMarket
    Duration MicrotickDuration
    Buyer sdk.AccAddress
    TradeType MicrotickTradeType
    Quantity MicrotickQuantity
}

func NewTxTrade(market MicrotickMarket, dur MicrotickDuration, buyer sdk.AccAddress,
    tradeType MicrotickTradeType, quantity MicrotickQuantity) TxTrade {
        
    return TxTrade {
        Market: market,
        Duration: dur,
        Buyer: buyer,
        TradeType: tradeType,
        Quantity: quantity,
    }
}

func (msg TxTrade) Route() string { return "microtick" }

func (msg TxTrade) Type() string { return "create_trade" }

func (msg TxTrade) ValidateBasic() sdk.Error {
    if len(msg.Market) == 0 {
        return sdk.ErrInternal("Unknown market")
    }
    if msg.Buyer.Empty() {
        return sdk.ErrInvalidAddress(msg.Buyer.String())
    }
    return nil
}

func (msg TxTrade) GetSignBytes() []byte {
    b, err := json.Marshal(msg)
    if err != nil {
        panic(err)
    }
    return sdk.MustSortJSON(b)
}

func (msg TxTrade) GetSigners() []sdk.AccAddress {
    return []sdk.AccAddress{ msg.Buyer }
}

// Handler

func handleTxTrade(ctx sdk.Context, keeper Keeper, msg TxTrade) sdk.Result {
    if !keeper.HasDataMarket(ctx, msg.Market) {
        return sdk.ErrInternal("No such market: " + msg.Market).Result()
    }
    
    if !ValidMicrotickDuration(msg.Duration) {
        return sdk.ErrInternal(fmt.Sprintf("Invalid duration: %d", msg.Duration)).Result()
    }
    
    // Step 1 - Obtain the strike spot price and create trade struct
    market, err2 := keeper.GetDataMarket(ctx, msg.Market)
    if err2 != nil {
        panic("Error fetching market")
    }
    trade := NewDataActiveTrade(msg.Market, msg.Duration, msg.TradeType,
        msg.Buyer, market.Consensus, msg.Quantity)
        
    matcher := NewMatcher(trade, func (id MicrotickId) DataActiveQuote {
        quote, err := keeper.GetActiveQuote(ctx, id)
        if err != nil {
            panic("Invalid quote ID")
        }
        return quote
    })
        
    // Step 2 - Compute premium for quantity requested
    market.MatchByQuantity(&matcher)
    
    if matcher.hasQuantity() {
        
        // Step 3 - Deduct premium from buyer account and add it to provider account
        // We do this first because if the funds aren't there we abort
        keeper.WithdrawDecCoin(ctx, msg.Buyer, NewMicrotickCoinFromDec(matcher.TotalPremium))
    
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
            "id", fmt.Sprintf("%d", matcher.Trade.Id),
            fmt.Sprintf("trade.%d", matcher.Trade.Id), "create",
        )
            
        return sdk.Result {
            Tags: tags,
        }
        
    } else {
        
        // No liquidity available
        return sdk.Result {}
        
    }
}
