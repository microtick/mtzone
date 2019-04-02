package microtick

import (
    "math"
    "strconv"
    "strings"
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
    Premium MicrotickPremium `json:"premium"`
    Quantity MicrotickQuantity `json:"quantity"`
    Spot MicrotickSpot `json:"spot"`
}

func NewDataActiveQuote(id MicrotickId, market MicrotickMarket, dur MicrotickDuration, 
    provider MicrotickAccount, backing sdk.Coins, spot MicrotickSpot, 
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
        Commission: sdk.Coins{sdk.NewInt64Coin(TokenType, 0)},
    }
}

func (daq *DataActiveQuote) ComputeQuantity() {
    pow := math.Pow(10, MicrotickQuantityDecimals)
    backStr := daq.Backing.String()
    backing, err2 := strconv.ParseFloat(strings.Replace(backStr, "fox", "", 1), 10)
    premium := float64(daq.Premium) / pow
    if err2 == nil {
        q := math.Round(pow * backing / (premium * Leverage))
        //fmt.Printf("Backing: %f\n", backing)
        //fmt.Printf("Premium: %f\n", premium)
        //fmt.Printf("Quantity: %d\n", MicrotickQuantity(q))
        daq.Quantity = MicrotickQuantity(q)
    }
}
