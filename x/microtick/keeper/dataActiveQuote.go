package keeper

import (
    "fmt"
    "time"
    sdk "github.com/cosmos/cosmos-sdk/types"
    
    mt "github.com/mjackson001/mtzone/x/microtick/types"
)

// DataActiveQuote

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
        
        Modified: now.Unix(),
        CanModify: now.Unix(),
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

func (daq *DataActiveQuote) Freeze(now time.Time, params mt.MicrotickParams) {
    expire, err := time.ParseDuration(fmt.Sprintf("%d", params.FreezeTime) + "s")
    if err != nil {
        panic("invalid time")
    }
    daq.Modified = now.Unix()
    daq.CanModify = now.Add(expire).Unix()
}

func (daq DataActiveQuote) Frozen(now time.Time) bool {
    if now.Before(time.Unix(daq.CanModify, 0)) {
        return true
    }
    return false
}

func (daq DataActiveQuote) Stale(now time.Time) bool {
    interval, err := time.ParseDuration(fmt.Sprintf("%d", daq.Duration * 2) + "s")
    if err != nil {
        panic("invalid time")
    }
    threshold := time.Unix(daq.Modified, 0).Add(interval)
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
