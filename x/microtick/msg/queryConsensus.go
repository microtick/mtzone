package msg

import (
    "fmt"
    "strings"
    
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
    abci "github.com/tendermint/tendermint/abci/types"
    
    mt "github.com/mjackson001/mtzone/x/microtick/types"
    "github.com/mjackson001/mtzone/x/microtick/keeper"
)

type ResponseMarketConsensus struct {
    Market mt.MicrotickMarket `json:"market"`
    Consensus mt.MicrotickSpot `json:"consensus"`
    SumBacking mt.MicrotickCoin `json:"sumBacking"`
    SumWeight mt.MicrotickQuantity `json:"sumWeight"`
}

func (rm ResponseMarketConsensus) String() string {
    return strings.TrimSpace(fmt.Sprintf(`Market: %s
Consensus: %s
Sum Backing: %s
Sum Weight: %s`, rm.Market, rm.Consensus.String(), rm.SumBacking.String(),
    rm.SumWeight.String()))
}

func QueryMarketConsensus(ctx sdk.Context, path []string, req abci.RequestQuery, keeper keeper.MicrotickKeeper) (res []byte, err sdk.Error) {
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
    
    bz, err2 := codec.MarshalJSONIndent(ModuleCdc, response)
    if err2 != nil {
        panic("Could not marshal result to JSON")
    }
    
    return bz, nil
}
