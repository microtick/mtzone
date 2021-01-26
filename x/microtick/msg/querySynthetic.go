package msg

import (
    "context"
    
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    
    "gitlab.com/microtick/mtzone/x/microtick/keeper"
    mt "gitlab.com/microtick/mtzone/x/microtick/types"
)

func (querier Querier) Synthetic(c context.Context, req *QuerySyntheticRequest) (*QuerySyntheticResponse, error) {
    ctx := sdk.UnwrapSDKContext(c)
    return baseQuerySynthetic(ctx, querier.Keeper, req)
}

func baseQuerySynthetic(ctx sdk.Context, keeper keeper.Keeper, req *QuerySyntheticRequest) (*QuerySyntheticResponse, error) {
    market := req.Market
    durName := req.Duration
    
    dataMarket, err := keeper.GetDataMarket(ctx, market)
    if err != nil {
        return nil, sdkerrors.Wrap(mt.ErrInvalidMarket, market)
    }
    
    syntheticBook := keeper.GetSyntheticBook(ctx, &dataMarket, durName, nil)
    
    asks := make([]*SyntheticQuote, len(syntheticBook.Asks))
    bids := make([]*SyntheticQuote, len(syntheticBook.Bids))
    for i := 0; i < len(syntheticBook.Asks); i++ {
        asks[i] = &SyntheticQuote {
            Spot: syntheticBook.Asks[i].Spot,
            Quantity: syntheticBook.Asks[i].Quantity,
        }
    }
    for i := 0; i < len(syntheticBook.Bids); i++ {
        bids[i] = &SyntheticQuote {
            Spot: syntheticBook.Bids[i].Spot,
            Quantity: syntheticBook.Bids[i].Quantity,
        }
    }
    response := QuerySyntheticResponse {
        Consensus:  dataMarket.Consensus,
        SumBacking: syntheticBook.SumBacking,
        SumWeight: syntheticBook.SumWeight,
        Asks: asks,
        Bids: bids,
    }
    
    return &response, nil
}
