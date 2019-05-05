package microtick

import (
    "fmt"
    "regexp"
    "strconv"
    "strings"
    "github.com/pkg/errors"
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/x/auth"
)

const TokenType = "fox"
const Leverage = 10

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

type MicrotickDuration = uint16
type MicrotickDurationName = string

var MicrotickDurations = []MicrotickDuration {
    300, // 5 minutes
    900, // 15 minutes
    3600, // 1 hour
    14400, // 4 hours
    43200, // 12 hours
}

var MicrotickDurationNames = []string {
    "5minute",
    "15minute",
    "1hour",
    "4hour",
    "12hour",
}

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
    return ""
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
type MicrotickTradeTypeName = string

const (
    MicrotickCall = false  // 0
    MicrotickPut = true    // 1
)

func MicrotickTradeTypeFromName(str string) MicrotickTradeType {
    if str == "call" { return MicrotickCall }
    if str == "put" { return MicrotickPut }
    panic(fmt.Sprintf("Invalid trade type: %s", str))
}

func MicrotickTradeNameFromType(mtt MicrotickTradeType) MicrotickTradeTypeName {
    if mtt {
        return "put"
    }
    return "call"
}

// Backing

type MicrotickCoin = sdk.DecCoin

var reDecCoin = regexp.MustCompile(fmt.Sprintf(`^(%s)%s(%s)$`, `(?:[[:digit:]]*[.])?[[:digit:]]+`, `[[:space:]]*`, `[a-z][a-z0-9]{2,15}`))

func parseDecCoin(coinStr string) (coin sdk.DecCoin, err error) {
	coinStr = strings.TrimSpace(coinStr)

	matches := reDecCoin.FindStringSubmatch(coinStr)
	if matches == nil {
		return sdk.DecCoin{}, fmt.Errorf("invalid decimal coin expression: %s", coinStr)
	}

	amountStr, denomStr := matches[1], matches[2]

	amount, err := sdk.NewDecFromStr(amountStr)
	if err != nil {
		return sdk.DecCoin{}, errors.Wrap(err, fmt.Sprintf("failed to parse decimal coin amount: %s", amountStr))
	}

	return sdk.NewDecCoinFromDec(denomStr, amount), nil
}

func NewMicrotickCoinFromInt(b int64) MicrotickCoin {
    return sdk.NewInt64DecCoin(TokenType, b)
}

func NewMicrotickCoinFromString(b string) MicrotickCoin {
    result, err2 := parseDecCoin(b)
    if err2 != nil || result.Denom != TokenType {
        panic(fmt.Sprintf("Invalid coin amount or token type: %s", b))
    }
    return result
}

func NewMicrotickCoinFromDec(d sdk.Dec) MicrotickCoin {
    return sdk.NewDecCoinFromDec(TokenType, d)
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
    return sdk.NewInt64DecCoin("premium", p)
}

func NewMicrotickPremiumFromString(p string) MicrotickPremium {
    result, err2 := parseDecCoin(p)
    if err2 != nil || result.Denom != "premium" {
        panic(fmt.Sprintf("Invalid premium: %s", p))
    }
    return result
}

func NewMicrotickPremiumFromDec(d sdk.Dec) MicrotickPremium {
    return sdk.NewDecCoinFromDec("premium", d)
}

// Generic tx generate struct

type GenTx struct {
    Tx auth.StdTx `json:"tx"`
    AccountNumber uint64 `json:"accountNumber"`
    ChainID string `json:"chainId"`
    Sequence uint64 `json:"sequence"`
}
