package msg

import (
    "context"
    
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    
    "gitlab.com/microtick/mtzone/x/microtick/keeper"
    mt "gitlab.com/microtick/mtzone/x/microtick/types"
)

func (querier Querier) OrderBook(c context.Context, req *QueryOrderBookRequest) (*QueryOrderBookResponse, error) {
    ctx := sdk.UnwrapSDKContext(c)    
    return baseQueryOrderBook(ctx, querier.Keeper, req)
}

func baseQueryOrderBook(ctx sdk.Context, keeper keeper.Keeper, req *QueryOrderBookRequest) (*QueryOrderBookResponse, error) {
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
    
    orderBook := dataMarket.GetOrderBook(durName)
    
    callasks := make([]*OrderBookQuote, 0)
    callbids := make([]*OrderBookQuote, 0)
    putasks := make([]*OrderBookQuote, 0)
    putbids := make([]*OrderBookQuote, 0)
    var count uint32
    var i int
    count = 0
    for i = int(req.Offset); i < len(orderBook.CallAsks.Data) && count < req.Limit; i++ {
        count = count + 1
        quote, _ := keeper.GetActiveQuote(ctx, orderBook.CallAsks.Data[i].Id)
        callasks = append(callasks, &OrderBookQuote {
            Id: quote.Id,
            Premium: quote.CallAsk(dataMarket.Consensus),
            Quantity: quote.Quantity,
        })
    }
    count = 0
    for i = int(req.Offset); i < len(orderBook.CallBids.Data) && count < req.Limit; i++ {
        count = count + 1
        j := len(orderBook.CallBids.Data) - i - 1
        quote, _ := keeper.GetActiveQuote(ctx, orderBook.CallBids.Data[j].Id)
        callbids = append(callbids, &OrderBookQuote {
            Id: quote.Id,
            Premium: quote.CallBid(dataMarket.Consensus),
            Quantity: quote.Quantity,
        })
    }
    count = 0
    for i = int(req.Offset); i < len(orderBook.PutAsks.Data) && count < req.Limit; i++ {
        count = count + 1
        quote, _ := keeper.GetActiveQuote(ctx, orderBook.PutAsks.Data[i].Id)
        putasks = append(putasks, &OrderBookQuote {
            Id: quote.Id,
            Premium: quote.PutAsk(dataMarket.Consensus),
            Quantity: quote.Quantity,
        })
    }
    count = 0
    for i = int(req.Offset); i < len(orderBook.PutBids.Data) && count < req.Limit; i++ {
        count = count + 1
        j := len(orderBook.PutBids.Data) - i - 1
        quote, _ := keeper.GetActiveQuote(ctx, orderBook.PutBids.Data[j].Id)
        putbids = append(putbids, &OrderBookQuote {
            Id: quote.Id,
            Premium: quote.PutBid(dataMarket.Consensus),
            Quantity: quote.Quantity,
        })
    }
    response := QueryOrderBookResponse {
        Consensus: dataMarket.Consensus,
        SumBacking: orderBook.SumBacking,
        SumWeight: orderBook.SumWeight,
        CallAsks: callasks,
        CallBids: callbids,
        PutAsks: putasks,
        PutBids: putbids,
    }
    
    return &response, nil
}
