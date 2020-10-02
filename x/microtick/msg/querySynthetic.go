package msg

import (
    "context"
    
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    
    mt "gitlab.com/microtick/mtzone/x/microtick/types"
)

func (querier Querier) Synthetic(c context.Context, req *QuerySyntheticRequest) (*QuerySyntheticResponse, error) {
    ctx := sdk.UnwrapSDKContext(c)
    market := req.Market
    durName := req.Duration
    
    dataMarket, err := querier.Keeper.GetDataMarket(ctx, market)
    if err != nil {
        return nil, sdkerrors.Wrap(mt.ErrInvalidMarket, market)
    }
    
    syntheticBook := querier.Keeper.GetSyntheticBook(ctx, &dataMarket, durName, nil)
    
    asks := make([]*SyntheticQuote, len(syntheticBook.Asks))
    bids := make([]*SyntheticQuote, len(syntheticBook.Bids))
    for i := 0; i < len(syntheticBook.Asks); i++ {
        asks[i] = &SyntheticQuote {
            AskId: syntheticBook.Asks[i].AskId,
            AskFill: syntheticBook.Asks[i].AskFill,
            BidId: syntheticBook.Asks[i].BidId,
            BidFill: syntheticBook.Asks[i].BidFill,
            Spot: syntheticBook.Asks[i].Spot,
            Quantity: syntheticBook.Asks[i].Quantity,
        }
    }
    for i := 0; i < len(syntheticBook.Bids); i++ {
        bids[i] = &SyntheticQuote {
            AskId: syntheticBook.Bids[i].AskId,
            AskFill: syntheticBook.Bids[i].AskFill,
            BidId: syntheticBook.Bids[i].BidId,
            BidFill: syntheticBook.Bids[i].BidFill,
            Spot: syntheticBook.Bids[i].Spot,
            Quantity: syntheticBook.Bids[i].Quantity,
        }
    }
    response := QuerySyntheticResponse {
        Consensus:  dataMarket.Consensus,
        Weight: syntheticBook.Weight,
        Asks: asks,
        Bids: bids,
    }
    
    return &response, nil
}
