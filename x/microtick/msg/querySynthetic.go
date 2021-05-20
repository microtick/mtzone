package msg

import (
    "context"
    
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    
    "github.com/microtick/mtzone/x/microtick/keeper"
    mt "github.com/microtick/mtzone/x/microtick/types"
)

func (querier Querier) Synthetic(c context.Context, req *QuerySyntheticRequest) (*QuerySyntheticResponse, error) {
    ctx := sdk.UnwrapSDKContext(c)
    return baseQuerySynthetic(ctx, querier.Keeper, req)
}

func baseQuerySynthetic(ctx sdk.Context, keeper keeper.Keeper, req *QuerySyntheticRequest) (*QuerySyntheticResponse, error) {
    market := req.Market
    durName := req.Duration
    
    if req.Limit == 0 {
        req.Limit = 10
    }
    if req.Limit > 100 {
        req.Limit = 100
    }
    
    dataMarket, err := keeper.GetDataMarket(ctx, market)
    if err != nil {
        return nil, sdkerrors.Wrap(mt.ErrInvalidMarket, market)
    }
    
    if !keeper.ValidDurationName(ctx, durName) {
        return nil, sdkerrors.Wrap(mt.ErrInvalidDuration, durName)
    }
    syntheticBook := keeper.GetSyntheticBook(ctx, &dataMarket, durName, nil)
    
    asks := make([]*SyntheticQuote, 0)
    bids := make([]*SyntheticQuote, 0)
    var count uint32
    var i int
    count = 0
    for i = int(req.Offset); i < len(syntheticBook.Asks) && count < req.Limit; i++ {
        count = count + 1
        asks = append(asks, &SyntheticQuote {
            Spot: syntheticBook.Asks[i].Spot,
            Quantity: syntheticBook.Asks[i].Quantity,
        })
    }
    count = 0
    for i = int(req.Offset); i < len(syntheticBook.Bids) && count < req.Limit; i++ {
        count = count + 1
        bids = append(bids, &SyntheticQuote {
            Spot: syntheticBook.Bids[i].Spot,
            Quantity: syntheticBook.Bids[i].Quantity,
        })
    }
    response := QuerySyntheticResponse {
        Consensus:  dataMarket.Consensus,
        SumBacking: syntheticBook.SumBacking,
        SumWeight: syntheticBook.SumWeight,
        Limit: req.Limit,
        Offset: req.Offset,
        TotalAsks: uint32(len(syntheticBook.Asks)),
        TotalBids: uint32(len(syntheticBook.Bids)),
        Asks: asks,
        Bids: bids,
    }
    
    return &response, nil
}
