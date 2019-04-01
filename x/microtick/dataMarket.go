package microtick

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
)

type DataOrderBook struct {
    Calls []MicrotickId `json:"calls"`
    Puts []MicrotickId `json:"puts"`
    SumWeight MicrotickQuantity `json:"sumWeight"`
}

type DataMarket struct {
    Market MicrotickMarket `json:"market"`
    Consensus MicrotickSpot `json:"consensus"`
    OrderBooks []DataOrderBook `json:"orderBooks"`
    SumBacking sdk.Coins `json:"sumBacking"`
    SumSpots MicrotickSpot `json:"sumSpots"`
    SumWeight MicrotickQuantity `json:"sumWeight"`
}

func NewDataMarket(market MicrotickMarket) DataMarket {
    return DataMarket {
        Market: market,
        Consensus: 0,
        OrderBooks: newOrderBooks(),
        SumBacking: sdk.Coins{sdk.NewInt64Coin(TokenType, 0)},
        SumSpots: 0,
        SumWeight: 0,
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
        Calls: make([]MicrotickId, 0),
        Puts: make([]MicrotickId, 0),
        SumWeight: 0,
    }
}
