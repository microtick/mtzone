package msg 

import (
    "context"
    
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    
    mt "github.com/mjackson001/mtzone/x/microtick/types"
)

func (querier Querier) Quote(c context.Context, req *QueryQuoteRequest) (*QueryQuoteResponse, error) {
    ctx := sdk.UnwrapSDKContext(c)
    id := req.Id
    data, err := querier.Keeper.GetActiveQuote(ctx, mt.MicrotickId(id))
    if err != nil {
        return nil, sdkerrors.Wrapf(mt.ErrInvalidQuote, "fetching %d", id)
    }
    dataMarket, err := querier.Keeper.GetDataMarket(ctx, data.Market)
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
        Delta: data.Spot.Amount.Sub(dataMarket.Consensus.Amount).QuoInt64(2),
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