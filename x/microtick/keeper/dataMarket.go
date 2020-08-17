package keeper

import (
    "time"
    sdk "github.com/cosmos/cosmos-sdk/types"
    
    mt "github.com/mjackson001/mtzone/x/microtick/types"
)

type DataOrderBook struct {
    Name string `json:"name"`
    CallAsks OrderedList `json:"callasks"`
    CallBids OrderedList `json:"callbids"`
    PutAsks OrderedList `json:"putasks"`
    PutBids OrderedList `json:"putbids"`
    SumBacking mt.MicrotickCoin `json:"sumBacking"`
    SumWeight mt.MicrotickQuantity `json:"sumWeight"`
}

type DataMarket struct {
    Market mt.MicrotickMarket `json:"market"`
    Description string `json:"description"`
    Consensus mt.MicrotickSpot `json:"consensus"`
    OrderBooks []DataOrderBook `json:"orderBooks"`
    SumBacking mt.MicrotickCoin `json:"sumBacking"`
    SumSpots sdk.Dec `json:"sumSpots"`
    SumWeight mt.MicrotickQuantity `json:"sumWeight"`
    // Internal: quote list ordered by time of maturity
    Quotes OrderedList `json:"quotes"`
}

func NewDataMarket(market mt.MicrotickMarket, description string, durs []string) DataMarket {
    orderBooks := make([]DataOrderBook, len(durs))
    for i := 0; i < len(durs); i++ {
        orderBooks[i] = newOrderBook(durs[i])
    }
    return DataMarket {
        Market: market,
        Description: description,
        Consensus: mt.NewMicrotickSpotFromInt(0),
        OrderBooks: orderBooks,
        SumBacking: mt.NewMicrotickCoinFromExtCoinInt(0),
        SumSpots: sdk.ZeroDec(),
        SumWeight: mt.NewMicrotickQuantityFromInt(0),
    }
}

func newOrderBook(name string) DataOrderBook {
    return DataOrderBook {
        Name: name,
        CallAsks: NewOrderedList(),
        PutAsks: NewOrderedList(),
        SumBacking: mt.NewMicrotickCoinFromExtCoinInt(0),
        SumWeight: mt.NewMicrotickQuantityFromInt(0),
    }
}

func (dm *DataMarket) GetOrderBook(name string) DataOrderBook {
    for i := 0; i < len(dm.OrderBooks); i++ {
        if dm.OrderBooks[i].Name == name {
            return dm.OrderBooks[i]
        }
    }
    panic("Invalid duration name")
}

func (dm *DataMarket) SetOrderBook(name string, ob DataOrderBook) {
    for i := 0; i < len(dm.OrderBooks); i++ {
        if dm.OrderBooks[i].Name == name {
            dm.OrderBooks[i] = ob
            return
        }
    }
    panic("Invalid duration name")
}

func (dm *DataMarket) FactorIn(quote DataActiveQuote, testInvariants bool) bool {
    // FactorIn is called from Tx's that originate from the quote provider, i.e.
    // create quote, update quote and from Tx's that result from a counter party, 
    // i.e. trade market, trade limit.
    // We only want to test invariants when we're not trade matching, because
    // the invariants apply only to creation or updates, not from the results
    // of market action over time.  (quotes can go stale, etc)

    dm.SumBacking = dm.SumBacking.Add(quote.Backing)
    dm.SumSpots = dm.SumSpots.Add(quote.Spot.Amount.Mul(
        quote.Quantity.Amount))
    dm.SumWeight = dm.SumWeight.Add(quote.Quantity)
    if dm.SumWeight.Amount.IsPositive() {
        dm.Consensus = mt.NewMicrotickSpotFromDec(dm.SumSpots.Quo(dm.SumWeight.Amount))
    }
    
    // Test quote invariant:
    if testInvariants {
        // Spot 2x limitation
        // A quote cannot be placed or updated that will be a free call or put on the 
        // resulting order book (spot more than 2x premium from resulting consensus)
        // Purpose: protects market maker from damaging quotes
        if quote.Spot.Amount.Sub(quote.Ask.Amount.MulInt64(2)).GT(dm.Consensus.Amount) {
            //fmt.Printf("Failed Spot Invariant (1): %d\n", quote.Id)
            return false
        }
        if quote.Spot.Amount.Add(quote.Ask.Amount.MulInt64(2)).LT(dm.Consensus.Amount) {
            //fmt.Printf("Failed Spot Invariant (2): %d\n", quote.Id)
            return false
        }
        // Premium must be less than 1/2 spot, otherwise if consensus moves less than spot,
        // it would be possible for the premium to reach negative price territory
        if quote.Spot.Amount.QuoInt64(2).LT(quote.Ask.Amount) {
            return false
        }
    }
    
    orderBook := dm.GetOrderBook(quote.DurationName)
    orderBook.SumBacking = orderBook.SumBacking.Add(quote.Backing)
    orderBook.SumWeight = orderBook.SumWeight.Add(quote.Quantity)
    
    dm.SetOrderBook(quote.DurationName, orderBook)
    return true
}

