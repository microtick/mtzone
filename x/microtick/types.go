package microtick

import (
    "fmt"
    "regexp"
    "strings"
    "github.com/pkg/errors"
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

var MicrotickDurationNames = []string {
    "5minute",
    "15minute",
    "1hour",
    "4hour",
    "12hour",
}

func NewMicrotickDurationFromString(dur string) (mtd MicrotickDuration, err sdk.Error) {
    for i, d := range MicrotickDurationNames {
        if dur == d {
            return MicrotickDurations[i], nil
        }
    }
    return 0, sdk.ErrInternal("Invalid duration: " + fmt.Sprintf("%d", dur))
}

func MicrotickDurationNameFromDur(dur MicrotickDuration) string {
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

const (
    MicrotickCall = false  // 0
    MicrotickPut = true    // 1
)

// Backing

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

type MicrotickCoin = sdk.DecCoin

func NewMicrotickCoinFromInt(b int64) MicrotickCoin {
    return sdk.NewInt64DecCoin(TokenType, b)
}

func NewMicrotickCoinFromString(b string) (mtq MicrotickQuantity, err sdk.Error) {
    result, err2 := parseDecCoin(b)
    if err2 != nil || result.Denom != TokenType {
        return result, sdk.ErrInternal("Invalid coin suffix")
    }
    return result, nil
}

// Quantity

type MicrotickQuantity = sdk.DecCoin

func NewMicrotickQuantityFromInt(q int64) MicrotickQuantity {
    return sdk.NewInt64DecCoin("quantity", q)
}

func NewMicrotickQuantityFromString(q string) (mtq MicrotickQuantity, err sdk.Error) {
    result, err2 := parseDecCoin(q)
    if err2 != nil || result.Denom != "quantity" {
        return result, sdk.ErrInternal("Invalid quantity")
    }
    return result, nil
}

// Spot

type MicrotickSpot = sdk.DecCoin

func NewMicrotickSpotFromInt(s int64) MicrotickQuantity {
    return sdk.NewInt64DecCoin("spot", s)
}


func NewMicrotickSpotFromString(s string) (mts MicrotickSpot, err sdk.Error) {
    result, err2 := parseDecCoin(s)
    if err2 != nil || result.Denom != "spot" {
        return result, sdk.ErrInternal("Invalid spot")
    }
    return result, nil
}

// Premium 

type MicrotickPremium = sdk.DecCoin

func NewMicrotickPremiumFromInt(p int64) MicrotickQuantity {
    return sdk.NewInt64DecCoin("premium", p)
}

func NewMicrotickPremiumFromString(p string) (mts MicrotickPremium, err sdk.Error) {
    result, err2 := parseDecCoin(p)
    if err2 != nil || result.Denom != "premium" {
        return result, sdk.ErrInternal("Invalid premium")
    }
    return result, nil
}
