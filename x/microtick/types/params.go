package types

import (
	"fmt"
	"errors"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// DefaultParamspace defines the default microtick module parameter subspace
const TimeFormat = "2006-01-02T15:04:05Z"
const HaltTimeString = "168h"

// Default parameter values
var (
  DefaultEuropeanOptions bool = true
  DefaultCommissionQuotePercent = sdk.MustNewDecFromStr("0.0004")
  DefaultCommissionTradeFixed = sdk.MustNewDecFromStr("0.025")
  DefaultCommissionUpdatePercent = sdk.MustNewDecFromStr("0.00005")
  DefaultSettleIncentive = sdk.MustNewDecFromStr("0.025")
  DefaultCommissionSettleFixed = sdk.MustNewDecFromStr("0.01")
  DefaultCommissionCancelPercent = sdk.MustNewDecFromStr("0.0001")
  DefaultFreezeTime = int32(30)
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
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&MicrotickParams{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of microtick module's parameters.
// nolint
func (p *MicrotickParams) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
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

func (p MicrotickParams) ValidateBasic() error {
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
	v, ok := i.(int32)
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
func DefaultParams() MicrotickParams {
    //interval, _ := time.ParseDuration(HaltTimeString)
    //defaultHaltTime := time.Now().UTC().Add(interval)
    defaultHaltTime, _ := time.Parse("2006-Jan-02", "2030-Jan-01")
	return MicrotickParams{
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
