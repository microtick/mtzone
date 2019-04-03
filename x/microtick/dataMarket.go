package microtick

type DataOrderBook struct {
    Calls OrderedList `json:"calls"`
    Puts OrderedList `json:"puts"`
    SumWeight MicrotickQuantity `json:"sumWeight"`
}

type DataMarket struct {
    Market MicrotickMarket `json:"market"`
    Consensus MicrotickSpot `json:"consensus"`
    OrderBooks []DataOrderBook `json:"orderBooks"`
    SumBacking MicrotickCoin `json:"sumBacking"`
    SumSpots MicrotickSpot `json:"sumSpots"`
    SumWeight MicrotickQuantity `json:"sumWeight"`
}

func NewDataMarket(market MicrotickMarket) DataMarket {
    return DataMarket {
        Market: market,
        Consensus: NewMicrotickSpotFromInt(0),
        OrderBooks: newOrderBooks(),
        SumBacking: NewMicrotickCoinFromInt(0),
        SumSpots: NewMicrotickSpotFromInt(0),
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
        SumWeight: NewMicrotickQuantityFromInt(0),
    }
}

//func (dm *DataMarket) factorIn(Data)
