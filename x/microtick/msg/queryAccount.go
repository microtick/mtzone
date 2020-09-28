package msg

import (
    "context"
    sdk "github.com/cosmos/cosmos-sdk/types"
    mt "github.com/mjackson001/mtzone/x/microtick/types"
)

func (querier Querier) Account(c context.Context, req *QueryAccountRequest) (*QueryAccountResponse, error) {
    ctx := sdk.UnwrapSDKContext(c)
    address := req.Account
    dai, tick := querier.Keeper.GetTotalBalance(ctx, address)
    data := querier.Keeper.GetAccountStatus(ctx, address)
    activeQuotes := make([]mt.MicrotickId, len(data.ActiveQuotes.Data))
    activeTrades := make([]mt.MicrotickId, len(data.ActiveTrades.Data))
    for i := 0; i < len(data.ActiveQuotes.Data); i++ {
        activeQuotes[i] = data.ActiveQuotes.Data[i].Id
    }
    for i := 0; i < len(data.ActiveTrades.Data); i++ {
        activeTrades[i] = data.ActiveTrades.Data[i].Id
    }
    response := QueryAccountResponse {
        Account: address,
        Balances: []mt.FractCoin {
            mt.FractCoin {
                Denom: "dai",
                Amount: dai,
            },
            mt.FractCoin {
                Denom: "tick",
                Amount: tick,
            },
        },
        PlacedQuotes: data.PlacedQuotes,
        PlacedTrades: data.PlacedTrades,
        ActiveQuotes: activeQuotes,
        ActiveTrades: activeTrades,
        QuoteBacking: data.QuoteBacking,
        TradeBacking: data.TradeBacking,
        SettleBacking: data.SettleBacking,
    }
    
    return &response, nil
}