func (dm *DataMarket) FactorOut(quote DataActiveQuote) {
    dm.SumBacking = dm.SumBacking.Sub(quote.Backing)
    dm.SumSpots = dm.SumSpots.Sub(quote.Spot.Amount.Mul(
        quote.Quantity.Amount))
    dm.SumWeight = dm.SumWeight.Sub(quote.Quantity)
    if dm.SumWeight.Amount.IsPositive() {
        dm.Consensus = mt.NewMicrotickSpotFromDec(dm.SumSpots.Quo(dm.SumWeight.Amount))
    }
    
    orderBook := dm.GetOrderBook(quote.DurationName)
    orderBook.SumBacking = orderBook.SumBacking.Sub(quote.Backing)
    orderBook.SumWeight = orderBook.SumWeight.Sub(quote.Quantity)
    dm.SetOrderBook(quote.DurationName, orderBook)
}

func (dm *DataMarket) AddQuote(quote DataActiveQuote) {
    orderBook := dm.GetOrderBook(quote.DurationName)
    
    spotDiv2 := quote.Spot.Amount.QuoInt64(2)
    orderBook.CallAsks.Insert(NewListItem(quote.Id, quote.Ask.Amount.Add(spotDiv2)))
    orderBook.CallBids.Insert(NewListItem(quote.Id, quote.Bid.Amount.Add(spotDiv2)))
    orderBook.PutAsks.Insert(NewListItem(quote.Id, quote.Ask.Amount.Sub(spotDiv2)))
    orderBook.PutBids.Insert(NewListItem(quote.Id, quote.Bid.Amount.Sub(spotDiv2)))
    
    dm.Quotes.Insert(NewListItem(quote.Id, sdk.NewDec(quote.CanModify.Unix())))
    
    dm.SetOrderBook(quote.DurationName, orderBook)
}

func (dm *DataMarket) DeleteQuote(quote DataActiveQuote) {
    orderBook := dm.GetOrderBook(quote.DurationName)
    
    orderBook.CallAsks.Delete(quote.Id)
    orderBook.CallBids.Delete(quote.Id)
    orderBook.PutAsks.Delete(quote.Id)
    orderBook.PutBids.Delete(quote.Id)
    
    dm.Quotes.Delete(quote.Id)
    
    dm.SetOrderBook(quote.DurationName, orderBook)
}

func (dm * DataMarket) CanSettle(now time.Time) bool {
    if len(dm.Quotes.Data) == 0 {
        return false
    }
    if dm.Quotes.Data[0].Value.GT(sdk.NewDec(now.Unix())) {
        return false
    }
    return true
}

