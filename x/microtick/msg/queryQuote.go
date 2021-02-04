package msg 

import (
    "context"
    
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    
    "gitlab.com/microtick/mtzone/x/microtick/keeper"
    mt "gitlab.com/microtick/mtzone/x/microtick/types"
)

func (querier Querier) Quote(c context.Context, req *QueryQuoteRequest) (*QueryQuoteResponse, error) {
    ctx := sdk.UnwrapSDKContext(c)
    return baseQueryQuote(ctx, querier.Keeper, req)
}

func baseQueryQuote(ctx sdk.Context, keeper keeper.Keeper, req *QueryQuoteRequest) (*QueryQuoteResponse, error) {
    id := req.Id
    data, err := keeper.GetActiveQuote(ctx, mt.MicrotickId(id))
    if err != nil {
        return nil, sdkerrors.Wrapf(mt.ErrInvalidQuote, "fetching %d", id)
    }
    dataMarket, err := keeper.GetDataMarket(ctx, data.Market)
    if err != nil {
        return nil, sdkerrors.Wrap(mt.ErrInvalidMarket, data.Market)
    }
    
    response := QueryQuoteResponse {
        Id: data.Id,
        Market: data.Market,
        Duration: data.DurationName,
        Provider: data.Provider,
        Backing: data.Backing,
        Spot: data.Spot,
        Consensus: dataMarket.Consensus,
        Ask: data.Ask,
        Bid: data.Bid,
        Quantity: data.Quantity,
        CallBid: data.CallBid(dataMarket.Consensus),
        CallAsk: data.CallAsk(dataMarket.Consensus),
        PutBid: data.PutBid(dataMarket.Consensus),
        PutAsk: data.PutAsk(dataMarket.Consensus),
        Modified: data.Modified,
        CanModify: data.CanModify,
    }
    
    return &response, nil
}
