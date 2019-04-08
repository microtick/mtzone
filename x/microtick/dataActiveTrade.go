package microtick

import (
    "fmt"
    "time"
    sdk "github.com/cosmos/cosmos-sdk/types"
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
    Cost MicrotickCoin `json:"cost"`
    FilledQuantity MicrotickQuantity `json:"quantity"`
    Start time.Time `json:"start"`
    Expiration time.Time `json:"expiration"`
    Strike MicrotickSpot `json:"strike"`
}

func NewDataActiveTrade(market MicrotickMarket, dur MicrotickDuration,
    ttype MicrotickTradeType, long MicrotickAccount, strike MicrotickSpot) DataActiveTrade {
        
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
        Cost: NewMicrotickCoinFromInt(0),
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
    Cost MicrotickCoin `json:"premium"`
    FilledQuantity MicrotickQuantity `json:"quantity"`
    Short MicrotickAccount `json:"short"`
    Quoted DataQuoteParams `json:"quoted"`
}

func NewDataCounterParty(backing MicrotickCoin, final bool, cost MicrotickCoin, 
    quantity MicrotickQuantity)  DataCounterParty {
    return DataCounterParty {
        Backing: backing,
        Cost: cost,
        FilledQuantity: quantity,
    }
}

func (dat DataActiveTrade) CurrentValue(current MicrotickSpot) MicrotickCoin {
    strike := dat.Strike.Amount
    var delta sdk.Dec
    if dat.Type {
        // Put
        delta = strike.Sub(current.Amount)
    } else {
        // Call
        delta = current.Amount.Sub(strike)
    }
    if delta.IsNegative() {
        return NewMicrotickCoinFromInt(0)
    }
    return NewMicrotickCoinFromDec(delta.Mul(dat.FilledQuantity.Amount))
}
