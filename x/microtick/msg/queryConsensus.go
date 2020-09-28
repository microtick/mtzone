package msg

import (
    "context"
    
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    
    mt "github.com/mjackson001/mtzone/x/microtick/types"
)

func (querier Querier) Consensus(c context.Context, req *QueryConsensusRequest) (*QueryConsensusResponse, error) {
    ctx := sdk.UnwrapSDKContext(c)
    market := req.Market
    data, err := querier.Keeper.GetDataMarket(ctx, market)
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
