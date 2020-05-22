package types

import (
	"fmt"
	"errors"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/x/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// DefaultParamspace defines the default microtick module parameter subspace
const TimeFormat = "2006-01-02T15:04:05Z"
const HaltTimeString = "168h"

type MarketParam struct {
	Name MicrotickMarket `json:"name"`
	Description string `json:"description"`
}

type DurationParam struct  {
	Name MicrotickDurationName `json:"name"`
	Seconds MicrotickDuration `json:"seconds"`
}

// Params defines the parameters for the microtick module.
type Params struct {
	  Markets	[]MarketParam `json:"markets"`
	  Durations []DurationParam `json:"durations"`
    EuropeanOptions bool `json:"european_options"`
    CommissionQuotePercent sdk.Dec `json:"commission_quote_percent"`
    CommissionTradeFixed sdk.Dec `json:"commission_trade_fixed"`
    CommissionUpdatePercent sdk.Dec `json:"commission_update_percent"`
    CommissionSettleFixed sdk.Dec `json:"commission_settle_fixed"`
    SettleIncentive sdk.Dec `json:"settle_incentive"`
    FreezeTime int8 `json:"freeze_time"`
    HaltTime string `json:"halt_time"`
}

// Default parameter values
var (
    DefaultEuropeanOptions bool = true
    DefaultCommissionQuotePercent = sdk.MustNewDecFromStr("0.0005")
    DefaultCommissionTradeFixed = sdk.MustNewDecFromStr("0.025")
    DefaultCommissionUpdatePercent = sdk.MustNewDecFromStr("0.00005")
    DefaultSettleIncentive = sdk.MustNewDecFromStr("0.025")
    DefaultCommissionSettleFixed = sdk.MustNewDecFromStr("0.01")
    DefaultFreezeTime = int8(30)
)

func DefaultMarkets() []MarketParam {
	var markets []MarketParam
	markets = append(markets, MarketParam{
		Name: "CHANGEME",
		Description: "Default market",
	})
	return markets
}

func DefaultDurations() []DurationParam {
	var durs []DurationParam
	durs = append(durs, DurationParam{
		Name: "5minute",
		Seconds: 300,
	}, DurationParam{
		Name: "10minute",
		Seconds: 600,
	}, DurationParam{
		Name: "15minute",
		Seconds: 900,
	}, DurationParam{
		Name: "30minute",
		Seconds: 1800,
	}, DurationParam{
		Name: "1hour",
		Seconds: 3600,
	}, DurationParam{
		Name: "2hour",
		Seconds: 7200,
	}, DurationParam{
		Name: "4hour",
		Seconds: 14400,
	}, DurationParam{
		Name: "8hour",
		Seconds: 28800,
	}, DurationParam{
		Name: "12hour",
		Seconds: 43200,
	}, DurationParam{
		Name: "1day",
		Seconds: 86400,
	})
	return durs
}

// Parameter keys
var (
	KeyMarkets = []byte("Markets")
	KeyDurations = []byte("Durations")
    KeyEuropeanOptions = []byte("EuropeanOptions")
    KeyCommissionQuotePercent = []byte("CommissionQuotePercent")
    KeyCommissionTradeFixed = []byte("KeyCommissionTradeFixed")
    KeyCommissionUpdatePercent = []byte("KeyCommissionUpdatePercent")
    KeyCommissionSettleFixed = []byte("KeyCommissionSettleFixed")
    KeySettleIncentive = []byte("KeySettleIncentive")
    KeyFreezeTime = []byte("KeyFreezeTime")
    KeyHaltTime = []byte("KeyHaltTime")
)

// ParamKeyTable for microtick module
func ParamKeyTable() params.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of microtick module's parameters.
// nolint
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		{KeyMarkets, &p.Markets, validateMarkets},
		{KeyDurations, &p.Durations, validateDurations},
	    {KeyEuropeanOptions, &p.EuropeanOptions, validateEuropeanOptions},
	    {KeyCommissionQuotePercent, &p.CommissionQuotePercent, validatePercent},
	    {KeyCommissionTradeFixed, &p.CommissionTradeFixed, validateFixed},
	    {KeyCommissionUpdatePercent, &p.CommissionUpdatePercent, validatePercent},
	    {KeyCommissionSettleFixed, &p.CommissionSettleFixed, validateFixed},
	    {KeySettleIncentive, &p.SettleIncentive, validateFixed},
	    {KeyFreezeTime, &p.FreezeTime, validateFreezeTime},
	    {KeyHaltTime, &p.HaltTime, validateTime},
	}
}

