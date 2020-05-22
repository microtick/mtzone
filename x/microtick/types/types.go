package types

import (
    "fmt"
    "regexp"
    "strconv"
    "strings"
    "github.com/pkg/errors"
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/x/auth"
)

const ExtTokenType = "udai"
const IntTokenType = "dai"
const ExtPerInt = 1000000

const Leverage = 10

const ModuleKey = "microtick"

// Account

type MicrotickAccount = sdk.AccAddress

// ID

type MicrotickId = uint32

func NewMicrotickIdFromString(s string) MicrotickId {
    id, err := strconv.Atoi(s)
    if err != nil {
        panic(fmt.Sprintf("Invalid ID: %s", s))
    }
    return MicrotickId(id)
}

// Market

type MicrotickMarket = string

// Duration

type MicrotickDuration = uint32
type MicrotickDurationName = string

var MicrotickDurations []MicrotickDuration
var MicrotickDurationNames []string 

func MicrotickDurationFromName(dur MicrotickDurationName) MicrotickDuration {
    for i, d := range MicrotickDurationNames {
        if dur == d {
            return MicrotickDurations[i]
        }
    }
    panic(fmt.Sprintf("Invalid duration: %s", dur))
}

func MicrotickDurationNameFromDur(dur MicrotickDuration) MicrotickDurationName {
    for i, d := range MicrotickDurations {
        if dur == d {
            return MicrotickDurationNames[i]
        }
    }
    panic(fmt.Sprintf("Invalid duration: %d", dur))
}

func ValidMicrotickDurationName(mtd MicrotickDurationName) bool {
    for i := 0; i < len(MicrotickDurationNames); i++ {
        if (mtd == MicrotickDurationNames[i]) {
            return true
        }
    }
    return false
}

// Type

type MicrotickTradeType = string
type MicrotickTradeTypeName = string

const (
    MicrotickCall = "call"
    MicrotickPut = "put"
)

func MicrotickTradeTypeFromName(str string) MicrotickTradeType {
    if str == "call" { return MicrotickCall }
    if str == "put" { return MicrotickPut }
    panic(fmt.Sprintf("Invalid trade type: %s", str))
}

func MicrotickTradeNameFromType(mtt MicrotickTradeType) MicrotickTradeTypeName {
    return mtt
}

// Backing

type MicrotickCoin = sdk.DecCoin
type ExtCoin = sdk.Coin

var reDecCoin = regexp.MustCompile(fmt.Sprintf(`^(%s)%s(%s)$`, `(?:[[:digit:]]*[.])?[[:digit:]]+`, `[[:space:]]*`, `[a-z][a-z0-9]{2,15}`))

func parseDecCoin(coinStr string) (coin sdk.DecCoin, err error) {
	coinStr = strings.TrimSpace(coinStr)

	matches := reDecCoin.FindStringSubmatch(coinStr)
	if matches == nil {
		return sdk.DecCoin{}, fmt.Errorf("Invalid decimal coin expression: %s", coinStr)
	}

	amountStr, denomStr := matches[1], matches[2]

	amount, err := sdk.NewDecFromStr(amountStr)
	if err != nil {
		return sdk.DecCoin{}, errors.Wrap(err, fmt.Sprintf("Failed to parse decimal coin amount: %s", amountStr))
	}

	return sdk.NewDecCoinFromDec(denomStr, amount), nil
}

// Input is in ExtToken units i.e. 1234000 -> 1.234 IntTokenType
func NewMicrotickCoinFromExtCoinInt(b int64) MicrotickCoin {
    result := sdk.NewInt64DecCoin(IntTokenType, b)
    result.Amount = result.Amount.QuoInt64(ExtPerInt)
    return result
}

