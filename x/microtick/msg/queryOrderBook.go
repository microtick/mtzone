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
    CallAsks []mt.MicrotickId `json:"callasks"`
    CallBids []mt.MicrotickId `json:"callbids"`
    PutAsks []mt.MicrotickId `json:"putasks"`
    PutBids []mt.MicrotickId `json:"putbids"`
}

func (rma ResponseOrderBook) String() string {
    return strings.TrimSpace(fmt.Sprintf(`Sum Backing: %s
SumWeight: %s
CallAsks: %v
CallBids: %v
PutAsks: %v
PutBids: %v`, rma.SumBacking, rma.SumWeight, rma.CallAsks, rma.CallBids, rma.PutAsks, rma.PutBids))
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
    
    callasks := make([]mt.MicrotickId, len(orderBook.CallAsks.Data))
    callbids := make([]mt.MicrotickId, len(orderBook.CallBids.Data))
    putasks := make([]mt.MicrotickId, len(orderBook.PutAsks.Data))
    putbids := make([]mt.MicrotickId, len(orderBook.PutBids.Data))
    for i := 0; i < len(orderBook.CallAsks.Data); i++ {
        callasks[i] = orderBook.CallAsks.Data[i].Id
    }
    for i := 0; i < len(orderBook.CallBids.Data); i++ {
        j := len(orderBook.CallBids.Data) - i - 1
        callbids[i] = orderBook.CallBids.Data[j].Id
    }
    for i := 0; i < len(orderBook.PutAsks.Data); i++ {
        putasks[i] = orderBook.PutAsks.Data[i].Id
    }
    for i := 0; i < len(orderBook.PutBids.Data); i++ {
        j := len(orderBook.PutBids.Data) - i - 1
        putbids[i] = orderBook.PutBids.Data[j].Id
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
