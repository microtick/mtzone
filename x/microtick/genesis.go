package microtick

import (
    "fmt"
    "time"
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/x/distribution/types"
    
    mt "github.com/mjackson001/mtzone/x/microtick/types"
    "github.com/mjackson001/mtzone/x/microtick/keeper"
)

type GenesisState struct {
    Params mt.Params `json:"params"`
    Pool mt.MicrotickCoin `json:"commission_pool"`
}

func NewGenesisState(params mt.Params, pool mt.MicrotickCoin) GenesisState {
    return GenesisState {
        Params: params,
        Pool: pool,
    }
}

func DefaultGenesisState() GenesisState {
    return NewGenesisState(mt.DefaultParams(), mt.NewMicrotickCoinFromInt(0))
}

func InitGenesis(ctx sdk.Context, keeper keeper.MicrotickKeeper, data GenesisState) {
    keeper.SetParams(ctx, data.Params)
    
    store := ctx.KVStore(keeper.appGlobalsKey)
    key := []byte("commissionPool")
    
    store.Set(key, keeper.cdc.MustMarshalBinaryBare(data.Pool))
    
    fmt.Printf("Prearranged halt time: %s\n", time.Unix(data.Params.HaltTime, 0).String())
}

func ExportGenesis(ctx sdk.Context, keeper keeper.MicrotickKeeper) GenesisState {
    keeper.distrKeeper.IterateValidatorOutstandingRewards(ctx, 
        func(addr sdk.ValAddress, rewards types.ValidatorOutstandingRewards) (stop bool) {
            fmt.Printf("Reward: %+v\n", rewards)
            return false
        },
    )
    
    store := ctx.KVStore(keeper.appGlobalsKey)
    key := []byte("commissionPool")
    var pool MicrotickCoin = NewMicrotickCoinFromInt(0)
    if store.Has(key) {
        bz := store.Get(key)
        keeper.cdc.MustUnmarshalBinaryBare(bz, &pool)
    }
    
    params := keeper.GetParams(ctx)
    
    return NewGenesisState(params, pool)
}

func ValidateGenesis(data GenesisState) error {
    return nil
}
