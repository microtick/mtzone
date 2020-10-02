package msg

import (
    "context"
    
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    
    mt "gitlab.com/microtick/mtzone/x/microtick/types"
)

func (querier Querier) OrderBook(c context.Context, req *QueryOrderBookRequest) (*QueryOrderBookResponse, error) {
    ctx := sdk.UnwrapSDKContext(c)    
    market := req.Market
    durName := req.Duration
    
    dataMarket, err := querier.Keeper.GetDataMarket(ctx, market)
    if err != nil {
        return nil, sdkerrors.Wrap(mt.ErrInvalidMarket, market)
    }
    
    orderBook := dataMarket.GetOrderBook(durName)
    
    callasks := make([]*OrderBookQuote, len(orderBook.CallAsks.Data))
    callbids := make([]*OrderBookQuote, len(orderBook.CallBids.Data))
    putasks := make([]*OrderBookQuote, len(orderBook.PutAsks.Data))
    putbids := make([]*OrderBookQuote, len(orderBook.PutBids.Data))
    for i := 0; i < len(orderBook.CallAsks.Data); i++ {
        quote, _ := querier.Keeper.GetActiveQuote(ctx, orderBook.CallAsks.Data[i].Id)
        callasks[i] = &OrderBookQuote {
            Id: quote.Id,
            Premium: quote.CallAsk(dataMarket.Consensus),
            Quantity: quote.Quantity,
        }
    }
    for i := 0; i < len(orderBook.CallBids.Data); i++ {
        j := len(orderBook.CallBids.Data) - i - 1
        quote, _ := querier.Keeper.GetActiveQuote(ctx, orderBook.CallBids.Data[j].Id)
        callbids[i] = &OrderBookQuote {
            Id: quote.Id,
            Premium: quote.CallBid(dataMarket.Consensus),
            Quantity: quote.Quantity,
        }
    }
    for i := 0; i < len(orderBook.PutAsks.Data); i++ {
        quote, _ := querier.Keeper.GetActiveQuote(ctx, orderBook.PutAsks.Data[i].Id)
        putasks[i] = &OrderBookQuote {
            Id: quote.Id,
            Premium: quote.PutAsk(dataMarket.Consensus),
            Quantity: quote.Quantity,
        }
    }
    for i := 0; i < len(orderBook.PutBids.Data); i++ {
        j := len(orderBook.PutBids.Data) - i - 1
        quote, _ := querier.Keeper.GetActiveQuote(ctx, orderBook.PutBids.Data[j].Id)
        putbids[i] = &OrderBookQuote {
            Id: quote.Id,
            Premium: quote.PutBid(dataMarket.Consensus),
            Quantity: quote.Quantity,
        }
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
