package microtick

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/x/params"
)

// DefaultParamspace defines the default microtick module parameter subspace
const DefaultParamspace = "mtmparams"

// Default parameter values
const (
    DefaultEuropeanOptions bool = true
)

// Parameter keys
var (
    KeyEuropeanOptions = []byte("EuropeanOptions")
)

var _ params.ParamSet = &Params{}

// Params defines the parameters for the microtick module.
type Params struct {
    EuropeanOptions bool `json:"european_options"`
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
	return Params{
	    EuropeanOptions: DefaultEuropeanOptions,
	}
}

// String implements the stringer interface.
func (p Params) String() string {
	var sb strings.Builder
	sb.WriteString("Params: \n")
	sb.WriteString(fmt.Sprintf("EuropeanOptions: %t\n", p.EuropeanOptions))
	return sb.String()
}
