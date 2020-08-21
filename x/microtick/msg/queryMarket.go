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

type ResponseMarketStatus struct {
    Market mt.MicrotickMarket `json:"market"`
    Description string `json:"description"`
    Consensus mt.MicrotickSpot `json:"consensus"`
    OrderBooks []ResponseMarketOrderBookStatus `json:"orderBooks"`
    SumBacking mt.MicrotickCoin `json:"sumBacking"`
    SumWeight mt.MicrotickQuantity `json:"sumWeight"`
}

type ResponseMarketOrderBookStatus struct {
    Name string `json:"name"`
    SumBacking mt.MicrotickCoin `json:"sumBacking"`
    SumWeight mt.MicrotickQuantity `json:"sumWeight"`
    InsideAsk mt.MicrotickSpot `json:"insideAsk"`
    InsideBid mt.MicrotickSpot `json:"insideBid"`
    InsideCallAsk mt.MicrotickPremium `json:"insideCallAsk"`
    InsideCallBid mt.MicrotickPremium `json:"insideCallBid"`
    InsidePutAsk mt.MicrotickPremium `json:"insidePutAsk"`
    InsidePutBid mt.MicrotickPremium `json:"insidePutBid"`
}

func (rm ResponseMarketStatus) String() string {
    var obStrings []string
    for i := 0; i < len(rm.OrderBooks); i++ {
        obStrings = append(obStrings, formatOrderBook(rm.OrderBooks[i]))
    }
    return strings.TrimSpace(fmt.Sprintf(`Market: %s
Description: %s
Consensus: %s
Orderbooks: %s
Sum Backing: %s
Sum Weight: %s`, rm.Market, rm.Description, rm.Consensus.String(), obStrings, rm.SumBacking.String(),
    rm.SumWeight.String()))
}

func formatOrderBook(rob ResponseMarketOrderBookStatus) string {
    return fmt.Sprintf(`
  %s:
    Sum Backing: %s
    Sum Weight: %s
    InsideAsk: %s
    InsideBid: %s
    Inside Call Ask: %s
    Inside Call Bid: %s
    Inside Put Ask: %s
    Inside Put Bid: %s`, 
        rob.Name,
        rob.SumBacking.String(), rob.SumWeight.String(),
        rob.InsideAsk.String(), rob.InsideBid.String(),
        rob.InsideCallAsk.String(), rob.InsideCallBid.String(),
        rob.InsidePutAsk.String(), rob.InsidePutBid.String())
}

func QueryMarketStatus(ctx sdk.Context, path []string, req abci.RequestQuery, keeper keeper.Keeper) (res []byte, err error) {
    market := path[0]
    data, err := keeper.GetDataMarket(ctx, market)
    if err != nil {
        return nil, sdkerrors.Wrap(mt.ErrInvalidMarket, market)
    }
    
    var orderbookStatus []ResponseMarketOrderBookStatus
    for k := 0; k < len(data.OrderBooks); k++ {
        if data.OrderBooks[k].SumBacking.Amount.GT(sdk.ZeroDec()) {
            callask, _ := keeper.GetActiveQuote(ctx, data.OrderBooks[k].CallAsks.First().Id)
            callbid, _ := keeper.GetActiveQuote(ctx, data.OrderBooks[k].CallBids.Last().Id)
            putask, _ := keeper.GetActiveQuote(ctx, data.OrderBooks[k].PutAsks.First().Id)
            putbid, _ := keeper.GetActiveQuote(ctx, data.OrderBooks[k].PutBids.Last().Id)
            CA := callask.CallAsk(data.Consensus)
            CB := callbid.CallBid(data.Consensus)
            PA := putask.PutAsk(data.Consensus)
            PB := putbid.PutBid(data.Consensus)
            orderbookStatus = append(orderbookStatus, ResponseMarketOrderBookStatus {
                Name: data.OrderBooks[k].Name,
                SumBacking: data.OrderBooks[k].SumBacking,
                SumWeight: data.OrderBooks[k].SumWeight,
                InsideAsk: mt.NewMicrotickSpotFromDec(data.Consensus.Amount.Add(CA.Amount).Sub(PB.Amount)),
                InsideBid: mt.NewMicrotickSpotFromDec(data.Consensus.Amount.Sub(PA.Amount).Add(CB.Amount)),
                InsideCallAsk: CA,
                InsideCallBid: CB,
                InsidePutAsk: PA,
                InsidePutBid: PB,
            })
        }
    }
    
    response := ResponseMarketStatus {
        Market: data.Market,
        Description: data.Description,
        Consensus: data.Consensus,
        OrderBooks: orderbookStatus,
        SumBacking: data.SumBacking,
        SumWeight: data.SumWeight,
    }
    
    bz, err := codec.MarshalJSONIndent(keeper.Cdc, response)
    if err != nil {
        panic("Could not marshal result to JSON")
    }
    
    return bz, nil
}
