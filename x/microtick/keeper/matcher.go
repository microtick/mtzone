package keeper

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
    
    mt "github.com/mjackson001/mtzone/x/microtick/types"
)

type QuoteFillInfo struct {
    Quote DataActiveQuote
    BuySell bool
    LegType mt.MicrotickLegType
    Quantity sdk.Dec
    Cost mt.MicrotickCoin
    Backing mt.MicrotickCoin
    FinalFill bool
}

func NewQuoteFillInfo(quote DataActiveQuote, buySell bool, legType mt.MicrotickLegType, 
    quantity sdk.Dec, cost mt.MicrotickCoin, backing mt.MicrotickCoin, finalFill bool) QuoteFillInfo {
    return QuoteFillInfo {
        Quote: quote,
        BuySell: buySell,
        LegType: legType,
        Quantity: quantity,
        Cost: cost,
        Backing: backing,
        FinalFill: finalFill,
    }
}

type FetchQuoteFunc func(mt.MicrotickId) DataActiveQuote

type Matcher struct {
    Trade DataActiveTrade
    HasQuantity bool
    FillInfo []QuoteFillInfo
    FetchQuote FetchQuoteFunc
}

func NewMatcher(trade DataActiveTrade, fetchQuoteFunc FetchQuoteFunc) Matcher {
    return Matcher {
        Trade: trade,
        HasQuantity: false,
        FetchQuote: fetchQuoteFunc,
    }
}

func (matcher *Matcher) MatchByQuantity(dm *DataMarket, totalQuantity mt.MicrotickQuantity) {
    orderBook := dm.GetOrderBook(matcher.Trade.DurationName)
    quantityToMatch := totalQuantity.Amount
    
    var list []mt.MicrotickId
    if matcher.Trade.Order == mt.MicrotickOrderBuyCall {
        for i := 0; i < len(orderBook.CallAsks.Data); i++ {
            list = append(list, orderBook.CallAsks.Data[i].Id)
        }
    }
    if matcher.Trade.Order == mt.MicrotickOrderBuyPut {
        for i := 0; i < len(orderBook.PutAsks.Data); i++ {
            list = append(list, orderBook.PutAsks.Data[i].Id)
        }
    }
    if matcher.Trade.Order == mt.MicrotickOrderSellCall {
        for i := 0; i < len(orderBook.CallBids.Data); i++ {
            j := len(orderBook.CallBids.Data) - i - 1
            list = append([]mt.MicrotickId{orderBook.CallBids.Data[j].Id}, list...)
        }
    }
    if matcher.Trade.Order == mt.MicrotickOrderSellPut {
        for i := 0; i < len(orderBook.PutBids.Data); i++ {
            j := len(orderBook.PutBids.Data) - i - 1
            list = append([]mt.MicrotickId{orderBook.PutBids.Data[j].Id}, list...)
        }
    }
    
    index := 0
    for index < len(list) && quantityToMatch.GT(sdk.ZeroDec()) {
        id := list[index]
        quote := matcher.FetchQuote(id)
        if !quote.Provider.Equals(matcher.Trade.Taker) {
            var premium mt.MicrotickPremium
            var buysell bool
            var legType mt.MicrotickLegType
            if matcher.Trade.Order == mt.MicrotickOrderBuyCall {
                buysell = true
                legType = mt.MicrotickLegCall
                premium = quote.CallAsk(matcher.Trade.Strike)
            }
            if matcher.Trade.Order == mt.MicrotickOrderBuyPut {
                buysell = true
                legType = mt.MicrotickLegPut
                premium = quote.PutAsk(matcher.Trade.Strike)
            }
            if matcher.Trade.Order == mt.MicrotickOrderSellCall {
                buysell = false
                legType = mt.MicrotickLegCall
                premium = quote.CallBid(matcher.Trade.Strike)
            }
            if matcher.Trade.Order == mt.MicrotickOrderSellPut {
                buysell = false
                legType = mt.MicrotickLegPut
                premium = quote.PutBid(matcher.Trade.Strike)
            }
            
            var quantity sdk.Dec
            finalFill := false
            
            if quote.Quantity.Amount.GTE(quantityToMatch) {
                quantity = quantityToMatch
                quantityToMatch = sdk.ZeroDec()
            } else {
                quantity = quote.Quantity.Amount
                quantityToMatch = quantityToMatch.Sub(quote.Quantity.Amount)
                finalFill = true
            }
            
            var backing mt.MicrotickCoin
            if finalFill {
                backing = quote.Backing
            } else {
                // For a partially filled quote, the backing transferred should be proportional to
                // the quantity purchased.
                backing = mt.NewMicrotickCoinFromDec(quote.Backing.Amount.Mul(quantity.Quo(quote.Quantity.Amount)))
            }
            
            matcher.HasQuantity = true
            cost := premium.Amount.Mul(quantity)
            
            matcher.FillInfo = append(matcher.FillInfo, NewQuoteFillInfo(quote, buysell, legType, quantity,
                mt.NewMicrotickCoinFromDec(cost), backing, finalFill))
        }
        index++
    }
}

