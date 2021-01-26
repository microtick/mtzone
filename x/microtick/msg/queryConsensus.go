package msg

import (
    "context"
    
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    
    "gitlab.com/microtick/mtzone/x/microtick/keeper"
    mt "gitlab.com/microtick/mtzone/x/microtick/types"
)

func (querier Querier) Consensus(c context.Context, req *QueryConsensusRequest) (*QueryConsensusResponse, error) {
    ctx := sdk.UnwrapSDKContext(c)
    return baseQueryConsensus(ctx, querier.Keeper, req)
}

func baseQueryConsensus(ctx sdk.Context, keeper keeper.Keeper, req* QueryConsensusRequest) (*QueryConsensusResponse, error) {
    market := req.Market
    data, err := keeper.GetDataMarket(ctx, market)
    if err != nil {
        return nil, sdkerrors.Wrap(mt.ErrInvalidMarket, market)
    }
    
    response := QueryConsensusResponse { 
        Market: data.Market,
        Consensus: data.Consensus,
        TotalBacking: data.TotalBacking,
        TotalWeight: data.TotalWeight,
    }
    
    return &response, nil
}
