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
    
    if req.Limit == 0 {
        req.Limit = 10
    }
    if req.Limit > 100 {
        req.Limit = 100
    }
    
    backing, tick := keeper.GetTotalBalance(ctx, address)
    data := keeper.GetAccountStatus(ctx, address)
    
    activeQuotes := make([]mt.MicrotickId, 0)
    activeTrades := make([]mt.MicrotickId, 0)
    var count uint32
    var i int
    count = 0
    for i = int(req.Offset); i < len(data.ActiveQuotes.Data) && count < req.Limit; i++ {
        count = count + 1
        activeQuotes = append(activeQuotes, data.ActiveQuotes.Data[i].Id)
    }
    count = 0
    for i = int(req.Offset); i < len(data.ActiveTrades.Data) && count < req.Limit; i++ {
        count = count + 1
        activeTrades = append(activeTrades, data.ActiveTrades.Data[i].Id)
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
