package microtick 

import (
    "fmt"
    "strings"
    
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
    abci "github.com/tendermint/tendermint/abci/types"
)

type ResponseMarketStatus struct {
    Market MicrotickMarket `json:"market"`
    Consensus MicrotickSpot `json:"consensus"`
    OrderBooks []DataOrderBook `json:"orderBooks"`
    SumBacking MicrotickCoin `json:"sumBacking"`
    SumWeight MicrotickQuantity `json:"sumWeight"`
}

func (rm ResponseMarketStatus) String() string {
    return strings.TrimSpace(fmt.Sprintf(`Market: %s
Consensus: %s
Sum Backing: %s
Sum Weight: %s`, rm.Market, rm.Consensus.String(), rm.SumBacking.String(),
    rm.SumWeight.String()))
}

func queryMarketStatus(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
    market := path[0]
    data, err2 := keeper.GetDataMarket(ctx, market)
    if err2 != nil {
        panic("could not fetch market data")
    }
    
    response := ResponseMarketStatus {
        Market: data.Market,
        Consensus: data.Consensus,
        OrderBooks: data.OrderBooks,
        SumBacking: data.SumBacking,
        SumWeight: data.SumWeight,
    }
    
    bz, err2 := codec.MarshalJSONIndent(keeper.cdc, response)
    if err2 != nil {
        panic("could not marshal result to JSON")
    }
    
    return bz, nil
}