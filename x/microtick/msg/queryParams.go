package msg

import (
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
    abci "github.com/tendermint/tendermint/abci/types"
    
    "github.com/mjackson001/mtzone/x/microtick/keeper"
)

func QueryParams(ctx sdk.Context, path []string, req abci.RequestQuery, keeper keeper.Keeper) (res []byte, err error) {
    params := keeper.GetParams(ctx)
    
    bz, err := codec.MarshalJSONIndent(keeper.Cdc, params)
    if err != nil {
        panic("Could not marshal result to JSON")
    }
    
    return bz, nil
}
