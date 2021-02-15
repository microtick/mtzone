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
  DefaultCommissionCreatePerunit = sdk.MustNewDecFromStr("0.0004")
  DefaultCommissionTradeFixed = sdk.MustNewDecFromStr("0.025")
  DefaultCommissionUpdatePerunit = sdk.MustNewDecFromStr("0.00005")
  DefaultCommissionSettleFixed = sdk.MustNewDecFromStr("0.01")
  DefaultCommissionCancelPerunit = sdk.MustNewDecFromStr("0.0001")
  DefaultSettleIncentive = sdk.MustNewDecFromStr("0.025")
  DefaultFreezeTime = int32(30)
  DefaultMintDenom = "utick"
  DefaultMintRewardCreatePerunit = sdk.MustNewDecFromStr("200") // utick per unit backing = 0.0004 * 1000000 / 2
  DefaultMintRewardUpdatePerunit = sdk.MustNewDecFromStr("25") // utick per unit backint = 0.00005 * 1000000 / 2
  DefaultMintRewardTradeFixed = sdk.MustNewDecFromStr("0")
  DefaultMintRewardSettleFixed = sdk.MustNewDecFromStr("0")
  DefaultCancelSlashRate = sdk.MustNewDecFromStr("0.01")
)

// Parameter keys
var (
  KeyEuropeanOptions = []byte("EuropeanOptions")
  KeyCommissionCreatePerunit = []byte("CommissionCreatePerunit")
  KeyCommissionTradeFixed = []byte("KeyCommissionTradeFixed")
  KeyCommissionUpdatePerunit = []byte("KeyCommissionUpdatePerunit")
  KeyCommissionSettleFixed = []byte("KeyCommissionSettleFixed")
  KeyCommissionCancelPerunit = []byte("KeyCommissionCancelPerunit")
  KeySettleIncentive = []byte("KeySettleIncentive")
  KeyFreezeTime = []byte("KeyFreezeTime")
  KeyHaltTime = []byte("KeyHaltTime")
  KeyMintDenom = []byte("KeyMintDenom")
  KeyMintRewardCreatePerunit = []byte("KeyMintRewardCreatePerunit")
  KeyMintRewardUpdatePerunit = []byte("KeyMintRewardUpdatePerunit")
  KeyMintRewardTradeFixed = []byte("KeyMintRewardTradeFixed")
  KeyMintRewardSettleFixed = []byte("KeyMintRewardSettleFixed")
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
    {KeyCommissionCreatePerunit, &p.CommissionCreatePerunit, validatePerunit},
    {KeyCommissionTradeFixed, &p.CommissionTradeFixed, validateFixed},
    {KeyCommissionUpdatePerunit, &p.CommissionUpdatePerunit, validatePerunit},
    {KeyCommissionSettleFixed, &p.CommissionSettleFixed, validateFixed},
    {KeyCommissionCancelPerunit, &p.CommissionCancelPerunit, validatePerunit},
    {KeySettleIncentive, &p.SettleIncentive, validateFixed},
    {KeyFreezeTime, &p.FreezeTime, validateFreezeTime},
    {KeyHaltTime, &p.HaltTime, validateTime},
    {KeyMintDenom, &p.MintDenom, validateMintDenom},
    {KeyMintRewardCreatePerunit, &p.MintRewardCreatePerunit, validatePerunit},
    {KeyMintRewardUpdatePerunit, &p.MintRewardUpdatePerunit, validatePerunit},
    {KeyMintRewardTradeFixed, &p.MintRewardTradeFixed, validateFixed},
    {KeyMintRewardSettleFixed, &p.MintRewardSettleFixed, validateFixed},
    {KeyCancelSlashRate, &p.CancelSlashRate, validateSlash},
	}
}

func (p MicrotickParams) ValidateBasic() error {
  if p.CommissionCreatePerunit.IsNegative() || p.CommissionCreatePerunit.GT(sdk.OneDec()) {
    return fmt.Errorf("invalid create commission: %s", p.CommissionCreatePerunit)
  }
  if p.CommissionTradeFixed.IsNegative() {
    return fmt.Errorf("invalid trade commission: %s", p.CommissionTradeFixed)
  }
  if p.CommissionUpdatePerunit.IsNegative() || p.CommissionUpdatePerunit.GT(sdk.OneDec()) {
    return fmt.Errorf("invalid update commission: %s", p.CommissionUpdatePerunit)
  }
  if p.CommissionSettleFixed.IsNegative() {
    return fmt.Errorf("invalid settle commission: %s", p.CommissionSettleFixed)
  }
  if p.CommissionCancelPerunit.IsNegative() {
    return fmt.Errorf("invalid cancel commission: %s", p.CommissionCancelPerunit)
  }
  if p.SettleIncentive.IsNegative() {
    return fmt.Errorf("invalid settle incentive: %s", p.SettleIncentive)
  }
  if p.MintRewardCreatePerunit.IsNegative() {
    return fmt.Errorf("invalid create reward: %s", p.MintRewardCreatePerunit)
  }
  if p.MintRewardUpdatePerunit.IsNegative() {
    return fmt.Errorf("invalid update reward: %s", p.MintRewardUpdatePerunit)
  }
  if p.MintRewardTradeFixed.IsNegative() {
    return fmt.Errorf("invalid trade reward: %s", p.MintRewardTradeFixed)
  }
  if p.MintRewardSettleFixed.IsNegative() {
    return fmt.Errorf("invalid settle reward: %s", p.MintRewardSettleFixed)
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

func validatePerunit(i interface{}) error {
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
    CommissionCreatePerunit: DefaultCommissionCreatePerunit,
    CommissionTradeFixed: DefaultCommissionTradeFixed,
    CommissionUpdatePerunit: DefaultCommissionUpdatePerunit,
    CommissionSettleFixed: DefaultCommissionSettleFixed,
    CommissionCancelPerunit: DefaultCommissionCancelPerunit,
    SettleIncentive: DefaultSettleIncentive,
    FreezeTime: DefaultFreezeTime,
    HaltTime: defaultHaltTime.Format(TimeFormat),
    MintDenom: DefaultMintDenom,
    MintRewardCreatePerunit: DefaultMintRewardCreatePerunit,
    MintRewardUpdatePerunit: DefaultMintRewardUpdatePerunit,
    MintRewardTradeFixed: DefaultMintRewardTradeFixed,
    MintRewardSettleFixed: DefaultMintRewardSettleFixed,
    CancelSlashRate: DefaultCancelSlashRate,
	}
}
