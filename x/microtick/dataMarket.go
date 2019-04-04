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

func (dm *DataMarket) factorIn(quote DataActiveQuote) {
    dm.SumBacking = dm.SumBacking.Plus(quote.Backing)
    dm.SumSpots = dm.SumSpots.Add(quote.Spot.Amount.Mul(
        quote.Quantity.Amount))
    dm.SumWeight = dm.SumWeight.Plus(quote.Quantity)
    if (dm.SumWeight.Amount.IsPositive()) {
        dm.Consensus = MicrotickSpot{
            Denom: "spot",
            Amount: dm.SumSpots.Quo(dm.SumWeight.Amount),
        }
    }
    
    var orderBookIndex, i int
    for i = 0; i < len(MicrotickDurations); i++ {
        if MicrotickDurations[i] == quote.Duration {
            orderBookIndex = i
        }
    }
    
    orderBook := dm.OrderBooks[orderBookIndex]
    
    orderBook.SumBacking = orderBook.SumBacking.Plus(quote.Backing)
    orderBook.SumWeight = orderBook.SumWeight.Plus(quote.Quantity)
}

func (dm *DataMarket) factorOut(quote DataActiveQuote) {
    dm.SumBacking = dm.SumBacking.Minus(quote.Backing)
    dm.SumSpots = dm.SumSpots.Sub(quote.Spot.Amount.Mul(
        quote.Quantity.Amount))
    dm.SumWeight = dm.SumWeight.Minus(quote.Quantity)
    if (dm.SumWeight.Amount.IsPositive()) {
        dm.Consensus = MicrotickSpot{
            Denom: "spot",
            Amount: dm.SumSpots.Quo(dm.SumWeight.Amount),
        }
    }
    
    var orderBookIndex, i int
    for i = 0; i < len(MicrotickDurations); i++ {
        if MicrotickDurations[i] == quote.Duration {
            orderBookIndex = i
        }
    }
    
    orderBook := dm.OrderBooks[orderBookIndex]
    
    orderBook.SumBacking = orderBook.SumBacking.Minus(quote.Backing)
    orderBook.SumWeight = orderBook.SumWeight.Minus(quote.Quantity)
}
