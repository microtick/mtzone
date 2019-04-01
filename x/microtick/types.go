package microtick

import (
    "fmt"
    "strconv"
    "math"
    sdk "github.com/cosmos/cosmos-sdk/types"
)

const TokenType = "fox"

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

// Type

type MicrotickTradeType = bool

const (
    MicrotickCall = false  // 0
    MicrotickPut = true    // 1
)

// Quantity

type MicrotickQuantity = uint32
const MicrotickQuantityDecimals = 6

func NewMicrotickQuantityFromString(q string) (mtq MicrotickQuantity, err sdk.Error) {
    tmp, err2 := strconv.ParseFloat(q, 64)
    if err2 != nil {
        return 0, sdk.ErrInternal("Not a floating point value: " + q)
    }
    pow := math.Pow(10, MicrotickQuantityDecimals)
    var intPart uint32 = uint32(tmp)
    var fracPart uint16 = uint16((tmp - float64(intPart)) * pow)
    fmt.Printf("Quantity Int part: %d\n", intPart)
    fmt.Printf("Quantity Frac part: %d\n", fracPart)
    result := MicrotickQuantity(intPart * uint32(pow) + uint32(fracPart))
    fmt.Printf("Quantity Result %d\n", result)
    return result, nil
}

// Spot

type MicrotickSpot = uint32
const MicrotickSpotDecimals = 4

func NewMicrotickSpotFromString(q string) (mts MicrotickSpot, err sdk.Error) {
    tmp, err2 := strconv.ParseFloat(q, 64)
    if err2 != nil {
        return 0, sdk.ErrInternal("Not a floating point value: " + q)
    }
    pow := math.Pow(10, MicrotickSpotDecimals)
    var intPart uint32 = uint32(tmp)
    var fracPart uint16 = uint16((tmp - float64(intPart)) * pow)
    fmt.Printf("Spot Int part: %d\n", intPart)
    fmt.Printf("Spot Frac part: %d\n", fracPart)
    result := MicrotickQuantity(intPart * uint32(pow) + uint32(fracPart))
    fmt.Printf("Spot Result %d\n", result)
    return result, nil
}

// Premium 

type MicrotickPremium = uint32
const MicrotickPremiumDecimals = 6

func NewMicrotickPremiumFromString(q string) (mts MicrotickPremium, err sdk.Error) {
    tmp, err2 := strconv.ParseFloat(q, 64)
    if err2 != nil {
        return 0, sdk.ErrInternal("Not a floating point value: " + q)
    }
    pow := math.Pow(10, MicrotickPremiumDecimals)
    var intPart uint32 = uint32(tmp)
    var fracPart uint16 = uint16((tmp - float64(intPart)) * pow)
    fmt.Printf("Premium Int part: %d\n", intPart)
    fmt.Printf("Premium Frac part: %d\n", fracPart)
    result := MicrotickQuantity(intPart * uint32(pow) + uint32(fracPart))
    fmt.Printf("Premium Result %d\n", result)
    return result, nil
}
