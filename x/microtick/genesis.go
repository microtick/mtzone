package microtick

import (
    "fmt"
    "time"
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/x/distribution/types"
    distr "github.com/cosmos/cosmos-sdk/x/distribution"
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
    
    fmt.Printf("Prearranged halt time: %s\n", time.Unix(data.Params.HaltTime, 0).String())
}

func ExportGenesis(ctx sdk.Context, keeper Keeper, distrKeeper distr.Keeper) GenesisState {
    distrKeeper.IterateValidatorOutstandingRewards(ctx, 
        func(addr sdk.ValAddress, rewards types.ValidatorOutstandingRewards) (stop bool) {
            fmt.Printf("Reward: %+v\n", rewards)
            return false
        },
    )
    
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
