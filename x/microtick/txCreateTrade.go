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
        
    fmt.Printf("Trade: %+v\n", trade)
    
    // Step 2 - Compute premium for quantity requested
    quantity, premium := market.Match(trade, func(id MicrotickId) DataActiveQuote {
        quote, err := keeper.GetActiveQuote(ctx, id)
        if err != nil {
            panic("Invalid quote ID")
        }
        return quote
    }, nil)
    
    fmt.Printf("Quantity: %s\n", quantity.String())
    fmt.Printf("Premium: %s\n", premium.String())
    
    if quantity.Amount.GT(sdk.ZeroDec()) {
        
        // Step 3 - Deduct premium from buyer account and add it to provider account
        keeper.WithdrawDecCoin(ctx, msg.Buyer, NewMicrotickCoinFromPremium(premium))
    
        // Step 4 - Finalize trade 
        trade.Id = keeper.GetNextActiveTradeId(ctx)
        market.Match(trade, func(id MicrotickId) DataActiveQuote {
            quote, err := keeper.GetActiveQuote(ctx, id)
            if err != nil {
                panic("Invalid quote ID")
            }
            return quote
        }, func(quote DataActiveQuote, boughtQuantity sdk.Dec, paidPremium MicrotickPremium) {
            
            // Pay premium
            keeper.DepositDecCoin(ctx, quote.Provider, NewMicrotickCoinFromPremium(paidPremium))
            
            accountStatus := keeper.GetAccountStatus(ctx, quote.Provider)
            
            // Adjust quote
            market.factorOut(quote)
            
            var backing MicrotickCoin
            if boughtQuantity.GTE(quote.Quantity.Amount) {
                backing = quote.Backing
            } else {
                backing = NewMicrotickCoinFromDec(quote.Backing.Amount.Mul(boughtQuantity.Quo(quote.Quantity.Amount)))
            }
            quote.Quantity = NewMicrotickQuantityFromDec(quote.Quantity.Amount.Sub(boughtQuantity))
            quote.Backing = quote.Backing.Minus(backing)
            
            fmt.Printf("Quote Quantity: %s\n", quote.Quantity.String())
            if quote.Quantity.Amount.IsZero() {
                market.DeleteQuote(quote)
                keeper.DeleteActiveQuote(ctx, quote.Id)
                accountStatus.ActiveQuotes.Delete(quote.Id)
            } else {
                market.factorIn(quote)
            }
            
            // Adjust trade
            trade.Backing = trade.Backing.Plus(backing)
            trade.Premium = trade.Premium.Plus(paidPremium)
            trade.FilledQuantity = NewMicrotickQuantityFromDec(trade.FilledQuantity.Amount.Add(boughtQuantity))
            
            params := DataQuoteParams {
                Id: quote.Id,
                Premium: quote.Premium,
                Quantity: quote.Quantity,
                Spot: quote.Spot,
            }
            trade.CounterParties = append(trade.CounterParties, DataCounterParty {
                Backing: backing,
                Premium: paidPremium,
                FilledQuantity: NewMicrotickQuantityFromDec(boughtQuantity),
                Short: quote.Provider,
                Quoted: params,
            })
            
            accountStatus.ActiveTrades.Insert(NewListItem(trade.Id, sdk.NewDec(trade.Expiration.UnixNano())))
            accountStatus.QuoteBacking = accountStatus.QuoteBacking.Minus(backing)
            accountStatus.TradeBacking = accountStatus.TradeBacking.Plus(backing)
            keeper.SetAccountStatus(ctx, quote.Provider, accountStatus)
        })
        
        accountStatus := keeper.GetAccountStatus(ctx, msg.Buyer)
        accountStatus.ActiveTrades.Insert(NewListItem(trade.Id, sdk.NewDec(trade.Expiration.UnixNano())))
        accountStatus.NumTrades++
        keeper.SetAccountStatus(ctx, msg.Buyer, accountStatus)
        
        // Commit changes
        keeper.SetDataMarket(ctx, market)
        keeper.SetActiveTrade(ctx, trade)
    
        tags := sdk.NewTags(
            "id", fmt.Sprintf("%d", trade.Id),
            fmt.Sprintf("trade.%d", trade.Id), "create",
        )
            
        return sdk.Result {
            Tags: tags,
        }
        
    } else {
        
        // No liquidity available
        return sdk.Result {}
        
    }
}
