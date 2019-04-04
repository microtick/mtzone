package microtick 

import (
    "fmt"
    "strings"
    
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
    abci "github.com/tendermint/tendermint/abci/types"
)

type ResponseMarketOrderBookStatus struct {
    SumBacking MicrotickCoin `json:"sumBacking"`
    SumWeight MicrotickQuantity `json:"sumWeight"`
}

type ResponseMarketStatus struct {
    Market MicrotickMarket `json:"market"`
    Consensus MicrotickSpot `json:"consensus"`
    OrderBooks []ResponseMarketOrderBookStatus `json:"orderBooks"`
    SumBacking MicrotickCoin `json:"sumBacking"`
    SumWeight MicrotickQuantity `json:"sumWeight"`
}

func (rm ResponseMarketStatus) String() string {
    obStrings := make([]string, len(MicrotickDurations))
    for i := 0; i < len(MicrotickDurations); i++ {
        obStrings[i] = formatOrderBook(MicrotickDurations[i], rm.OrderBooks[i])
    }
    return strings.TrimSpace(fmt.Sprintf(`Market: %s
Consensus: %s
Orderbooks: %s
Sum Backing: %s
Sum Weight: %s`, rm.Market, rm.Consensus.String(), obStrings, rm.SumBacking.String(),
    rm.SumWeight.String()))
}

func formatOrderBook(dur MicrotickDuration, rob ResponseMarketOrderBookStatus) string {
    return fmt.Sprintf(`
  %s:
    Sum Backing: %s
    Sum Weight: %s`, 
        MicrotickDurationNameFromDur(dur),
        rob.SumBacking.String(), rob.SumWeight.String())
}

func queryMarketStatus(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
    market := path[0]
    data, err2 := keeper.GetDataMarket(ctx, market)
    if err2 != nil {
        panic("could not fetch market data")
    }
    
    orderbookStatus := make([]ResponseMarketOrderBookStatus, len(MicrotickDurations))
    for i := 0; i < len(MicrotickDurations); i++ {
        orderbookStatus[i].SumBacking = data.OrderBooks[i].SumBacking
        orderbookStatus[i].SumWeight = data.OrderBooks[i].SumWeight
    }
    
    response := ResponseMarketStatus {
        Market: data.Market,
        Consensus: data.Consensus,
        OrderBooks: orderbookStatus,
        SumBacking: data.SumBacking,
        SumWeight: data.SumWeight,
    }
    
    bz, err2 := codec.MarshalJSONIndent(keeper.cdc, response)
    if err2 != nil {
        panic("could not marshal result to JSON")
    }
    
    return bz, nil
}
