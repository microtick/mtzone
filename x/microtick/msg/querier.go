package msg

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"gitlab.com/microtick/mtzone/x/microtick/keeper"
)

// Querier if used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper
type Querier struct {
	keeper.Keeper
}

func NewQuerier(keeper keeper.Keeper, legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
			default:
			  return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown %s query endpoint: %s", types.ModuleName, path[0])
		}
	}
}
