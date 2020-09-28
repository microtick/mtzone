package types

import (
    "fmt"
    "regexp"
    "strconv"
    "strings"
    "github.com/pkg/errors"
    sdk "github.com/cosmos/cosmos-sdk/types"
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

// Types

type MicrotickOrderType = string
type MicrotickOrderTypeName = string

const (
    MicrotickOrderBuyCall = "buy-call"
    MicrotickOrderSellCall = "sell-call"
    MicrotickOrderBuyPut = "buy-put"
    MicrotickOrderSellPut = "sell-put"
    MicrotickOrderBuySyn = "buy-syn"
    MicrotickOrderSellSyn = "sell-syn"
)

func MicrotickOrderTypeFromName(str string) MicrotickOrderType {
    if str == "buy-call" { return MicrotickOrderBuyCall }
    if str == "sell-call" { return MicrotickOrderSellCall }
    if str == "buy-put" { return MicrotickOrderBuyPut }
    if str == "sell-put" { return MicrotickOrderSellPut }
    if str == "buy-syn" { return MicrotickOrderBuySyn }
    if str == "sell-syn" { return MicrotickOrderSellSyn }
    panic(fmt.Sprintf("Invalid order type: %s", str))
}

func MicrotickOrderNameFromType(mot MicrotickOrderType) MicrotickOrderTypeName {
    return mot
}

type MicrotickLegType = bool
type MicrotickLegTypeName = string

const (
    MicrotickLegCall = true
    MicrotickLegPut = false
)

func MicrotickLegTypeFromName(str string) MicrotickLegType {
    if str == "call" { return MicrotickLegCall }
    if str == "put" { return MicrotickLegPut }
    panic(fmt.Sprintf("Invalid trade leg type: %s", str))
}

func MicrotickLegNameFromType(mlt MicrotickLegType) MicrotickLegTypeName {
    if mlt {
        return "call"
    } else {
        return "put"
    }
}

// Backing

type MicrotickCoin = sdk.DecCoin
type ExtCoin = sdk.Coin

var reDecCoin = regexp.MustCompile(fmt.Sprintf(`^(%s)%s(%s)$`, `(?:[[:digit:]]*[.])?[[:digit:]]+`, `[[:space:]]*`, `[a-z][a-z0-9]{2,15}`))

func parseDecCoin(coinStr string, allowNegativeValue bool) (coin sdk.DecCoin, err error) {
	coinStr = strings.TrimSpace(coinStr)
	
	isNegative := false
	if allowNegativeValue && strings.HasPrefix(coinStr, "-") {
	    isNegative = true
	    coinStr = coinStr[1:]
	}

	matches := reDecCoin.FindStringSubmatch(coinStr)
	if matches == nil {
		return sdk.DecCoin{}, fmt.Errorf("Invalid decimal coin expression: %s", coinStr)
	}

	amountStr, denomStr := matches[1], matches[2]

	amount, err := sdk.NewDecFromStr(amountStr)
	if err != nil {
		return sdk.DecCoin{}, errors.Wrap(err, fmt.Sprintf("Failed to parse decimal coin amount: %s", amountStr))
	}
	
	if isNegative {
	    amount = amount.Neg()
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
    result, err := parseDecCoin(str, false)
    if err != nil || (result.Denom != IntTokenType && result.Denom != ExtTokenType) {
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
    result, err := parseDecCoin(q, false)
    if err != nil || result.Denom != "quantity" {
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
    result, err := parseDecCoin(s, false)
    if err != nil || result.Denom != "spot" {
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
    result, err := parseDecCoin(p, false)
    if err != nil || result.Denom != "premium" {
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
