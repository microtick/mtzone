package microtick

import (
    "fmt"
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

func (dm *DataMarket) factorIn(quote DataActiveQuote) {
    dm.SumBacking = dm.SumBacking.Plus(quote.Backing)
    dm.SumSpots = dm.SumSpots.Add(quote.Spot.Amount.Mul(
        quote.Quantity.Amount))
    dm.SumWeight = dm.SumWeight.Plus(quote.Quantity)
    if (dm.SumWeight.Amount.IsPositive()) {
        dm.Consensus = NewMicrotickSpotFromDec(dm.SumSpots.Quo(dm.SumWeight.Amount))
    }
    
    orderBook := dm.GetOrderBook(quote.Duration)
    orderBook.SumBacking = orderBook.SumBacking.Plus(quote.Backing)
    orderBook.SumWeight = orderBook.SumWeight.Plus(quote.Quantity)
    dm.SetOrderBook(quote.Duration, orderBook)
}

func (dm *DataMarket) factorOut(quote DataActiveQuote) {
    dm.SumBacking = dm.SumBacking.Minus(quote.Backing)
    dm.SumSpots = dm.SumSpots.Sub(quote.Spot.Amount.Mul(
        quote.Quantity.Amount))
    dm.SumWeight = dm.SumWeight.Minus(quote.Quantity)
    if (dm.SumWeight.Amount.IsPositive()) {
        dm.Consensus = NewMicrotickSpotFromDec(dm.SumSpots.Quo(dm.SumWeight.Amount))
    }
    
    orderBook := dm.GetOrderBook(quote.Duration)
    orderBook.SumBacking = orderBook.SumBacking.Minus(quote.Backing)
    orderBook.SumWeight = orderBook.SumWeight.Minus(quote.Quantity)
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
    fmt.Printf("Order Book: %+v\n", orderBook)
    dm.SetOrderBook(quote.Duration, orderBook)
}

type FetchQuoteFunc func(MicrotickId) DataActiveQuote
type AddCounterPartyFunc func(DataActiveQuote, sdk.Dec, MicrotickCoin)

func (dm *DataMarket) Match(trade DataActiveTrade, fetchQuote FetchQuoteFunc, 
    addCounterParty AddCounterPartyFunc) (MicrotickQuantity, MicrotickPremium) {
    orderBook := dm.GetOrderBook(trade.Duration)
    totalQuantity := sdk.ZeroDec()
    totalPremium := sdk.ZeroDec()
    quantityToMatch := trade.RequestedQuantity.Amount
    
    var list OrderedList
    if trade.Type == MicrotickCall {
        list = orderBook.Calls
    }
    if trade.Type == MicrotickPut {
        list = orderBook.Puts
    }
    
    index := 0
    for index < len(list.Data) && quantityToMatch.GT(sdk.ZeroDec()) {
        id := list.Data[index].Id 
        quote := fetchQuote(id)
        if !quote.Provider.Equals(trade.Long) {
            fmt.Printf("Matching quote %d\n", id)
            var premium MicrotickPremium
            if trade.Type == MicrotickCall {
                premium = quote.PremiumAsCall(trade.Strike)
            }
            if trade.Type == MicrotickPut {
                premium = quote.PremiumAsPut(trade.Strike)
            }
            fmt.Printf("  quote quantity: %s\n", quote.Quantity.Amount.String())
            fmt.Printf("  quantity to match: %s\n", quantityToMatch.String())
            fmt.Printf("  premium: %s\n", premium.String())
            
            var boughtQuantity sdk.Dec
            
            if quote.Quantity.Amount.GTE(quantityToMatch) {
                boughtQuantity = quantityToMatch
                quantityToMatch = sdk.ZeroDec()
            } else {
                boughtQuantity = quote.Quantity.Amount
                quantityToMatch = quantityToMatch.Sub(quote.Quantity.Amount)
            }
            
            totalQuantity = totalQuantity.Add(boughtQuantity)
            paidPremium := premium.Amount.Mul(boughtQuantity)
            totalPremium = totalPremium.Add(paidPremium)
            
            if addCounterParty != nil {
                addCounterParty(quote, boughtQuantity, NewMicrotickCoinFromDec(paidPremium))
            }
        } else {
            fmt.Printf("Skipping quote %d\n", id)
        }
        index++
    }
    return NewMicrotickQuantityFromDec(totalQuantity), NewMicrotickPremiumFromDec(totalPremium)
}
