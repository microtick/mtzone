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
        CommissionQuotePercent: params.CommissionQuotePercent.String(),
        CommissionTradeFixed: params.CommissionTradeFixed.String(),
        CommissionUpdatePercent: params.CommissionUpdatePercent.String(),
        CommissionSettleFixed: params.CommissionSettleFixed.String(),
        CommissionCancelPercent: params.CommissionCancelPercent.String(),
        SettleIncentive: params.SettleIncentive.String(),
        FreezeTime: params.FreezeTime,
        HaltTime: params.HaltTime,
        MintDenom: params.MintDenom,
        MintRatio: params.MintRatio.String(),
        CancelSlashRate: params.CancelSlashRate.String(),
        BackingDenom: extDenom,
        BackingRatio: extPerInt,
    }
    
    return &response, nil
}
