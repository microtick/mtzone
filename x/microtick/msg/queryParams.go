package msg

import (
    "context"
    sdk "github.com/cosmos/cosmos-sdk/types"
)

func (querier Querier) Params(c context.Context, req *QueryParamsRequest) (*QueryParamsResponse, error) {
    ctx := sdk.UnwrapSDKContext(c)
    params := querier.Keeper.GetParams(ctx)
    
    response := QueryParamsResponse {
        EuropeanOptions: params.EuropeanOptions,
        CommissionQuotePercent: params.CommissionQuotePercent,
        CommissionTradeFixed: params.CommissionTradeFixed,
        CommissionUpdatePercent: params.CommissionUpdatePercent,
        CommissionSettleFixed: params.CommissionSettleFixed,
        CommissionCancelPercent: params.CommissionCancelPercent,
        SettleIncentive: params.SettleIncentive,
        FreezeTime: params.FreezeTime,
        HaltTime: params.HaltTime,
        MintDenom: params.MintDenom,
        MintRatio: params.MintRatio,
        CancelSlashRate: params.CancelSlashRate,
    }
    
    return &response, nil
}
