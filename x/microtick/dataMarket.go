package microtick

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
)

type DataOrderBook struct {
    Calls OrderedList `json:"calls"`
    Puts OrderedList `json:"puts"`
    SumBacking MicrotickCoin `json:"sumBacking"`
    SumWeight MicrotickQuantity `json:"sumWeight"`
}

type DataMarket struct {
    Market MicrotickMarket `json:"market"`
    Consensus MicrotickSpot `json:"consensus"`
    OrderBooks []DataOrderBook `json:"orderBooks"`
    SumBacking MicrotickCoin `json:"sumBacking"`
    SumSpots sdk.Dec `json:"sumSpots"`
    SumWeight MicrotickQuantity `json:"sumWeight"`
}

func NewDataMarket(market MicrotickMarket) DataMarket {
    return DataMarket {
        Market: market,
        Consensus: NewMicrotickSpotFromInt(0),
        OrderBooks: newOrderBooks(),
        SumBacking: NewMicrotickCoinFromInt(0),
        SumSpots: sdk.ZeroDec(),
        SumWeight: NewMicrotickQuantityFromInt(0),
    }
}

func newOrderBooks() []DataOrderBook {
    orderBooks := make([]DataOrderBook, len(MicrotickDurations))
    for i := range MicrotickDurations {
        orderBooks[i] = newOrderBook()
    }
    return orderBooks
}

func newOrderBook() DataOrderBook {
    return DataOrderBook {
        Calls: NewOrderedList(),
        Puts: NewOrderedList(),
        SumBacking: NewMicrotickCoinFromInt(0),
        SumWeight: NewMicrotickQuantityFromInt(0),
    }
}

func (dm *DataMarket) GetOrderBook(dur MicrotickDuration) DataOrderBook {
    for i := 0; i < len(MicrotickDurations); i++ {
        if MicrotickDurations[i] == dur {
            return dm.OrderBooks[i]
        }
    }
    panic("Invalid duration")
}

func (dm *DataMarket) SetOrderBook(dur MicrotickDuration, ob DataOrderBook) {
    for i := 0; i < len(MicrotickDurations); i++ {
        if MicrotickDurations[i] == dur {
            dm.OrderBooks[i] = ob
            return
        }
    }
    panic("Invalid duration")
}

func (dm *DataMarket) factorIn(quote DataActiveQuote) bool {
    dm.SumBacking = dm.SumBacking.Add(quote.Backing)
    dm.SumSpots = dm.SumSpots.Add(quote.Spot.Amount.Mul(
        quote.Quantity.Amount))
    dm.SumWeight = dm.SumWeight.Add(quote.Quantity)
    if dm.SumWeight.Amount.IsPositive() {
        dm.Consensus = NewMicrotickSpotFromDec(dm.SumSpots.Quo(dm.SumWeight.Amount))
    }
    
    // Test quote invariant:
    // Spot 2x limitation
    // A quote cannot be placed or updated that will be a free call or put on the 
    // resulting order book (spot more than 2x premium from resulting consensus)
    // Purpose: protects market maker from damaging quotes
    if quote.Spot.Amount.Sub(quote.Premium.Amount.MulInt64(2)).GT(dm.Consensus.Amount) {
        return false
    }
    if quote.Spot.Amount.Add(quote.Premium.Amount.MulInt64(2)).LT(dm.Consensus.Amount) {
        return false
    }
    
    orderBook := dm.GetOrderBook(quote.Duration)
    orderBook.SumBacking = orderBook.SumBacking.Add(quote.Backing)
    orderBook.SumWeight = orderBook.SumWeight.Add(quote.Quantity)
    
    // Test quote invariant:
    // Premium 2x limitation
    // A quote cannot be placed or updated with a premium of more than 2x the 
    // current market consensus premium (backing / (leverage * weight) for that time duration
    // Purpose: keeps premium realistic and tradeable within the quote's time frame
    if orderBook.SumWeight.Amount.IsPositive() {
        if orderBook.SumBacking.Amount.Quo(orderBook.SumWeight.Amount.MulInt64(Leverage)).MulInt64(2).LT(quote.Premium.Amount) {
            return false
        }
    }
    
    dm.SetOrderBook(quote.Duration, orderBook)
    return true
}

func (dm *DataMarket) factorOut(quote DataActiveQuote) {
    dm.SumBacking = dm.SumBacking.Sub(quote.Backing)
    dm.SumSpots = dm.SumSpots.Sub(quote.Spot.Amount.Mul(
        quote.Quantity.Amount))
    dm.SumWeight = dm.SumWeight.Sub(quote.Quantity)
    if dm.SumWeight.Amount.IsPositive() {
        dm.Consensus = NewMicrotickSpotFromDec(dm.SumSpots.Quo(dm.SumWeight.Amount))
    }
    
    orderBook := dm.GetOrderBook(quote.Duration)
    orderBook.SumBacking = orderBook.SumBacking.Sub(quote.Backing)
    orderBook.SumWeight = orderBook.SumWeight.Sub(quote.Quantity)
    dm.SetOrderBook(quote.Duration, orderBook)
}

