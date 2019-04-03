package microtick

import (
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
