package keeper

import (
    "fmt"
    "time"
    sdk "github.com/cosmos/cosmos-sdk/types"
    
    mt "github.com/mjackson001/mtzone/x/microtick/types"
)

// DataActiveQuote

type DataActiveQuote struct {
    Id mt.MicrotickId `json:"id"`
    Market mt.MicrotickMarket `json:"market"`
    Duration mt.MicrotickDuration `json:"duration"`
    DurationName mt.MicrotickDurationName `json:"duration_name"`
    Provider mt.MicrotickAccount `json:"provider"`
    Modified time.Time `json:"modified"`
    CanModify time.Time `json:"canModify"`
    Backing mt.MicrotickCoin `json:"backing"`
    Ask mt.MicrotickPremium `json:"ask"`
    Bid mt.MicrotickPremium `json:"bid"`
    Quantity mt.MicrotickQuantity `json:"quantity"`
    Spot mt.MicrotickSpot `json:"spot"`
}

func NewDataActiveQuote(now time.Time, id mt.MicrotickId, market mt.MicrotickMarket, dur mt.MicrotickDuration, 
    durName mt.MicrotickDurationName, provider mt.MicrotickAccount, backing mt.MicrotickCoin, spot mt.MicrotickSpot, 
    ask mt.MicrotickPremium, bid mt.MicrotickPremium) DataActiveQuote {
        
    return DataActiveQuote {
        Id: id,
        Market: market,
        Duration: dur,
        DurationName: durName,
        Provider: provider,
        Backing: backing,
        Spot: spot,
        Ask: ask,
        Bid: bid,
        
        Modified: now,
        CanModify: now,
    }
}

func (daq *DataActiveQuote) ComputeQuantity() {
    averagePremium := daq.Ask.Amount.Add(daq.Bid.Amount).QuoInt64(2)
    actualLeverage := averagePremium.Mul(sdk.NewDec(mt.Leverage))
    daq.Quantity = mt.MicrotickQuantity{
        Denom: "quantity",
        Amount: daq.Backing.Amount.Quo(actualLeverage),
    }
}

func (daq *DataActiveQuote) Freeze(now time.Time, params mt.Params) {
    expire, err := time.ParseDuration(fmt.Sprintf("%d", params.FreezeTime) + "s")
    if err != nil {
        panic("invalid time")
    }
    daq.Modified = now
    daq.CanModify = now.Add(expire)
}

func (daq DataActiveQuote) Frozen(now time.Time) bool {
    if now.Before(daq.CanModify) {
        return true
    }
    return false
}

func (daq DataActiveQuote) Stale(now time.Time) bool {
    interval, err := time.ParseDuration(fmt.Sprintf("%d", daq.Duration * 2) + "s")
    if err != nil {
        panic("invalid time")
    }
    threshold := daq.Modified.Add(interval)
    if now.After(threshold) {
        return true
    }
    return false
}

func (daq DataActiveQuote) CallAsk(strike mt.MicrotickSpot) mt.MicrotickPremium {
    ask := daq.Ask.Amount
    delta := strike.Amount.Sub(daq.Spot.Amount)
    delta = delta.QuoInt64(2)
    if ask.LT(delta) {
        return mt.NewMicrotickPremiumFromInt(0)
    }
    return mt.NewMicrotickPremiumFromDec(ask.Sub(delta))
}

func (daq DataActiveQuote) PutAsk(strike mt.MicrotickSpot) mt.MicrotickPremium {
    ask := daq.Ask.Amount
    delta := daq.Spot.Amount.Sub(strike.Amount)
    delta = delta.QuoInt64(2)
    if ask.LT(delta) {
        return mt.NewMicrotickPremiumFromInt(0)
    }
    return mt.NewMicrotickPremiumFromDec(ask.Sub(delta))
}

func (daq DataActiveQuote) CallBid(strike mt.MicrotickSpot) mt.MicrotickPremium {
    bid := daq.Bid.Amount
    delta := strike.Amount.Sub(daq.Spot.Amount)
    delta = delta.QuoInt64(2)
    if bid.LT(delta) {
        return mt.NewMicrotickPremiumFromInt(0)
    }
    return mt.NewMicrotickPremiumFromDec(bid.Sub(delta))
}

func (daq DataActiveQuote) PutBid(strike mt.MicrotickSpot) mt.MicrotickPremium {
    bid := daq.Bid.Amount
    delta := daq.Spot.Amount.Sub(strike.Amount)
    delta = delta.QuoInt64(2)
    if bid.LT(delta) {
        return mt.NewMicrotickPremiumFromInt(0)
    }
    return mt.NewMicrotickPremiumFromDec(bid.Sub(delta))
}
