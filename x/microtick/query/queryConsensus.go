package query

import (
    "fmt"
    "strings"
    
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
    abci "github.com/tendermint/tendermint/abci/types"
)

type ResponseMarketConsensus struct {
    Market MicrotickMarket `json:"market"`
    Consensus MicrotickSpot `json:"consensus"`
    SumBacking MicrotickCoin `json:"sumBacking"`
    SumWeight MicrotickQuantity `json:"sumWeight"`
}

func (rm ResponseMarketConsensus) String() string {
    return strings.TrimSpace(fmt.Sprintf(`Market: %s
Consensus: %s
Sum Backing: %s
Sum Weight: %s`, rm.Market, rm.Consensus.String(), rm.SumBacking.String(),
    rm.SumWeight.String()))
}

func queryMarketConsensus(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
    market := path[0]
    data, err2 := keeper.GetDataMarket(ctx, market)
    if err2 != nil {
        return nil, sdk.ErrInternal(fmt.Sprintf("Could not fetch market data: %s", err2))
    }
    
    response := ResponseMarketConsensus {
        Market: data.Market,
        Consensus: data.Consensus,
        SumBacking: data.SumBacking,
        SumWeight: data.SumWeight,
    }
    
    bz, err2 := codec.MarshalJSONIndent(keeper.cdc, response)
    if err2 != nil {
        panic("Could not marshal result to JSON")
    }
    
    return bz, nil
}
