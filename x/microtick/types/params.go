package types

import (
	"fmt"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/x/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultParamspace defines the default microtick module parameter subspace
const DefaultParamspace = "mtmparams"
const TimeFormat = "2006-01-02T15:04:05Z"

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

// Parameter keys
var (
    KeyEuropeanOptions = []byte("EuropeanOptions")
    KeyCommissionQuotePercent = []byte("CommissionQuotePercent")
    KeyCommissionTradeFixed = []byte("KeyCommissionTradeFixed")
    KeyCommissionUpdatePercent = []byte("KeyCommissionUpdatePercent")
    KeyCommissionSettleFixed = []byte("KeyCommissionSettleFixed")
    KeySettleIncentive = []byte("KeySettleIncentive")
    KeyFreezeTime = []byte("KeyFreezeTime")
    KeyHaltTime = []byte("KeyHaltTime")
)

var _ params.ParamSet = &Params{}

// Params defines the parameters for the microtick module.
type Params struct {
    EuropeanOptions bool `json:"european_options"`
    CommissionQuotePercent sdk.Dec `json:"commission_quote_percent"`
    CommissionTradeFixed sdk.Dec `json:"commission_trade_fixed"`
    CommissionUpdatePercent sdk.Dec `json:"commission_update_percent"`
    CommissionSettleFixed sdk.Dec `json:"commission_settle_fixed"`
    SettleIncentive sdk.Dec `json:"settle_incentive"`
    FreezeTime int8 `json:"freeze_time"`
    HaltTime string `json:"halt_time"`
}

// ParamKeyTable for microtick module
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of microtick module's parameters.
// nolint
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
	    {KeyEuropeanOptions, &p.EuropeanOptions},
	    {KeyCommissionQuotePercent, &p.CommissionQuotePercent},
	    {KeyCommissionTradeFixed, &p.CommissionTradeFixed},
	    {KeyCommissionUpdatePercent, &p.CommissionUpdatePercent},
	    {KeyCommissionSettleFixed, &p.CommissionSettleFixed},
	    {KeySettleIncentive, &p.SettleIncentive},
	    {KeyFreezeTime, &p.FreezeTime},
	    {KeyHaltTime, &p.HaltTime},
	}
}

// Equal returns a boolean determining if two Params types are identical.
func (p Params) Equal(p2 Params) bool {
    if p.EuropeanOptions != p2.EuropeanOptions {
        return false
    }
    return true
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
    interval, _ := time.ParseDuration("8h")
    defaultHaltTime := time.Now().UTC().Add(interval)
	return Params{
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
