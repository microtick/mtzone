package msg

import (
    "fmt"
    "strings"
    
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    abci "github.com/tendermint/tendermint/abci/types"
    
    mt "github.com/mjackson001/mtzone/x/microtick/types"
    "github.com/mjackson001/mtzone/x/microtick/keeper"
)

type ResponseOrderBook struct {
    SumBacking mt.MicrotickCoin `json:"sumBacking"`
    SumWeight mt.MicrotickQuantity `json:"sumWeight"`
    CallAsks []ResponseOrderBookQuote `json:"callasks"`
    CallBids []ResponseOrderBookQuote `json:"callbids"`
    PutAsks []ResponseOrderBookQuote `json:"putasks"`
    PutBids []ResponseOrderBookQuote `json:"putbids"`
}

type ResponseOrderBookQuote struct {
    Id mt.MicrotickId `json:"id"`
    Premium sdk.Dec `json:"premium"`
    Quantity sdk.Dec `json:"quantity"`
}

func (rma ResponseOrderBook) String() string {
    var i int
    var ca, cb, pa, pb string
    for i = 0; i < len(rma.CallAsks); i++ {
        ca += formatQuote(rma.CallAsks[i]) + "\n"
    }
    for i = 0; i < len(rma.CallBids); i++ {
        cb += formatQuote(rma.CallBids[i]) + "\n"
    }
    for i = 0; i < len(rma.PutAsks); i++ {
        pa += formatQuote(rma.PutAsks[i]) + "\n"
    }
    for i = 0; i < len(rma.PutBids); i++ {
        pb += formatQuote(rma.PutBids[i]) + "\n"
    }
    return strings.TrimSpace(fmt.Sprintf(`Sum Backing: %s
SumWeight: %s
CallAsks: 
%sCallBids: 
%sPutAsks: 
%sPutBids: 
%s`, rma.SumBacking, rma.SumWeight, ca, cb, pa, pb))
}

func formatQuote(robq ResponseOrderBookQuote) string {
    return fmt.Sprintf(`  %d premium: %s quantity: %s`, robq.Id, robq.Premium.String(), robq.Quantity.String())
}

func QueryOrderBook(ctx sdk.Context, path []string, 
    req abci.RequestQuery, keeper keeper.Keeper)(res []byte, err error) {
        
    market := path[0]
    durName := path[1]
    
    dataMarket, err := keeper.GetDataMarket(ctx, market)
    if err != nil {
        return nil, sdkerrors.Wrap(mt.ErrInvalidMarket, market)
    }
    
    orderBook := dataMarket.GetOrderBook(durName)
    
    callasks := make([]ResponseOrderBookQuote, len(orderBook.CallAsks.Data))
    callbids := make([]ResponseOrderBookQuote, len(orderBook.CallBids.Data))
    putasks := make([]ResponseOrderBookQuote, len(orderBook.PutAsks.Data))
    putbids := make([]ResponseOrderBookQuote, len(orderBook.PutBids.Data))
    for i := 0; i < len(orderBook.CallAsks.Data); i++ {
        quote, _ := keeper.GetActiveQuote(ctx, orderBook.CallAsks.Data[i].Id)
        callasks[i] = ResponseOrderBookQuote {
            Id: quote.Id,
            Premium: quote.CallAsk(dataMarket.Consensus).Amount,
            Quantity: quote.Quantity.Amount,
        }
    }
    for i := 0; i < len(orderBook.CallBids.Data); i++ {
        j := len(orderBook.CallBids.Data) - i - 1
        quote, _ := keeper.GetActiveQuote(ctx, orderBook.CallBids.Data[j].Id)
        callbids[i] = ResponseOrderBookQuote {
            Id: quote.Id,
            Premium: quote.CallBid(dataMarket.Consensus).Amount,
            Quantity: quote.Quantity.Amount,
        }
    }
    for i := 0; i < len(orderBook.PutAsks.Data); i++ {
        quote, _ := keeper.GetActiveQuote(ctx, orderBook.PutAsks.Data[i].Id)
        putasks[i] = ResponseOrderBookQuote {
            Id: quote.Id,
            Premium: quote.PutAsk(dataMarket.Consensus).Amount,
            Quantity: quote.Quantity.Amount,
        }
    }
    for i := 0; i < len(orderBook.PutBids.Data); i++ {
        j := len(orderBook.PutBids.Data) - i - 1
        quote, _ := keeper.GetActiveQuote(ctx, orderBook.PutBids.Data[j].Id)
        putbids[i] = ResponseOrderBookQuote {
            Id: quote.Id,
            Premium: quote.PutBid(dataMarket.Consensus).Amount,
            Quantity: quote.Quantity.Amount,
        }
    }
    response := ResponseOrderBook {
        SumBacking: orderBook.SumBacking,
        SumWeight: orderBook.SumWeight,
        CallAsks: callasks,
        CallBids: callbids,
        PutAsks: putasks,
        PutBids: putbids,
    }
    
    bz, err := codec.MarshalJSONIndent(keeper.Cdc, response)
    if err != nil {
        panic("Could not marshal result to JSON")
    }
    
    return bz, nil
}
