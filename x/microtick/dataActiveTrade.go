package microtick

import (
    "fmt"
    "time"
)

type DataActiveTrade struct {
    Id MicrotickId `json:"id"`
    Market MicrotickMarket `json:"market"`
    Duration MicrotickDuration `json:"duration"`
    Type MicrotickTradeType `json:"type"`
    Commission MicrotickCoin `json:"commission"`
    CounterParties []DataCounterParty `json:"counterParties"`
    Long MicrotickAccount `json:"long"`
    Backing MicrotickCoin `json:"backing"`
    Premium MicrotickCoin `json:"premium"`  // for trades, premium is in Coin not Premium type
    RequestedQuantity MicrotickQuantity `json:"requestedQuantity"`
    FilledQuantity MicrotickQuantity `json:"quantity"`
    Start time.Time `json:"start"`
    Expiration time.Time `json:"expiration"`
    Strike MicrotickSpot `json:"strike"`
}

func NewDataActiveTrade(market MicrotickMarket, dur MicrotickDuration,
    ttype MicrotickTradeType, long MicrotickAccount, strike MicrotickSpot,
    quantity MicrotickQuantity) DataActiveTrade {
        
    now := time.Now()    
    expire, err := time.ParseDuration(fmt.Sprintf("%d", dur) + "s")
    if err != nil {
        panic("invalid time")
    }
    return DataActiveTrade {
        Id: 0, // set actual trade ID later after premium has been verified
        Market: market,
        Duration: dur,
        Type: ttype,
        Commission: NewMicrotickCoinFromInt(0), // commission computed later
        Long: long,
        Backing: NewMicrotickCoinFromInt(0),
        Premium: NewMicrotickCoinFromInt(0),
        RequestedQuantity: quantity,
        FilledQuantity: NewMicrotickQuantityFromInt(0), // computed later
        Start: now,
        Expiration: now.Add(expire),
        Strike: strike,
    }
}

func (trade DataActiveTrade) AddCounterParty(cp DataCounterParty) {
    trade.CounterParties = append(trade.CounterParties, cp)
}

type DataQuoteParams struct {
    Id MicrotickId `json:"quoteId"`
    Premium MicrotickPremium `json:"premium"`
    Quantity MicrotickQuantity `json:"quantity"`
    Spot MicrotickSpot `json:"spot"`
}

func NewDataQuoteParams(id MicrotickId, premium MicrotickPremium, quantity MicrotickQuantity,
    spot MicrotickSpot) DataQuoteParams {
    return DataQuoteParams {
        Id: id,
        Premium: premium,
        Quantity: quantity,
        Spot: spot,
    }
}

type DataCounterParty struct {
    Backing MicrotickCoin `json:"backing"`
    Premium MicrotickPremium `json:"premium"`
    FilledQuantity MicrotickQuantity `json:"quantity"`
    Short MicrotickAccount `json:"short"`
    Quoted DataQuoteParams `json:"quoted"`
}

func NewDataCounterParty(backing MicrotickCoin, final bool, premium MicrotickCoin, 
    quantity MicrotickQuantity)  DataCounterParty {
    return DataCounterParty {
        Backing: backing,
        Premium: premium,
        FilledQuantity: quantity,
    }
}