// Input string can be IntTokenType or ExtTokenType
// "1.234IntTokenType" -> 1.234 IntTokenType
// "1234000ExtTokenType" -> 1.234 ExtTokenType
func NewMicrotickCoinFromString(str string) MicrotickCoin {
    result, err2 := parseDecCoin(str)
    if err2 != nil || (result.Denom != IntTokenType && result.Denom != ExtTokenType) {
        panic(fmt.Sprintf("Invalid coin amount or token type: %s", str))
    }
    if result.Denom == ExtTokenType {
        result.Amount = result.Amount.TruncateDec().QuoInt64(ExtPerInt)
        result.Denom = IntTokenType
    } else {
        result.Amount = result.Amount.MulInt64(ExtPerInt).TruncateDec().QuoInt64(ExtPerInt)
    }
    //fmt.Printf("Parsed: %s\n", result.String())
    return result
}

func NewMicrotickCoinFromDec(d sdk.Dec) MicrotickCoin {
    result := sdk.NewDecCoinFromDec(IntTokenType, d)
    result.Amount = result.Amount.MulInt64(ExtPerInt).TruncateDec().QuoInt64(ExtPerInt)
    return result
}

func MicrotickCoinToExtCoin(mc MicrotickCoin) ExtCoin {
    if mc.Denom != IntTokenType {
        panic(fmt.Sprintf("Not internal token type: %s", mc.Denom))
    }
    mc.Amount = mc.Amount.MulInt64(ExtPerInt)
    extCoin, _ := mc.TruncateDecimal()
    extCoin.Denom = ExtTokenType
    return extCoin
}

func ExtCoinToMicrotickCoin(ext sdk.Coin) MicrotickCoin {
    var amt = ext.Amount.Int64()
    var mc MicrotickCoin = NewMicrotickCoinFromExtCoinInt(amt)
    return mc
}

// Quantity

type MicrotickQuantity = sdk.DecCoin

func NewMicrotickQuantityFromInt(q int64) MicrotickQuantity {
    return sdk.NewInt64DecCoin("quantity", q)
}

func NewMicrotickQuantityFromString(q string) MicrotickQuantity {
    result, err2 := parseDecCoin(q)
    if err2 != nil || result.Denom != "quantity" {
        panic(fmt.Sprintf("Invalid quantity: %s", q))
    }
    return result
}

func NewMicrotickQuantityFromDec(d sdk.Dec) MicrotickQuantity {
    return sdk.NewDecCoinFromDec("quantity", d)
}


// Spot

type MicrotickSpot = sdk.DecCoin

func NewMicrotickSpotFromInt(s int64) MicrotickSpot {
    return sdk.NewInt64DecCoin("spot", s)
}


func NewMicrotickSpotFromString(s string) MicrotickSpot {
    result, err2 := parseDecCoin(s)
    if err2 != nil || result.Denom != "spot" {
        panic(fmt.Sprintf("Invalid spot: %s", s))
    }
    return result
}

func NewMicrotickSpotFromDec(d sdk.Dec) MicrotickSpot {
    return sdk.NewDecCoinFromDec("spot", d)
}

// Premium 

type MicrotickPremium = sdk.DecCoin

func NewMicrotickPremiumFromInt(p int64) MicrotickPremium {
    prem := sdk.NewInt64DecCoin("premium", p)
    //fmt.Println("NewMicrotickPremiumFromInt: %s\n", prem.String())
    return prem
}

func NewMicrotickPremiumFromString(p string) MicrotickPremium {
    result, err2 := parseDecCoin(p)
    if err2 != nil || result.Denom != "premium" {
        panic(fmt.Sprintf("Invalid premium: %s", p))
    }
    //fmt.Println("NewMicrotickPremiumFromString: %s\n", result.String())
    return result
}

func NewMicrotickPremiumFromDec(d sdk.Dec) MicrotickPremium {
    prem := sdk.NewDecCoinFromDec("premium", d)
    //fmt.Println("NewMicrotickPremiumFromDec: %s\n", prem.String())
    return prem
}

// Generic tx generate struct

type GenTx struct {
    Tx auth.StdTx `json:"tx"`
    AccountNumber uint64 `json:"accountNumber"`
    ChainID string `json:"chainId"`
    Sequence uint64 `json:"sequence"`
}
