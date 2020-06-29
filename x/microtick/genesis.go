package microtick

import (
    "fmt"
    sdk "github.com/cosmos/cosmos-sdk/types"
    
    mt "github.com/mjackson001/mtzone/x/microtick/types"
    "github.com/mjackson001/mtzone/x/microtick/keeper"
)

func GenesisAccountFromDataAccountStatus(das keeper.DataAccountStatus) mt.GenesisAccount {
    var ga mt.GenesisAccount
    ga.Account = das.Account
    ga.NumQuotes = das.NumQuotes
    ga.NumTrades = das.NumTrades
    return ga
}

func InitGenesis(ctx sdk.Context, mtKeeper keeper.Keeper, data mt.GenesisState) {
    mtKeeper.SetParams(ctx, data.Params)
    
    for _, acct := range data.Accounts {
        status := mtKeeper.GetAccountStatus(ctx, acct.Account)
        status.NumQuotes = acct.NumQuotes
        status.NumTrades = acct.NumTrades
        mtKeeper.SetAccountStatus(ctx, acct.Account, status)
    }
    
	durArray := make([]string, len(data.Durations))
	
    for i, dur := range data.Durations {
        fmt.Printf("Genesis Duration %d: %s %d\n", i, dur.Name, dur.Seconds)
        durArray[i] = dur.Name
        mtKeeper.AddDuration(ctx, dur.Name, dur.Seconds)
    }
    
	for _, market := range data.Markets {
        fmt.Printf("Genesis Market: %s \"%s\"\n", market.Name, market.Description)
	    mtKeeper.SetDataMarket(ctx, keeper.NewDataMarket(market.Name, market.Description, durArray))
	}
}

func ExportGenesis(ctx sdk.Context, mtKeeper keeper.Keeper) mt.GenesisState {
    params := mtKeeper.GetParams(ctx)
    
    var accounts []mt.GenesisAccount
    mtKeeper.IterateAccountStatus(ctx, 
        func(acct keeper.DataAccountStatus) (stop bool) {
            genAcct := GenesisAccountFromDataAccountStatus(acct)
            accounts = append(accounts, genAcct)
            return false
        },
    )
    
    var durations []mt.GenesisDuration
    mtKeeper.IterateDurations(ctx, func(name mt.MicrotickDurationName, seconds mt.MicrotickDuration) (stop bool) {
            durations = append(durations, mt.GenesisDuration{
                Name: name,
                Seconds: seconds,
            })
            return false
        },
    )
    
    var markets []mt.GenesisMarket
    mtKeeper.IterateMarkets(ctx, func(market keeper.DataMarket) (stop bool) {
            markets = append(markets, mt.GenesisMarket{
                Name: market.Market,
                Description: market.Description,
            })
            return false
        },
    )
    
    return mt.NewGenesisState(params, accounts, markets, durations)
}
