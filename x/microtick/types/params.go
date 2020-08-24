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

// Params defines the parameters for the microtick module.
type Params struct {
  EuropeanOptions bool `json:"european_options"`
  CommissionQuotePercent sdk.Dec `json:"commission_quote_percent"`
  CommissionTradeFixed sdk.Dec `json:"commission_trade_fixed"`
  CommissionUpdatePercent sdk.Dec `json:"commission_update_percent"`
  CommissionSettleFixed sdk.Dec `json:"commission_settle_fixed"`
  CommissionCancelPercent sdk.Dec `json:"commission_cancel_percent"`
  SettleIncentive sdk.Dec `json:"settle_incentive"`
  FreezeTime int8 `json:"freeze_time"`
  HaltTime string `json:"halt_time"`
  MintDenom string `json:"mint_denom"`
  MintRatio sdk.Dec `json:"mint_ratio"`
  CancelSlashRate sdk.Dec `json:"cancel_slash_rate"`
}

// Default parameter values
var (
  DefaultEuropeanOptions bool = true
  DefaultCommissionQuotePercent = sdk.MustNewDecFromStr("0.0005")
  DefaultCommissionTradeFixed = sdk.MustNewDecFromStr("0.025")
  DefaultCommissionUpdatePercent = sdk.MustNewDecFromStr("0.00005")
  DefaultSettleIncentive = sdk.MustNewDecFromStr("0.025")
  DefaultCommissionSettleFixed = sdk.MustNewDecFromStr("0.01")
  DefaultCommissionCancelPercent = sdk.MustNewDecFromStr("0.001")
  DefaultFreezeTime = int8(30)
  DefaultMintDenom = "utick"
  DefaultMintRatio = sdk.MustNewDecFromStr("0.5")
  DefaultCancelSlashRate = sdk.MustNewDecFromStr("0.01")
)

// Parameter keys
var (
  KeyEuropeanOptions = []byte("EuropeanOptions")
  KeyCommissionQuotePercent = []byte("CommissionQuotePercent")
  KeyCommissionTradeFixed = []byte("KeyCommissionTradeFixed")
  KeyCommissionUpdatePercent = []byte("KeyCommissionUpdatePercent")
  KeyCommissionSettleFixed = []byte("KeyCommissionSettleFixed")
  KeyCommissionCancelPercent = []byte("KeyCommissionCancelPercent")
  KeySettleIncentive = []byte("KeySettleIncentive")
  KeyFreezeTime = []byte("KeyFreezeTime")
  KeyHaltTime = []byte("KeyHaltTime")
  KeyMintDenom = []byte("KeyMintDenom")
  KeyMintRatio = []byte("KeyMintRatio")
  KeyCancelSlashRate = []byte("KeyCancelSlashRate")
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
    {KeyEuropeanOptions, &p.EuropeanOptions, validateEuropeanOptions},
    {KeyCommissionQuotePercent, &p.CommissionQuotePercent, validatePercent},
    {KeyCommissionTradeFixed, &p.CommissionTradeFixed, validateFixed},
    {KeyCommissionUpdatePercent, &p.CommissionUpdatePercent, validatePercent},
    {KeyCommissionSettleFixed, &p.CommissionSettleFixed, validateFixed},
    {KeyCommissionCancelPercent, &p.CommissionCancelPercent, validatePercent},
    {KeySettleIncentive, &p.SettleIncentive, validateFixed},
    {KeyFreezeTime, &p.FreezeTime, validateFreezeTime},
    {KeyHaltTime, &p.HaltTime, validateTime},
    {KeyMintDenom, &p.MintDenom, validateMintDenom},
    {KeyMintRatio, &p.MintRatio, validateMintRatio},
    {KeyCancelSlashRate, &p.CancelSlashRate, validateSlash},
	}
}

func (p Params) ValidateBasic() error {
  if p.CommissionQuotePercent.IsNegative() || p.CommissionQuotePercent.GT(sdk.OneDec()) {
    return fmt.Errorf("invalid quote commission: %s", p.CommissionQuotePercent)
  }
  if p.CommissionTradeFixed.IsNegative() {
    return fmt.Errorf("invalid trade commission: %s", p.CommissionTradeFixed)
  }
  if p.CommissionUpdatePercent.IsNegative() || p.CommissionUpdatePercent.GT(sdk.OneDec()) {
    return fmt.Errorf("invalid quote update commission: %s", p.CommissionUpdatePercent)
  }
  if p.CommissionSettleFixed.IsNegative() {
    return fmt.Errorf("invalid settle commission: %s", p.CommissionSettleFixed)
  }
  if p.CommissionCancelPercent.IsNegative() {
    return fmt.Errorf("invalid cancel commission: %s", p.CommissionCancelPercent)
  }
  if p.SettleIncentive.IsNegative() {
    return fmt.Errorf("invalid settle incentive: %s", p.SettleIncentive)
  }
  if p.MintRatio.IsNegative() {
    return fmt.Errorf("invalid mint ratio: %s", p.MintRatio)
  }
  if p.CancelSlashRate.IsNegative() || p.CancelSlashRate.GT(sdk.OneDec()) {
    return fmt.Errorf("invalid cancel slash rate: %s", p.CancelSlashRate)
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

func validateSlash(i interface{}) error {
	_, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	return nil
}

func validateMintRatio(i interface{}) error {
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

func validateMintDenom(i interface{}) error {
  v, ok := i.(string)
  if !ok {
    return fmt.Errorf("invalid parameter type: %T", i)
  }

  if strings.TrimSpace(v) == "" {
    return errors.New("mint denom cannot be blank")
  }
  if err := sdk.ValidateDenom(v); err != nil {
    return err
  }

  return nil
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
    //interval, _ := time.ParseDuration(HaltTimeString)
    //defaultHaltTime := time.Now().UTC().Add(interval)
    defaultHaltTime, _ := time.Parse("2006-Jan-02", "2030-Jan-01")
	return Params{
    EuropeanOptions: DefaultEuropeanOptions,
    CommissionQuotePercent: DefaultCommissionQuotePercent,
    CommissionTradeFixed: DefaultCommissionTradeFixed,
    CommissionUpdatePercent: DefaultCommissionUpdatePercent,
    CommissionSettleFixed: DefaultCommissionSettleFixed,
    CommissionCancelPercent: DefaultCommissionCancelPercent,
    SettleIncentive: DefaultSettleIncentive,
    FreezeTime: DefaultFreezeTime,
    HaltTime: defaultHaltTime.Format(TimeFormat),
    MintDenom: DefaultMintDenom,
    MintRatio: DefaultMintRatio,
    CancelSlashRate: DefaultCancelSlashRate,
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
	sb.WriteString(fmt.Sprintf("CommissionCancelPercent: %t\n", p.CommissionCancelPercent))
	sb.WriteString(fmt.Sprintf("SettleIncentive: %t\n", p.SettleIncentive))
	sb.WriteString(fmt.Sprintf("FreezeTime: %t\n", p.FreezeTime))
	sb.WriteString(fmt.Sprintf("HaltTime: %s\n", p.HaltTime))
	sb.WriteString(fmt.Sprintf("MintDenom: %s\n", p.MintDenom))
	sb.WriteString(fmt.Sprintf("MintRatio: %t\n", p.MintRatio))
	sb.WriteString(fmt.Sprintf("CancelSlashRate: %t\n", p.CancelSlashRate))
	return sb.String()
}