func (dm *DataMarket) MatchByQuantity(matcher *Matcher, quantity mt.MicrotickQuantity) {
    orderBook := dm.GetOrderBook(matcher.Trade.DurationName)
    quantityToMatch := quantity.Amount
    
    var list OrderedList
    if matcher.Trade.Type == mt.MicrotickCall {
        list = orderBook.CallAsks
    }
    if matcher.Trade.Type == mt.MicrotickPut {
        list = orderBook.PutAsks
    }
    
    index := 0
    for index < len(list.Data) && quantityToMatch.GT(sdk.ZeroDec()) {
        id := list.Data[index].Id 
        quote := matcher.FetchQuote(id)
        if !quote.Provider.Equals(matcher.Trade.Long) {
            var premium mt.MicrotickPremium
            if matcher.Trade.Type == mt.MicrotickCall {
                premium = quote.CallAsk(matcher.Trade.Strike)
            }
            if matcher.Trade.Type == mt.MicrotickPut {
                premium = quote.PutAsk(matcher.Trade.Strike)
            }
            
            var boughtQuantity sdk.Dec
            finalFill := false
            
            if quote.Quantity.Amount.GTE(quantityToMatch) {
                boughtQuantity = quantityToMatch
                quantityToMatch = sdk.ZeroDec()
            } else {
                boughtQuantity = quote.Quantity.Amount
                quantityToMatch = quantityToMatch.Sub(quote.Quantity.Amount)
                finalFill = true
            }
            
            matcher.TotalQuantity = matcher.TotalQuantity.Add(boughtQuantity)
            cost := premium.Amount.Mul(boughtQuantity)
            matcher.TotalCost = matcher.TotalCost.Add(mt.NewMicrotickCoinFromDec(cost))
            
            matcher.FillInfo = append(matcher.FillInfo, QuoteFillInfo {
                Quote: quote,
                BoughtQuantity: boughtQuantity,
                Cost: mt.NewMicrotickCoinFromDec(cost),
                FinalFill: finalFill,
            })
        }
        index++
    }
}

func (dm *DataMarket) MatchByLimit(matcher *Matcher, limit mt.MicrotickPremium, maxCost mt.MicrotickCoin) {
    orderBook := dm.GetOrderBook(matcher.Trade.DurationName)
    
    var list OrderedList
    if matcher.Trade.Type == mt.MicrotickCall {
        list = orderBook.CallAsks
    }
    if matcher.Trade.Type == mt.MicrotickPut {
        list = orderBook.PutAsks
    }
    
    index := 0
    for index < len(list.Data) {
        id := list.Data[index].Id 
        quote := matcher.FetchQuote(id)
        if !quote.Provider.Equals(matcher.Trade.Long) {
            var premium mt.MicrotickPremium
            if matcher.Trade.Type == mt.MicrotickCall {
                premium = quote.CallAsk(matcher.Trade.Strike)
            }
            if matcher.Trade.Type == mt.MicrotickPut {
                premium = quote.PutAsk(matcher.Trade.Strike)
            }
            
            if premium.Amount.LTE(limit.Amount) {
                var boughtQuantity sdk.Dec = quote.Quantity.Amount
                
                // Assume we're buying the entire quote's quantity
                cost := premium.Amount.Mul(boughtQuantity)
                
                // Check if cost is > max cost
                finalFill := true
                if cost.GT(maxCost.Amount) {
                    // Adjust quantity to fit max amount
                    boughtQuantity = maxCost.Amount.Quo(premium.Amount)
                    cost = premium.Amount.Mul(boughtQuantity)
                    maxCost.Amount = sdk.NewDec(0)
                    finalFill = false
                }
                
                matcher.TotalQuantity = matcher.TotalQuantity.Add(boughtQuantity)
                matcher.TotalCost = matcher.TotalCost.Add(mt.NewMicrotickCoinFromDec(cost))
                maxCost.Amount = maxCost.Amount.Sub(cost)
                
                matcher.FillInfo = append(matcher.FillInfo, QuoteFillInfo {
                    Quote: quote,
                    BoughtQuantity: boughtQuantity,
                    Cost: mt.NewMicrotickCoinFromDec(cost),
                    FinalFill: finalFill,
                })
                
                if maxCost.IsZero() {
                    break
                }
                
            } else {
                
                // terminate - premium is > limit
                break
                
            }
        }
        index++
    }
}
