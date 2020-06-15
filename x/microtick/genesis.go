package microtick

import (
    "fmt"
    sdk "github.com/cosmos/cosmos-sdk/types"
    
    mt "github.com/mjackson001/mtzone/x/microtick/types"
    "github.com/mjackson001/mtzone/x/microtick/keeper"
)

type GenesisAccount struct {
    Account mt.MicrotickAccount `json:"account"`
    NumQuotes uint32 `json:"numQuotes"`
    NumTrades uint32 `json:"numTrades"`
}

type GenesisMarket struct {
    Name mt.MicrotickMarket `json:"name"`
    Description string `json:"description"`
}

type GenesisDuration struct {
    Name mt.MicrotickDurationName `json:"name"`
    Seconds mt.MicrotickDuration `json:"seconds"`
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
    Accounts []GenesisAccount `json:"accounts"`
    Markets []GenesisMarket `json:"markets"`
    Durations []GenesisDuration `json:"durations"`
}

func NewGenesisState(params mt.Params, accounts []GenesisAccount, markets []GenesisMarket, durations []GenesisDuration) GenesisState {
    return GenesisState {
        Params: params,
        Accounts: accounts,
        Markets: markets,
        Durations: durations,
    }
}

func DefaultGenesisState() GenesisState {
    return NewGenesisState(mt.DefaultParams(), nil, nil, nil)
}

func InitGenesis(ctx sdk.Context, mtKeeper keeper.Keeper, data GenesisState) {
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

func ExportGenesis(ctx sdk.Context, mtKeeper keeper.Keeper) GenesisState {
    params := mtKeeper.GetParams(ctx)
    
    var accounts []GenesisAccount
    mtKeeper.IterateAccountStatus(ctx, 
        func(acct keeper.DataAccountStatus) (stop bool) {
            genAcct := GenesisAccountFromDataAccountStatus(acct)
            accounts = append(accounts, genAcct)
            return false
        },
    )
    
    var durations []GenesisDuration
    mtKeeper.IterateDurations(ctx, func(name mt.MicrotickDurationName, seconds mt.MicrotickDuration) (stop bool) {
            durations = append(durations, GenesisDuration{
                Name: name,
                Seconds: seconds,
            })
            return false
        },
    )
    
    var markets []GenesisMarket
    mtKeeper.IterateMarkets(ctx, func(market keeper.DataMarket) (stop bool) {
            markets = append(markets, GenesisMarket{
                Name: market.Market,
                Description: market.Description,
            })
            return false
        },
    )
    
    return NewGenesisState(params, accounts, markets, durations)
}

func ValidateGenesis(data GenesisState) error {
    return nil
}
