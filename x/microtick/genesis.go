package microtick

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/x/distribution/types"
    
    mt "github.com/mjackson001/mtzone/x/microtick/types"
    "github.com/mjackson001/mtzone/x/microtick/keeper"
)

type GenesisAccount struct {
    Account mt.MicrotickAccount `json:"account"`
    NumQuotes uint32 `json:"numQuotes"`
    NumTrades uint32 `json:"numTrades"`
}

func GenesisAccountFromDataAccountStatus(das keeper.DataAccountStatus) GenesisAccount {
    var ga GenesisAccount
    ga.Account = das.Account
    ga.NumQuotes = das.NumQuotes
    ga.NumTrades = das.NumTrades
    return ga
}

type GenesisState struct {
    Params mt.Params `json:"params"`
    Pool mt.MicrotickCoin `json:"commission_pool"`
    Accounts []GenesisAccount `json:"accounts"`
}

func NewGenesisState(params mt.Params, pool mt.MicrotickCoin, 
    accounts []GenesisAccount) GenesisState {
    return GenesisState {
        Params: params,
        Pool: pool,
        Accounts: accounts,
    }
}

func DefaultGenesisState() GenesisState {
    return NewGenesisState(mt.DefaultParams(), mt.NewExtTokenTypeFromInt(0), nil)
}

func InitGenesis(ctx sdk.Context, mtKeeper keeper.Keeper, data GenesisState) {
    mtKeeper.SetParams(ctx, data.Params)
    
    store := ctx.KVStore(mtKeeper.AppGlobalsKey)
    key := []byte(keeper.MTPoolName)
    
    store.Set(key, mtKeeper.GetCodec().MustMarshalBinaryBare(data.Pool))
    
    for _, acct := range data.Accounts {
        status := mtKeeper.GetAccountStatus(ctx, acct.Account)
        status.NumQuotes = acct.NumQuotes
        status.NumTrades = acct.NumTrades
        mtKeeper.SetAccountStatus(ctx, acct.Account, status)
    }
    
    //fmt.Printf("Prearranged halt time: %s\n", data.Params.HaltTime)
}

func ExportGenesis(ctx sdk.Context, mtKeeper keeper.Keeper) GenesisState {
    mtKeeper.DistrKeeper.IterateValidatorOutstandingRewards(ctx, 
        func(addr sdk.ValAddress, rewards types.ValidatorOutstandingRewards) (stop bool) {
            return false
        },
    )
    
    store := ctx.KVStore(mtKeeper.AppGlobalsKey)
    key := []byte(keeper.MTPoolName)
    var pool mt.MicrotickCoin = mt.NewMicrotickCoinFromInt(0)
    if store.Has(key) {
        bz := store.Get(key)
        mtKeeper.GetCodec().MustUnmarshalBinaryBare(bz, &pool)
    }
    
    params := mtKeeper.GetParams(ctx)
    
    var accounts []GenesisAccount
    mtKeeper.IterateAccountStatus(ctx, 
        func(acct keeper.DataAccountStatus) (stop bool) {
            genAcct := GenesisAccountFromDataAccountStatus(acct)
            accounts = append(accounts, genAcct)
            return false
        },
    )
    
    return NewGenesisState(params, pool, accounts)
}

func ValidateGenesis(data GenesisState) error {
    return nil
}
