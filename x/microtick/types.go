package microtick

import (
    "fmt"
    "strconv"
    sdk "github.com/cosmos/cosmos-sdk/types"
)

const TokenType = "fox"
const Leverage = 10

// Account

type MicrotickAccount = string

// ID

type MicrotickId = uint32

// Market

type MicrotickMarket = string

// Duration

type MicrotickDuration = uint16

var MicrotickDurations = []MicrotickDuration {
    300, // 5 minutes
    900, // 15 minutes
    3600, // 1 hour
    14400, // 4 hours
    43200, // 12 hours
}

func NewMicrotickDurationFromString(d string) (mtd MicrotickDuration, err sdk.Error) {
    var dur MicrotickDuration
    dur2, err2 := strconv.ParseInt(d, 10, 16)
    if err2 != nil {
        return 0, sdk.ErrInternal("Not an integer value: " + d)
    }
    dur = MicrotickDuration(dur2)
    for _, d := range MicrotickDurations {
        if dur == d {
            return dur, nil
        }
    }
    return 0, sdk.ErrInternal("Invalid duration: " + fmt.Sprintf("%d", dur))
}

func ValidMicrotickDuration(mtd MicrotickDuration) bool {
    for i := 0; i < len(MicrotickDurations); i++ {
        if (mtd == MicrotickDurations[i]) {
            return true
        }
    }
    return false
}

// Type

type MicrotickTradeType = bool

const (
    MicrotickCall = false  // 0
    MicrotickPut = true    // 1
)

// Backing

type MicrotickCoin = sdk.DecCoin

func NewMicrotickCoinFromInt(b int64) MicrotickCoin {
    return sdk.NewInt64DecCoin(TokenType, b)
}

func NewMicrotickCoinFromString(b string) (mtq MicrotickQuantity, err sdk.Error) {
    amount, err := sdk.NewDecFromStr(b)
    var result MicrotickCoin
    if err != nil {
        return result, err
    }
    result = sdk.NewDecCoinFromDec(TokenType, amount)
    fmt.Printf("Coin: %s\n", result.String())
    return result, nil
}

// Quantity

type MicrotickQuantity = sdk.DecCoin

func NewMicrotickQuantityFromInt(q int64) MicrotickQuantity {
    return sdk.NewInt64DecCoin("quantity", q)
}

func NewMicrotickQuantityFromString(q string) (mtq MicrotickQuantity, err sdk.Error) {
    amount, err := sdk.NewDecFromStr(q)
    var result MicrotickQuantity
    if err != nil {
        return result, err
    }
    result = sdk.NewDecCoinFromDec("quantity", amount)
    fmt.Printf("Quantity: %s\n", result.String())
    return result, nil
}

// Spot

type MicrotickSpot = sdk.DecCoin

func NewMicrotickSpotFromInt(s int64) MicrotickQuantity {
    return sdk.NewInt64DecCoin("spot", s)
}


func NewMicrotickSpotFromString(s string) (mts MicrotickSpot, err sdk.Error) {
    amount, err := sdk.NewDecFromStr(s)
    var result MicrotickSpot
    if err != nil {
        return result, err
    }
    result = sdk.NewDecCoinFromDec("spot", amount)
    fmt.Printf("Spot: %s\n", result.String())
    return result, nil
}

// Premium 

type MicrotickPremium = sdk.DecCoin

func NewMicrotickPremiumFromInt(p int64) MicrotickQuantity {
    return sdk.NewInt64DecCoin("premium", p)
}

func NewMicrotickPremiumFromString(p string) (mts MicrotickPremium, err sdk.Error) {
    amount, err := sdk.NewDecFromStr(p)
    var result MicrotickPremium
    if err != nil {
        return result, err
    }
    result = sdk.NewDecCoinFromDec("premium", amount)
    fmt.Printf("Premium: %s\n", result.String())
    return result, nil
}
