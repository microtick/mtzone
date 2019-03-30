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
    Backing sdk.Coins `json:"backing"`
    Commission sdk.Coins `json:"commission"`
    Premium sdk.Coins `json:"premium"`
    Quantity MicrotickQuantity `json:"quantity"`
    Spot MicrotickSpot `json:"spot"`
}

func NewDataActiveQuote(id MicrotickId, market MicrotickMarket, dur MicrotickDuration, 
    provider MicrotickAccount, backing sdk.Coins, commission sdk.Coins, premium sdk.Coins, 
    quantity MicrotickQuantity, spot MicrotickSpot) DataActiveQuote {
        
    now := time.Now()
    return DataActiveQuote {
        Id: id,
        Market: market,
        Duration: dur,
        Provider: provider,
        Modified: now,
        CanModify: now,
        Backing: backing,
        Commission: commission,
        Premium: premium,
        Quantity: quantity,
        Spot: spot,
    }
}
