package msg

import (
    "context"
    sdk "github.com/cosmos/cosmos-sdk/types"
    
    "gitlab.com/microtick/mtzone/x/microtick/keeper"
)

func (querier Querier) Params(c context.Context, req *QueryParamsRequest) (*QueryParamsResponse, error) {
    ctx := sdk.UnwrapSDKContext(c)
    return baseQueryParams(ctx, querier.Keeper, req)
}

func baseQueryParams(ctx sdk.Context, keeper keeper.Keeper, req *QueryParamsRequest) (*QueryParamsResponse, error) {
    params := keeper.GetParams(ctx)
    extDenom := keeper.GetExtTokenType(ctx)
    extPerInt := keeper.GetExtPerInt(ctx)
    
    response := QueryParamsResponse {
        EuropeanOptions: params.EuropeanOptions,
        CommissionCreatePerunit: params.CommissionCreatePerunit.String(),
        CommissionTradeFixed: params.CommissionTradeFixed.String(),
        CommissionUpdatePerunit: params.CommissionUpdatePerunit.String(),
        CommissionSettleFixed: params.CommissionSettleFixed.String(),
        CommissionCancelPerunit: params.CommissionCancelPerunit.String(),
        SettleIncentive: params.SettleIncentive.String(),
        FreezeTime: params.FreezeTime,
        MintDenom: params.MintDenom,
        MintRewardCreatePerunit: params.MintRewardCreatePerunit.String(),
        MintRewardUpdatePerunit: params.MintRewardUpdatePerunit.String(),
        MintRewardTradeFixed: params.MintRewardTradeFixed.String(),
        MintRewardSettleFixed: params.MintRewardSettleFixed.String(),
        CancelSlashRate: params.CancelSlashRate.String(),
        BackingDenom: extDenom,
        BackingRatio: extPerInt,
    }
    
    return &response, nil
}
