package msg

import (
    "fmt"
    "strings"
    "errors"
    
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
    abci "github.com/tendermint/tendermint/abci/types"
    
    mt "github.com/mjackson001/mtzone/x/microtick/types"
    "github.com/mjackson001/mtzone/x/microtick/keeper"
)

type ResponseMarketOrderBookStatus struct {
    Name string `json:"name"`
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
    var obStrings []string
    for i := 0; i < len(rm.OrderBooks); i++ {
        obStrings = append(obStrings, formatOrderBook(rm.OrderBooks[i]))
    }
    return strings.TrimSpace(fmt.Sprintf(`Market: %s
Consensus: %s
Orderbooks: %s
Sum Backing: %s
Sum Weight: %s`, rm.Market, rm.Consensus.String(), obStrings, rm.SumBacking.String(),
    rm.SumWeight.String()))
}

func formatOrderBook(rob ResponseMarketOrderBookStatus) string {
    return fmt.Sprintf(`
  %s:
    Sum Backing: %s
    Sum Weight: %s
    Inside Call: %s
    Inside Put: %s`, 
        rob.Name,
        rob.SumBacking.String(), rob.SumWeight.String(),
        rob.InsideCall.String(), rob.InsidePut.String())
}

func QueryMarketStatus(ctx sdk.Context, path []string, req abci.RequestQuery, keeper keeper.Keeper) (res []byte, err error) {
    market := path[0]
    data, err2 := keeper.GetDataMarket(ctx, market)
    if err2 != nil {
        return nil, errors.New(fmt.Sprintf("Could not fetch market data: %s", err2))
    }
    
    var orderbookStatus []ResponseMarketOrderBookStatus
    for i := 0; i < len(data.OrderBooks); i++ {
        if data.OrderBooks[i].SumBacking.Amount.GT(sdk.ZeroDec()) {
            call, _ := keeper.GetActiveQuote(ctx, data.OrderBooks[i].Calls.Data[0].Id)
            put, _ := keeper.GetActiveQuote(ctx, data.OrderBooks[i].Puts.Data[0].Id)
            orderbookStatus = append(orderbookStatus, ResponseMarketOrderBookStatus {
                Name: mt.MicrotickDurationNames[i],
                SumBacking: data.OrderBooks[i].SumBacking,
                SumWeight: data.OrderBooks[i].SumWeight,
                InsideCall: call.PremiumAsCall(data.Consensus),
                InsidePut: put.PremiumAsPut(data.Consensus),
            })
        }
    }
    
    response := ResponseMarketStatus {
        Market: data.Market,
        Consensus: data.Consensus,
        OrderBooks: orderbookStatus,
        SumBacking: data.SumBacking,
        SumWeight: data.SumWeight,
    }
    
    bz, err2 := codec.MarshalJSONIndent(keeper.Cdc, response)
    if err2 != nil {
        panic("Could not marshal result to JSON")
    }
    
    return bz, nil
}
