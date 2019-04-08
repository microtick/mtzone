package microtick

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
)

type GenesisState struct {
    Params Params `json:"params"`
}

func NewGenesisState(params Params) GenesisState {
    return GenesisState {
        Params: params,
    }
}

func DefaultGenesisState() GenesisState {
    return NewGenesisState(DefaultParams())
}

func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) {
    keeper.SetParams(ctx, data.Params)
}

func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {
    params := keeper.GetParams(ctx)
    return NewGenesisState(params)
}