func (matcher *Matcher) MatchQuote(quote DataActiveQuote) {
    var premium mt.MicrotickPremium
    var buysell bool
    var legType mt.MicrotickLegType
    if matcher.Trade.Order == mt.MicrotickOrderBuyCall {
        buysell = true
        legType = mt.MicrotickLegCall
        premium = quote.CallAsk(matcher.Trade.Strike)
    }
    if matcher.Trade.Order == mt.MicrotickOrderBuyPut {
        buysell = true
        legType = mt.MicrotickLegPut
        premium = quote.PutAsk(matcher.Trade.Strike)
    }
    if matcher.Trade.Order == mt.MicrotickOrderSellCall {
        buysell = false
        legType = mt.MicrotickLegCall
        premium = quote.CallBid(matcher.Trade.Strike)
    }
    if matcher.Trade.Order == mt.MicrotickOrderSellPut {
        buysell = false
        legType = mt.MicrotickLegPut
        premium = quote.PutBid(matcher.Trade.Strike)
    }
        
    cost := mt.NewMicrotickCoinFromDec(premium.Amount.Mul(quote.Quantity.Amount))
    
    matcher.FillInfo = append(matcher.FillInfo, NewQuoteFillInfo(quote, buysell, legType,
        quote.Quantity.Amount, cost, quote.Backing, true))
}

func (matcher *Matcher) AssignCounterparties(ctx sdk.Context, keeper Keeper, market *DataMarket) error {
    for i, thisFill := range matcher.FillInfo {
        thisQuote := thisFill.Quote
        
        var long mt.MicrotickAccount
        var short mt.MicrotickAccount
        var premium mt.MicrotickPremium
        if thisFill.BuySell {
            long = matcher.Trade.Taker
            short = thisQuote.Provider
            premium = thisQuote.Ask
        } else {
            long = thisQuote.Provider
            short = matcher.Trade.Taker
            premium = thisQuote.Bid
        }
        
        quotedParams := NewDataQuotedParams(
            thisQuote.Id,
            premium,
            thisQuote.Quantity,
            thisQuote.Spot,
        )
        
        longAccountStatus := keeper.GetAccountStatus(ctx, long)
        shortAccountStatus := keeper.GetAccountStatus(ctx, short)
        var quoteProviderAccountStatus *DataAccountStatus
        if thisQuote.Provider.Equals(long) {
            quoteProviderAccountStatus = &longAccountStatus
        } else {
            quoteProviderAccountStatus = &shortAccountStatus
        }
        
        // Pay premium
        err := keeper.WithdrawMicrotickCoin(ctx, long, thisFill.Cost)
        if err != nil {
            return err
        }
        err = keeper.DepositMicrotickCoin(ctx, short, thisFill.Cost)
        if err != nil {
            return err
        }
        
        // Refund quote backing to quote provider
        err = keeper.DepositMicrotickCoin(ctx, thisQuote.Provider, thisFill.Backing)
        if err != nil {
            return err
        }
        
        // Back trade from short account
        err = keeper.WithdrawMicrotickCoin(ctx, short, thisFill.Backing)
        if err != nil {
            return err
        }
        
        // Adjust quote
        market.FactorOut(thisQuote)
        
        // Subtract out bought quantity and corresponding backing
        thisQuote.Quantity = mt.NewMicrotickQuantityFromDec(thisQuote.Quantity.Amount.Sub(thisFill.Quantity))
        thisQuote.Backing = thisQuote.Backing.Sub(thisFill.Backing)
        
        if thisQuote.Quantity.Amount.IsZero() {
            // If no quantity is left, delete quote from market, active quote list, and
            // account active quote list
            market.DeleteQuote(thisQuote)
            keeper.DeleteActiveQuote(ctx, thisQuote.Id)
            quoteProviderAccountStatus.ActiveQuotes.Delete(thisQuote.Id)
        } else {
            // else, factor quote back into market consensus
            market.FactorIn(thisQuote, false)
            keeper.SetActiveQuote(ctx, thisQuote)
        }
        
        // Update the account status of counterparties
        if !longAccountStatus.ActiveTrades.Contains(matcher.Trade.Id) {
            longAccountStatus.ActiveTrades.Insert(NewListItem(matcher.Trade.Id, sdk.NewDec(matcher.Trade.Expiration.UnixNano())))
        }
        if !shortAccountStatus.ActiveTrades.Contains(matcher.Trade.Id) {
            shortAccountStatus.ActiveTrades.Insert(NewListItem(matcher.Trade.Id, sdk.NewDec(matcher.Trade.Expiration.UnixNano())))
        }
        quoteProviderAccountStatus.QuoteBacking = quoteProviderAccountStatus.QuoteBacking.Sub(thisFill.Backing)
        shortAccountStatus.TradeBacking = shortAccountStatus.TradeBacking.Add(thisFill.Backing)
        
        // Save the counterparty account status in the store
        keeper.SetAccountStatus(ctx, long, longAccountStatus)
        keeper.SetAccountStatus(ctx, short, shortAccountStatus)
        
        // Append this counter party fill to the trade counterparty list
        matcher.Trade.Legs = append(matcher.Trade.Legs, NewDataTradeLeg(mt.MicrotickId(i), thisFill.LegType, thisFill.Backing, 
            thisFill.Cost, mt.NewMicrotickQuantityFromDec(thisFill.Quantity),
            thisFill.FinalFill, long, short, quotedParams,
        ))
    }
    return nil
}
