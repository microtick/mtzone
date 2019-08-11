package microtick

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
)

type GenesisState struct {
    Params Params `json:"params"`
    Pool MicrotickCoin `json:"commission_pool"`
}

func NewGenesisState(params Params, pool MicrotickCoin) GenesisState {
    return GenesisState {
        Params: params,
        Pool: pool,
    }
}

func DefaultGenesisState() GenesisState {
    return NewGenesisState(DefaultParams(), NewMicrotickCoinFromInt(0))
}

func InitGenesis(ctx sdk.Context, keeper Keeper, data GenesisState) {
    keeper.SetParams(ctx, data.Params)
    
    store := ctx.KVStore(keeper.storeKeys.AppGlobals)
    key := []byte("commissionPool")
    
    store.Set(key, keeper.cdc.MustMarshalBinaryBare(data.Pool))
}

func ExportGenesis(ctx sdk.Context, keeper Keeper) GenesisState {
    store := ctx.KVStore(keeper.storeKeys.AppGlobals)
    key := []byte("commissionPool")
    var pool MicrotickCoin = NewMicrotickCoinFromInt(0)
    if store.Has(key) {
        bz := store.Get(key)
        keeper.cdc.MustUnmarshalBinaryBare(bz, &pool)
    }
    
    params := keeper.GetParams(ctx)
    
    return NewGenesisState(params, pool)
}
