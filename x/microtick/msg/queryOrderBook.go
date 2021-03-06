package msg

import (
    "fmt"
    "strings"
    
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    abci "github.com/tendermint/tendermint/abci/types"
    
    mt "gitlab.com/microtick/mtzone/x/microtick/types"
    "gitlab.com/microtick/mtzone/x/microtick/keeper"
)

type ResponseOrderBook struct {
    SumBacking mt.MicrotickCoin `json:"sumBacking"`
    SumWeight mt.MicrotickQuantity `json:"sumWeight"`
    Calls []mt.MicrotickId `json:"calls"`
    Puts []mt.MicrotickId `json:"puts"`
}

func (rma ResponseOrderBook) String() string {
    return strings.TrimSpace(fmt.Sprintf(`Sum Backing: %s
SumWeight: %s
Calls: %v
Puts: %v`, rma.SumBacking, rma.SumWeight, rma.Calls, rma.Puts))
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
    
    calls := make([]mt.MicrotickId, len(orderBook.Calls.Data))
    puts := make([]mt.MicrotickId, len(orderBook.Puts.Data))
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
    
    bz, err := codec.MarshalJSONIndent(keeper.Cdc, response)
    if err != nil {
        panic("Could not marshal result to JSON")
    }
    
    return bz, nil
}
