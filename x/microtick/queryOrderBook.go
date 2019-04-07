package microtick

import (
    "fmt"
    "strings"
    
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
    abci "github.com/tendermint/tendermint/abci/types"
)

type ResponseOrderBook struct {
    SumBacking MicrotickCoin `json:"sumBacking"`
    SumWeight MicrotickQuantity `json:"sumWeight"`
    Calls []MicrotickId `json:"calls"`
    Puts []MicrotickId `json:"puts"`
}

func (rma ResponseOrderBook) String() string {
    return strings.TrimSpace(fmt.Sprintf(`Sum Backing: %s
SumWeight: %s
Calls: %v
Puts: %v`, rma.SumBacking, rma.SumWeight, rma.Calls, rma.Puts))
}

func queryOrderBook(ctx sdk.Context, path []string, 
    req abci.RequestQuery, keeper Keeper)(res []byte, err sdk.Error) {
        
    market := path[0]
    dur := NewMicrotickDurationFromString(path[1])
    
    dataMarket, err3 := keeper.GetDataMarket(ctx, market)
    if err3 != nil {
        panic("Invalid market")
    }
    
    orderBook := dataMarket.GetOrderBook(dur)
    
    calls := make([]MicrotickId, len(orderBook.Calls.Data))
    puts := make([]MicrotickId, len(orderBook.Puts.Data))
    for i := 0; i < len(orderBook.Calls.Data); i++ {
        calls[i] = orderBook.Calls.Data[i].Id
    }
    for i := 0; i < len(orderBook.Puts.Data); i++ {
        puts[i] = orderBook.Puts.Data[i].Id
    }
    response := ResponseOrderBook {
        SumBacking: orderBook.SumBacking,
        SumWeight: orderBook.SumWeight,
        Calls: calls,
        Puts: puts,
    }
    
    bz, err4 := codec.MarshalJSONIndent(keeper.cdc, response)
    if err4 != nil {
        panic("could not marshal result to JSON")
    }
    
    return bz, nil
}