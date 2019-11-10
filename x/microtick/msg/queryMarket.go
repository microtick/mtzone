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

type ResponseMarketOrderBookStatus struct {
    SumBacking mt.MicrotickCoin `json:"sumBacking"`
    SumWeight mt.MicrotickQuantity `json:"sumWeight"`
    InsideCall mt.MicrotickPremium `json:"insideCall"`
    InsidePut mt.MicrotickPremium `json:"insidePut"`
}

type ResponseMarketStatus struct {
    Market mt.MicrotickMarket `json:"market"`
    Consensus mt.MicrotickSpot `json:"consensus"`
    OrderBooks []ResponseMarketOrderBookStatus `json:"orderBooks"`
    SumBacking mt.MicrotickCoin `json:"sumBacking"`
    SumWeight mt.MicrotickQuantity `json:"sumWeight"`
}

func (rm ResponseMarketStatus) String() string {
    obStrings := make([]string, len(mt.MicrotickDurations))
    for i := 0; i < len(mt.MicrotickDurations); i++ {
        obStrings[i] = formatOrderBook(mt.MicrotickDurations[i], rm.OrderBooks[i])
    }
    return strings.TrimSpace(fmt.Sprintf(`Market: %s
Consensus: %s
Orderbooks: %s
Sum Backing: %s
Sum Weight: %s`, rm.Market, rm.Consensus.String(), obStrings, rm.SumBacking.String(),
    rm.SumWeight.String()))
}

func formatOrderBook(dur mt.MicrotickDuration, rob ResponseMarketOrderBookStatus) string {
    return fmt.Sprintf(`
  %s:
    Sum Backing: %s
    Sum Weight: %s
    Inside Call: %s
    Inside Put: %s`, 
        mt.MicrotickDurationNameFromDur(dur),
        rob.SumBacking.String(), rob.SumWeight.String(),
        rob.InsideCall.String(), rob.InsidePut.String())
}

func queryMarketStatus(ctx sdk.Context, path []string, req abci.RequestQuery, keeper keeper.MicrotickKeeper) (res []byte, err sdk.Error) {
    market := path[0]
    data, err2 := keeper.GetDataMarket(ctx, market)
    if err2 != nil {
        return nil, sdk.ErrInternal(fmt.Sprintf("Could not fetch market data: %s", err2))
    }
    
    orderbookStatus := make([]ResponseMarketOrderBookStatus, len(mt.MicrotickDurations))
    for i := 0; i < len(mt.MicrotickDurations); i++ {
        orderbookStatus[i].SumBacking = data.OrderBooks[i].SumBacking
        orderbookStatus[i].SumWeight = data.OrderBooks[i].SumWeight
        
        if len(data.OrderBooks[i].Calls.Data) > 0 {
            call, _ := keeper.GetActiveQuote(ctx, data.OrderBooks[i].Calls.Data[0].Id)
            orderbookStatus[i].InsideCall = call.PremiumAsCall(data.Consensus)
        }
        if len(data.OrderBooks[i].Puts.Data) > 0 {
            put, _ := keeper.GetActiveQuote(ctx, data.OrderBooks[i].Puts.Data[0].Id)
            orderbookStatus[i].InsidePut = put.PremiumAsPut(data.Consensus)
        }
    }
    
    response := ResponseMarketStatus {
        Market: data.Market,
        Consensus: data.Consensus,
        OrderBooks: orderbookStatus,
        SumBacking: data.SumBacking,
        SumWeight: data.SumWeight,
    }
    
    bz, err2 := codec.MarshalJSONIndent(ModuleCdc, response)
    if err2 != nil {
        panic("Could not marshal result to JSON")
    }
    
    return bz, nil
}
