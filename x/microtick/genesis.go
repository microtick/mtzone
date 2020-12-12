package microtick

import (
    "fmt"
    sdk "github.com/cosmos/cosmos-sdk/types"
    
    mt "gitlab.com/microtick/mtzone/x/microtick/types"
    "gitlab.com/microtick/mtzone/x/microtick/keeper"
)

func GenesisAccountFromDataAccountStatus(das keeper.DataAccountStatus) mt.GenesisAccount {
    var ga mt.GenesisAccount
    ga.Account = das.Account
    ga.PlacedQuotes = das.PlacedQuotes
    ga.PlacedTrades = das.PlacedTrades
    return ga
}

func InitGenesis(ctx sdk.Context, mtKeeper keeper.Keeper, data mt.GenesisMicrotick) {
    logger := mtKeeper.Logger(ctx)
    
   	mtKeeper.SetExtTokenType(ctx, data.ExtDenom)
	mtKeeper.SetExtPerInt(ctx, data.ExtPerInt)
    logger.Info(fmt.Sprintf("External token type: %s", mtKeeper.GetExtTokenType(ctx)))
    logger.Info(fmt.Sprintf("ExtPerInt: %d", mtKeeper.GetExtPerInt(ctx)))
	
    mtKeeper.SetParams(ctx, data.Params)
    
    for _, acct := range data.Accounts {
        status := mtKeeper.GetAccountStatus(ctx, acct.Account)
        status.PlacedQuotes = acct.PlacedQuotes
        status.PlacedTrades = acct.PlacedTrades
        mtKeeper.SetAccountStatus(ctx, acct.Account, status)
    }
    
    for i, dur := range data.Durations {
        logger.Info(fmt.Sprintf("Genesis Duration %d: %s %d\n", i, dur.Name, dur.Seconds))
        mtKeeper.AddDuration(ctx, dur.Name, dur.Seconds)
    }
    
	for _, market := range data.Markets {
        logger.Info(fmt.Sprintf("Genesis Market: %s \"%s\"\n", market.Name, market.Description))
	    mtKeeper.SetDataMarket(ctx, keeper.NewDataMarket(market.Name, market.Description))
	}
}

func ExportGenesis(ctx sdk.Context, mtKeeper keeper.Keeper) mt.GenesisMicrotick {
    params := mtKeeper.GetParams(ctx)
    
    var accounts []mt.GenesisAccount
    mtKeeper.IterateAccountStatus(ctx, 
        func(acct *keeper.DataAccountStatus) (stop bool) {
            genAcct := GenesisAccountFromDataAccountStatus(*acct)
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
    mtKeeper.IterateMarkets(ctx, func(market *keeper.DataMarket) (stop bool) {
            markets = append(markets, mt.GenesisMarket{
                Name: market.Market,
                Description: market.Description,
            })
            return false
        },
    )
    
    return mt.NewGenesisState(mtKeeper.GetExtTokenType(ctx), uint32(mtKeeper.GetExtPerInt(ctx)), 
      params, accounts, markets, durations)
}
