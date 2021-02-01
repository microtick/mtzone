package keeper

import (
    "fmt"
    sdk "github.com/cosmos/cosmos-sdk/types"
   	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    
    mt "gitlab.com/microtick/mtzone/x/microtick/types"
)

type QuoteFillInfo struct {
    QuoteId mt.MicrotickId
    BuySell bool
    LegType mt.MicrotickLegType
    Quantity mt.MicrotickQuantity
    Premium mt.MicrotickPremium
    Cost mt.MicrotickCoin
    Backing mt.MicrotickCoin
    Refund mt.MicrotickCoin
}

func NewQuoteFillInfo(quoteId mt.MicrotickId, buySell bool, legType mt.MicrotickLegType, 
    quantity mt.MicrotickQuantity, premium mt.MicrotickPremium, cost mt.MicrotickCoin, 
    backing mt.MicrotickCoin, refund mt.MicrotickCoin) QuoteFillInfo {
    return QuoteFillInfo {
        QuoteId: quoteId,
        BuySell: buySell,
        LegType: legType,
        Quantity: quantity,
        Premium: premium,
        Cost: cost,
        Backing: backing,
        Refund: refund,
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

func (matcher *Matcher) DebugMatcherFillInfo() {
    for i, fi := range matcher.FillInfo {
        if i > 0 {
            fmt.Println()
        }
        var bs string
        if fi.BuySell {
            bs = "Buy"
        } else {
            bs = "Sell"
        }
        fmt.Printf("Fill %d\n", i)
        fmt.Printf("  Quote ID: %d\n", fi.QuoteId)
        fmt.Printf("  %s %s\n", bs, mt.MicrotickLegNameFromType(fi.LegType))
        fmt.Printf("  Quantity: %s\n", fi.Quantity.String())
        fmt.Printf("  Premium: %s\n", fi.Premium.String())
        fmt.Printf("  Cost: %s\n", fi.Cost.String())
        fmt.Printf("  Backing: %s\n", fi.Backing.String())
        fmt.Printf("  Refund: %s\n", fi.Refund.String())
    }
}

func (matcher *Matcher) MatchByQuantity(k Keeper, dm *DataMarket, order mt.MicrotickOrderType, totalQuantity mt.MicrotickQuantity) error {
    orderBook := dm.GetOrderBook(matcher.Trade.Duration)
    quantityToMatch := totalQuantity.Amount
    
    var list []mt.MicrotickId
    if order == mt.MicrotickOrderBuyCall {
        for i := 0; i < len(orderBook.CallAsks.Data); i++ {
            list = append(list, orderBook.CallAsks.Data[i].Id)
        }
    }
    if order == mt.MicrotickOrderBuyPut {
        for i := 0; i < len(orderBook.PutAsks.Data); i++ {
            list = append(list, orderBook.PutAsks.Data[i].Id)
        }
    }
    if order == mt.MicrotickOrderSellCall {
        for i := 0; i < len(orderBook.CallBids.Data); i++ {
            j := len(orderBook.CallBids.Data) - i - 1
            list = append(list, orderBook.CallBids.Data[j].Id)
        }
    }
    if order == mt.MicrotickOrderSellPut {
        for i := 0; i < len(orderBook.PutBids.Data); i++ {
            j := len(orderBook.PutBids.Data) - i - 1
            list = append(list, orderBook.CallBids.Data[j].Id)
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
            if order == mt.MicrotickOrderBuyCall {
                buysell = true
                legType = mt.MicrotickLegCall
                premium = quote.CallAsk(matcher.Trade.Strike)
            }
            if order == mt.MicrotickOrderBuyPut {
                buysell = true
                legType = mt.MicrotickLegPut
                premium = quote.PutAsk(matcher.Trade.Strike)
            }
            if order == mt.MicrotickOrderSellCall {
                buysell = false
                legType = mt.MicrotickLegCall
                premium = quote.CallBid(matcher.Trade.Strike)
            }
            if order == mt.MicrotickOrderSellPut {
                buysell = false
                legType = mt.MicrotickLegPut
                premium = quote.PutBid(matcher.Trade.Strike)
            }
            
            var quantity sdk.Dec
            var finalFill bool
            if quote.Quantity.Amount.GT(quantityToMatch) {
                quantity = quantityToMatch
                quantityToMatch = sdk.ZeroDec()
                finalFill = false
            } else {
                quantity = quote.Quantity.Amount
                quantityToMatch = quantityToMatch.Sub(quote.Quantity.Amount)
                finalFill = true
            }
            
            matcher.HasQuantity = true
            cost := premium.Amount.Mul(quantity)
            
            var backing mt.MicrotickCoin
            var refund mt.MicrotickCoin
            if finalFill {
                backing = quote.Backing
            } else {
                // For a partially filled quote, the backing transferred should be proportional to
                // the quantity purchased.
                backing = mt.NewMicrotickCoinFromDec(quote.Backing.Amount.Mul(quantity.Quo(quote.Quantity.Amount)))
            }
            refund = backing
            
            matcher.FillInfo = append(matcher.FillInfo, NewQuoteFillInfo(quote.Id, buysell, legType, 
                mt.NewMicrotickQuantityFromDec(quantity), premium, mt.NewMicrotickCoinFromDec(cost), 
                backing, refund))
        }
        index++
    }
    return nil
}

func (matcher *Matcher) MatchSynthetic(k Keeper, sob *DataSyntheticBook, dm *DataMarket, totalQuantity mt.MicrotickQuantity) error {
    if totalQuantity.Amount.GT(sob.SumWeight.Amount) {
        return sdkerrors.Wrap(mt.ErrTradeMatch, "insufficient quantity")
    }
    quantityToMatch := totalQuantity.Amount
    
    var list []DataSyntheticQuote
    var buysell bool
    if matcher.Trade.Order == mt.MicrotickOrderBuySyn {
        list = sob.Asks
        buysell = true
    }
    if matcher.Trade.Order == mt.MicrotickOrderSellSyn {
        list = sob.Bids
        buysell = false
    }
    
    index := 0
    for index < len(list) && quantityToMatch.GT(sdk.ZeroDec()) {
        li := list[index]
        
        // verify both quotes have enough quantity
        quoteAsk := matcher.FetchQuote(li.AskId)
        quoteBid := matcher.FetchQuote(li.BidId)
        
        var quantity sdk.Dec
        var synthFill bool
        if li.Quantity.Amount.GT(quantityToMatch) {
            quantity = quantityToMatch
            quantityToMatch = sdk.ZeroDec()
            synthFill = false
        } else {
            quantity = li.Quantity.Amount
            quantityToMatch = quantityToMatch.Sub(li.Quantity.Amount)
            synthFill = true
        }
        
        var legType mt.MicrotickLegType
        var premium mt.MicrotickPremium
        var backing mt.MicrotickCoin
        var refund mt.MicrotickCoin
        if buysell {
            premium = quoteAsk.CallAsk(matcher.Trade.Strike)
            legType = mt.MicrotickLegCall
        } else {
            premium = quoteAsk.PutAsk(matcher.Trade.Strike)
            legType = mt.MicrotickLegPut
        }
        cost := premium.Amount.Mul(quantity)
        backing = mt.NewMicrotickCoinFromDec(quoteAsk.Backing.Amount.Mul(quantity.Quo(quoteAsk.Quantity.Amount)))
        if synthFill && li.AskFill && !li.BidFill {
            refund = quoteAsk.Backing
        } else {
            refund = backing
        }
        
        // Append ask
        matcher.FillInfo = append(matcher.FillInfo, NewQuoteFillInfo(quoteAsk.Id, true, legType, 
            mt.NewMicrotickQuantityFromDec(quantity),
            premium, mt.NewMicrotickCoinFromDec(cost), backing, refund))
            
        if buysell {
            premium = quoteBid.PutBid(matcher.Trade.Strike)
            legType = mt.MicrotickLegPut
        } else {
            premium = quoteBid.CallBid(matcher.Trade.Strike)
            legType = mt.MicrotickLegCall
        }
        
        cost = premium.Amount.Mul(quantity)
        if synthFill && li.BidFill {
            backing = quoteBid.Backing
        } else {
            backing = mt.NewMicrotickCoinFromDec(quoteBid.Backing.Amount.Mul(quantity.Quo(quoteBid.Quantity.Amount)))
        }
        refund = backing
        
        // Append bid
        matcher.FillInfo = append(matcher.FillInfo, NewQuoteFillInfo(quoteBid.Id, false, legType, 
            mt.NewMicrotickQuantityFromDec(quantity),
            premium, mt.NewMicrotickCoinFromDec(cost), backing, refund))
        
        // At least one time through - we have liquidity
        matcher.HasQuantity = true
        index++
    }
    //matcher.DebugMatcherFillInfo()
    return nil
}

func (matcher *Matcher) MatchQuote(k Keeper, order mt.MicrotickOrderType, quote DataActiveQuote) {
    var premium mt.MicrotickPremium
    var buysell bool
    var legType mt.MicrotickLegType
    if order == mt.MicrotickOrderBuyCall {
        buysell = true
        legType = mt.MicrotickLegCall
        premium = quote.CallAsk(matcher.Trade.Strike)
    }
    if order == mt.MicrotickOrderBuyPut {
        buysell = true
        legType = mt.MicrotickLegPut
        premium = quote.PutAsk(matcher.Trade.Strike)
    }
    if order == mt.MicrotickOrderSellCall {
        buysell = false
        legType = mt.MicrotickLegCall
        premium = quote.CallBid(matcher.Trade.Strike)
    }
    if order == mt.MicrotickOrderSellPut {
        buysell = false
        legType = mt.MicrotickLegPut
        premium = quote.PutBid(matcher.Trade.Strike)
    }
        
    cost := premium.Amount.Mul(quote.Quantity.Amount)
    
    backing := quote.Backing
    refund := backing
    
    matcher.HasQuantity = true
    matcher.FillInfo = append(matcher.FillInfo, NewQuoteFillInfo(quote.Id, buysell, legType,
        quote.Quantity, premium, mt.NewMicrotickCoinFromDec(cost), backing, refund))
}

func (matcher *Matcher) AssignCounterparties(ctx sdk.Context, keeper Keeper, market *DataMarket) error {
    for i, thisFill := range matcher.FillInfo {
        thisQuote, err := keeper.GetActiveQuote(ctx, thisFill.QuoteId)
        if err != nil {
            return err
        }
        
        var long mt.MicrotickAccount
        var short mt.MicrotickAccount
        if thisFill.BuySell {
            long = matcher.Trade.Taker
            short = thisQuote.Provider
        } else {
            long = thisQuote.Provider
            short = matcher.Trade.Taker
        }
        
        
        longAccountStatus := keeper.GetAccountStatus(ctx, long)
        shortAccountStatus := keeper.GetAccountStatus(ctx, short)
        var quoteProviderAccountStatus *DataAccountStatus
        if thisQuote.Provider.Equals(long) {
            quoteProviderAccountStatus = &longAccountStatus
        } else {
            quoteProviderAccountStatus = &shortAccountStatus
        }
        
        // Pay premium
        if thisFill.Cost.Amount.GT(sdk.ZeroDec()) {
            err = keeper.WithdrawMicrotickCoin(ctx, long, thisFill.Cost)
            if err != nil {
                return err
            }
            err = keeper.DepositMicrotickCoin(ctx, short, thisFill.Cost)
            if err != nil {
                return err
            }
        }
        
        // Refund quote backing to quote provider
        err = keeper.DepositMicrotickCoin(ctx, thisQuote.Provider, thisFill.Refund)
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
        thisQuote.Quantity = mt.NewMicrotickQuantityFromDec(thisQuote.Quantity.Amount.Sub(thisFill.Quantity.Amount))
        thisQuote.Backing = thisQuote.Backing.Sub(thisFill.Refund)
        
        var finalFill bool
        if thisQuote.Backing.Amount.IsZero() {
            // If no quantity is left, delete quote from market, active quote list, and
            // account active quote list
            finalFill = true
            market.DeleteQuote(thisQuote)
            keeper.DeleteActiveQuote(ctx, thisQuote.Id)
            quoteProviderAccountStatus.ActiveQuotes.Delete(thisQuote.Id)
        } else {
            // else, factor quote back into market consensus
            finalFill = false
            market.FactorIn(thisQuote, false)
            keeper.SetActiveQuote(ctx, thisQuote)
        }
        
        // Update the account status of counterparties
        if !longAccountStatus.ActiveTrades.Contains(matcher.Trade.Id) {
            longAccountStatus.ActiveTrades.Insert(NewListItem(matcher.Trade.Id, sdk.NewDec(matcher.Trade.Expiration)))
        }
        if !shortAccountStatus.ActiveTrades.Contains(matcher.Trade.Id) {
            shortAccountStatus.ActiveTrades.Insert(NewListItem(matcher.Trade.Id, sdk.NewDec(matcher.Trade.Expiration)))
        }
        quoteProviderAccountStatus.QuoteBacking = quoteProviderAccountStatus.QuoteBacking.Sub(thisFill.Refund)
        shortAccountStatus.TradeBacking = shortAccountStatus.TradeBacking.Add(thisFill.Backing)
        
        // Save the counterparty account status in the store
        keeper.SetAccountStatus(ctx, long, longAccountStatus)
        keeper.SetAccountStatus(ctx, short, shortAccountStatus)
        
        quotedPremium := thisQuote.Ask
        if !thisFill.BuySell {
            quotedPremium = thisQuote.Bid
        }
        quotedParams := NewDataQuotedParams(
            thisQuote.Id,
            finalFill,
            quotedPremium,
            thisQuote.ComputeUnitBacking(),
            thisQuote.Spot,
        )
        
        // Append this counter party fill to the trade counterparty list
        matcher.Trade.Quantity.Amount = matcher.Trade.Quantity.Amount.Add(thisFill.Quantity.Amount)
        matcher.Trade.Legs = append(matcher.Trade.Legs, NewDataTradeLeg(mt.MicrotickId(i), thisFill.LegType, thisFill.Backing, 
            thisFill.Premium, thisFill.Cost, thisFill.Quantity, long, short, quotedParams,
        ))
    }
    return nil
}
