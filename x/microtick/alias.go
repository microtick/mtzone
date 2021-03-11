package microtick

import (
    "github.com/microtick/mtzone/x/microtick/keeper"
    "github.com/microtick/mtzone/x/microtick/types"
)

type (
    Keeper = keeper.Keeper
)

const (
    GlobalsKey = types.GlobalsKey
    AccountStatusKey = types.AccountStatusKey
    ActiveQuotesKey = types.ActiveQuotesKey
    ActiveTradesKey = types.ActiveTradesKey
    MarketsKey = types.MarketsKey
    DurationsKey = types.DurationsKey
)

var (
    NewKeeper = keeper.NewKeeper
)
