package microtick

import (
    "time"
    sdk "github.com/cosmos/cosmos-sdk/types"
)

type DataActiveTrade struct {
    Id MicrotickId `json:"id"`
    Market MicrotickMarket `json:"market"`
    Duration MicrotickDuration `json:"duration"`
    Type MicrotickTradeType `json:"type"`
    Commission sdk.Coins `json:"commission"`
    CounterParties []DataCounterParty `json:"counterParties"`
    Long MicrotickAccount `json:"long"`
    Premium sdk.Coins `json:"premium"`
    Quantity MicrotickQuantity `json:"quantity"`
    Start time.Time `json:"start"`
    Expiration time.Time `json:"expiration"`
    Strike MicrotickSpot `json:"strike"`
}

func NewDataActiveTrade(id MicrotickId, market MicrotickMarket, dur MicrotickDuration,
    ttype MicrotickTradeType, commission sdk.Coins, long MicrotickAccount, premium sdk.Coins, 
    quantity MicrotickQuantity, strike MicrotickSpot) DataActiveTrade {
        
    now := time.Now()    
    return DataActiveTrade {
        Id: id,
        Market: market,
        Duration: dur,
        Type: ttype,
        Commission: commission,
        CounterParties: make([]DataCounterParty, 0),
        Long: long,
        Premium: premium,
        Quantity: quantity,
        Start: now,
        Expiration: now.Add(time.Duration(dur)),
        Strike: strike,
    }
}

func (trade DataActiveTrade) AddCounterParty(cp DataCounterParty) {
    trade.CounterParties = append(trade.CounterParties, cp)
}

type DataQuoteParams struct {
    Id MicrotickId `json:"quoteId"`
    Premium sdk.Coins `json:"premium"`
    Quantity MicrotickQuantity `json:"quantity"`
    Spot MicrotickSpot `json:"spot"`
}

func NewDataQuoteParams(id MicrotickId, premium sdk.Coins, quantity MicrotickQuantity,
    spot MicrotickSpot) DataQuoteParams {
    return DataQuoteParams {
        Id: id,
        Premium: premium,
        Quantity: quantity,
        Spot: spot,
    }
}

type DataCounterParty struct {
    Backing sdk.Coins `json:"backing"`
    Final bool `json:"final"`
    Premium sdk.Coins `json:"premium"`
    Quoted DataQuoteParams `json:"quoted"`
    Quantity MicrotickQuantity `json:"quantity"`
    Short MicrotickAccount `json:"short"`
}

func NewDataCounterParty(backing sdk.Coins, final bool, premium sdk.Coins, 
    quoted DataQuoteParams, quantity MicrotickQuantity, 
    short MicrotickAccount)  DataCounterParty {
    return DataCounterParty {
        Backing: backing,
        Final: final,
        Premium: premium,
        Quoted: quoted,
        Quantity: quantity,
        Short: short,
    }
}