func validateMarkets(i interface{}) error {
	markets, ok := i.([]MarketParam)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	for i:=0; i<len(markets); i++ {
		if markets[i].Name == "" {
			return errors.New("market name must not be blank")
		}
	}
	return nil
}

func validateDurations(i interface{}) error {
	durs, ok := i.([]DurationParam)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	for i:=0; i<len(durs); i++ {
		if durs[i].Name == "" {
			return errors.New("duration name must not be blank")
		}
		if durs[i].Seconds == 0 {
			return errors.New("duration seconds must be positive")
		}
	}
	return nil
}

func validateEuropeanOptions(i interface{}) error {
	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validatePercent(i interface{}) error {
	_, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateFixed(i interface{}) error {
	_, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateFreezeTime(i interface{}) error {
	v, ok := i.(int8)
	if !ok {
		return fmt.Errorf("invalid freeze time: %s", v)
	}
	return nil
}

func validateTime(i interface{}) error {
	_, ok := time.Parse(TimeFormat, i.(string))
	if ok != nil {
		return fmt.Errorf("invalid time: %T", i)
	}
	return nil
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
    //interval, _ := time.ParseDuration(HaltTimeString)
    //defaultHaltTime := time.Now().UTC().Add(interval)
    defaultHaltTime, _ := time.Parse("2006-Jan-02", "2030-Jan-01")
	return Params{
		Markets: DefaultMarkets(),
		Durations: DefaultDurations(),
	    EuropeanOptions: DefaultEuropeanOptions,
	    CommissionQuotePercent: DefaultCommissionQuotePercent,
	    CommissionTradeFixed: DefaultCommissionTradeFixed,
	    CommissionUpdatePercent: DefaultCommissionUpdatePercent,
	    CommissionSettleFixed: DefaultCommissionSettleFixed,
	    SettleIncentive: DefaultSettleIncentive,
	    FreezeTime: DefaultFreezeTime,
	    HaltTime: defaultHaltTime.Format(TimeFormat),
	}
}

// String implements the stringer interface.
func (p Params) String() string {
	var sb strings.Builder
	sb.WriteString("Params: \n")
	sb.WriteString(fmt.Sprintf("Markets: %t\n", p.Markets))
	sb.WriteString(fmt.Sprintf("Durations: %t\n", p.Durations))
	sb.WriteString(fmt.Sprintf("EuropeanOptions: %t\n", p.EuropeanOptions))
	sb.WriteString(fmt.Sprintf("CommissionQuotePercent: %t\n", p.CommissionQuotePercent))
	sb.WriteString(fmt.Sprintf("CommissionTradeFixed: %t\n", p.CommissionTradeFixed))
	sb.WriteString(fmt.Sprintf("CommissionUpdatePercent: %t\n", p.CommissionUpdatePercent))
	sb.WriteString(fmt.Sprintf("CommissionSettleFixed: %t\n", p.CommissionSettleFixed))
	sb.WriteString(fmt.Sprintf("SettleIncentive: %t\n", p.SettleIncentive))
	sb.WriteString(fmt.Sprintf("FreezeTime: %t\n", p.FreezeTime))
	sb.WriteString(fmt.Sprintf("HaltTime: %s\n", p.HaltTime))
	return sb.String()
}
