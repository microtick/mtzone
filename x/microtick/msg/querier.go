package msg

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/mjackson001/mtzone/x/microtick/keeper"
)

// Querier if used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper
type Querier struct {
	keeper.Keeper
}

var _ GRPCServer = Querier{}

func NewQuerier(keeper keeper.Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		return []byte{}, sdkerrors.ErrUnknownRequest
	}
}