func (dm *DataMarket) AddQuote(quote DataActiveQuote) {
    orderBook := dm.GetOrderBook(quote.Duration)
    callValue := quote.Premium.Amount.Add(quote.Spot.Amount.QuoInt64(2))
    orderBook.Calls.Insert(NewListItem(quote.Id, callValue))
    putValue := quote.Premium.Amount.Sub(quote.Spot.Amount.QuoInt64(2))
    orderBook.Puts.Insert(NewListItem(quote.Id, putValue))
    dm.SetOrderBook(quote.Duration, orderBook)
}

func (dm *DataMarket) DeleteQuote(quote DataActiveQuote) {
    orderBook := dm.GetOrderBook(quote.Duration)
    orderBook.Calls.Delete(quote.Id)
    orderBook.Puts.Delete(quote.Id)
    dm.SetOrderBook(quote.Duration, orderBook)
}

func (dm *DataMarket) MatchByQuantity(matcher *Matcher, quantity MicrotickQuantity) {
    orderBook := dm.GetOrderBook(MicrotickDurationFromName(matcher.Trade.Duration))
    quantityToMatch := quantity.Amount
    
    var list OrderedList
    if matcher.Trade.Type == MicrotickCall {
        list = orderBook.Calls
    }
    if matcher.Trade.Type == MicrotickPut {
        list = orderBook.Puts
    }
    
    index := 0
    for index < len(list.Data) && quantityToMatch.GT(sdk.ZeroDec()) {
        id := list.Data[index].Id 
        quote := matcher.FetchQuote(id)
        if !quote.Provider.Equals(matcher.Trade.Long) {
            var premium MicrotickPremium
            if matcher.Trade.Type == MicrotickCall {
                premium = quote.PremiumAsCall(matcher.Trade.Strike)
            }
            if matcher.Trade.Type == MicrotickPut {
                premium = quote.PremiumAsPut(matcher.Trade.Strike)
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
            matcher.TotalCost = matcher.TotalCost.Add(cost)
            
            matcher.FillInfo = append(matcher.FillInfo, QuoteFillInfo {
                Quote: quote,
                BoughtQuantity: boughtQuantity,
                Cost: NewMicrotickCoinFromDec(cost),
                FinalFill: finalFill,
            })
        //} else {
            //fmt.Printf("Skipping quote %d\n", id)
        }
        index++
    }
}

func (dm *DataMarket) MatchByLimit(matcher *Matcher, limit MicrotickPremium, maxCost MicrotickCoin) {
    orderBook := dm.GetOrderBook(MicrotickDurationFromName(matcher.Trade.Duration))
    
    var list OrderedList
    if matcher.Trade.Type == MicrotickCall {
        list = orderBook.Calls
    }
    if matcher.Trade.Type == MicrotickPut {
        list = orderBook.Puts
    }
    
    index := 0
    for index < len(list.Data) {
        id := list.Data[index].Id 
        quote := matcher.FetchQuote(id)
        if !quote.Provider.Equals(matcher.Trade.Long) {
            var premium MicrotickPremium
            if matcher.Trade.Type == MicrotickCall {
                premium = quote.PremiumAsCall(matcher.Trade.Strike)
            }
            if matcher.Trade.Type == MicrotickPut {
                premium = quote.PremiumAsPut(matcher.Trade.Strike)
            }
            
            if premium.Amount.LTE(limit.Amount) {
                var boughtQuantity sdk.Dec = quote.Quantity.Amount
                
                // Assume we're buying the entire quote's quantity
                cost := premium.Amount.Mul(boughtQuantity)
                
                // Check if cost is > max cost
                if cost.GT(maxCost.Amount) {
                    // Adjust quantity to fit max amount
                    boughtQuantity = maxCost.Amount.Quo(premium.Amount)
                    cost = premium.Amount.Mul(boughtQuantity)
                    maxCost.Amount = sdk.NewDec(0)
                }
                
                matcher.TotalQuantity = matcher.TotalQuantity.Add(boughtQuantity)
                matcher.TotalCost = matcher.TotalCost.Add(cost)
                maxCost.Amount = maxCost.Amount.Sub(cost)
                
                matcher.FillInfo = append(matcher.FillInfo, QuoteFillInfo {
                    Quote: quote,
                    BoughtQuantity: boughtQuantity,
                    Cost: NewMicrotickCoinFromDec(cost),
                    FinalFill: true,
                })
                
                if maxCost.IsZero() {
                    break
                }
                
            } else {
                
                // terminate - premium is > limit
                break
                
            }
        //} else {
            //fmt.Printf("Skipping quote %d\n", id)
        }
        index++
    }
}
