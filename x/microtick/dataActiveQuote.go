package microtick

import (
    "fmt"
    "time"
    sdk "github.com/cosmos/cosmos-sdk/types"
)

// DataActiveQuote

type DataActiveQuote struct {
    Id MicrotickId `json:"id"`
    Market MicrotickMarket `json:"market"`
    Duration MicrotickDuration `json:"duration"`
    Provider MicrotickAccount `json:"provider"`
    Modified time.Time `json:"modified"`
    CanModify time.Time `json:"canModify"`
    Backing MicrotickCoin `json:"backing"`
    Commission MicrotickCoin `json:"commission"`
    Premium MicrotickPremium `json:"premium"`
    Quantity MicrotickQuantity `json:"quantity"`
    Spot MicrotickSpot `json:"spot"`
}

func NewDataActiveQuote(id MicrotickId, market MicrotickMarket, dur MicrotickDuration, 
    provider MicrotickAccount, backing MicrotickCoin, spot MicrotickSpot, 
    premium MicrotickPremium) DataActiveQuote {
        
    now := time.Now()
    return DataActiveQuote {
        Id: id,
        Market: market,
        Duration: dur,
        Provider: provider,
        Backing: backing,
        Spot: spot,
        Premium: premium,
        
        Modified: now,
        CanModify: now,
        Commission: NewMicrotickCoinFromInt(0),
    }
}

func (daq *DataActiveQuote) ComputeQuantity() {
    premiumLeverage := daq.Premium.Amount.Mul(sdk.NewDec(Leverage))
    daq.Quantity = MicrotickQuantity{
        Denom: "quantity",
        Amount: daq.Backing.Amount.Quo(premiumLeverage),
    }
}

func (daq *DataActiveQuote) Freeze(params Params) {
    now := time.Now()    
    expire, err := time.ParseDuration(fmt.Sprintf("%d", params.FreezeTime) + "s")
    if err != nil {
        panic("invalid time")
    }
    daq.Modified = now
    daq.CanModify = now.Add(expire)
}

func (daq DataActiveQuote) Frozen() bool {
    now := time.Now()
    if now.Before(daq.CanModify) {
        return true
    }
    return false
}

func (daq DataActiveQuote) PremiumAsCall(strike MicrotickSpot) MicrotickPremium {
    premium := daq.Premium.Amount
    delta := strike.Amount.Sub(daq.Spot.Amount)
    delta = delta.QuoInt64(2)
    if premium.LT(delta) {
        return NewMicrotickPremiumFromInt(0)
    }
    return NewMicrotickPremiumFromDec(premium.Sub(delta))
}

func (daq DataActiveQuote) PremiumAsPut(strike MicrotickSpot) MicrotickPremium {
    premium := daq.Premium.Amount
    delta := daq.Spot.Amount.Sub(strike.Amount)
    delta = delta.QuoInt64(2)
    if premium.LT(delta) {
        return NewMicrotickPremiumFromInt(0)
    }
    return NewMicrotickPremiumFromDec(premium.Sub(delta))
}
