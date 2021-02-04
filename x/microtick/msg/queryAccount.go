package msg

import (
    "context"
    sdk "github.com/cosmos/cosmos-sdk/types"
    "gitlab.com/microtick/mtzone/x/microtick/keeper"
    mt "gitlab.com/microtick/mtzone/x/microtick/types"
)

func (querier Querier) Account(c context.Context, req *QueryAccountRequest) (*QueryAccountResponse, error) {
    ctx := sdk.UnwrapSDKContext(c)
    return baseQueryAccount(ctx, querier.Keeper, req)
}

func baseQueryAccount(ctx sdk.Context, keeper keeper.Keeper, req* QueryAccountRequest) (*QueryAccountResponse, error) {
    params := keeper.GetParams(ctx)
    address := req.Account
    backing, tick := keeper.GetTotalBalance(ctx, address)
    data := keeper.GetAccountStatus(ctx, address)
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
                Denom: "backing",
                Amount: backing,
            },
            mt.FractCoin {
                Denom: params.MintDenom,
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
