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
    Commission mt.MicrotickCoin `json:"commission"`
    Premium mt.MicrotickPremium `json:"premium"`
    Quantity mt.MicrotickQuantity `json:"quantity"`
    Spot mt.MicrotickSpot `json:"spot"`
}

func NewDataActiveQuote(now time.Time, id mt.MicrotickId, market mt.MicrotickMarket, dur mt.MicrotickDuration, 
    durName mt.MicrotickDurationName, provider mt.MicrotickAccount, backing mt.MicrotickCoin, spot mt.MicrotickSpot, 
    premium mt.MicrotickPremium) DataActiveQuote {
        
    return DataActiveQuote {
        Id: id,
        Market: market,
        Duration: dur,
        DurationName: durName,
        Provider: provider,
        Backing: backing,
        Spot: spot,
        Premium: premium,
        
        Modified: now,
        CanModify: now,
        Commission: mt.NewMicrotickCoinFromExtCoinInt(0),
    }
}

func (daq *DataActiveQuote) ComputeQuantity() {
    premiumLeverage := daq.Premium.Amount.Mul(sdk.NewDec(mt.Leverage))
    daq.Quantity = mt.MicrotickQuantity{
        Denom: "quantity",
        Amount: daq.Backing.Amount.Quo(premiumLeverage),
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

func (daq DataActiveQuote) PremiumAsCall(strike mt.MicrotickSpot) mt.MicrotickPremium {
    premium := daq.Premium.Amount
    delta := strike.Amount.Sub(daq.Spot.Amount)
    delta = delta.QuoInt64(2)
    if premium.LT(delta) {
        return mt.NewMicrotickPremiumFromInt(0)
    }
    return mt.NewMicrotickPremiumFromDec(premium.Sub(delta))
}

func (daq DataActiveQuote) PremiumAsPut(strike mt.MicrotickSpot) mt.MicrotickPremium {
    premium := daq.Premium.Amount
    delta := daq.Spot.Amount.Sub(strike.Amount)
    delta = delta.QuoInt64(2)
    if premium.LT(delta) {
        return mt.NewMicrotickPremiumFromInt(0)
    }
    return mt.NewMicrotickPremiumFromDec(premium.Sub(delta))
}
