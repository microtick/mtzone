package msg 

import (
    "fmt"
    "strconv"
    "strings"
    "time"
    
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    abci "github.com/tendermint/tendermint/abci/types"
    
    mt "github.com/mjackson001/mtzone/x/microtick/types"
    "github.com/mjackson001/mtzone/x/microtick/keeper"
)

type ResponseQuoteStatus struct {
    Id mt.MicrotickId `json:"id"`
    Provider mt.MicrotickAccount `json:"provider"`
    Market mt.MicrotickMarket `json:"market"`
    Duration mt.MicrotickDurationName `json:"duration"`
    Backing mt.MicrotickCoin `json:"backing"`
    Ask mt.MicrotickPremium `json:"ask"`
    Bid mt.MicrotickPremium `json:"bid"`
    Quantity mt.MicrotickQuantity `json:"quantity"`
    Consensus mt.MicrotickSpot `json:"consensus"`
    Spot mt.MicrotickSpot `json:"spot"`
    Delta sdk.Dec `json:"delta"`
    CallAsk mt.MicrotickPremium `json:"callAsk"`
    CallBid mt.MicrotickPremium `json:"callBid"`
    PutAsk mt.MicrotickPremium `json:"putAsk"`
    PutBid mt.MicrotickPremium `json:"putBid"`
    Modified time.Time `json:"modified"`
    CanModify time.Time `json:"canModify"`
}

func (rqs ResponseQuoteStatus) String() string {
    return strings.TrimSpace(fmt.Sprintf(`Quote Id: %d
Provider: %s
Market: %s
Duration: %s
Backing: %s
Ask: %s
Bid: %s
Quantity: %s
Current Consensus: %s
Quoted Spot: %s
Delta/2: %spremium
CallAsk: %s
CallBid: %s
PutAsk: %s
PutBid: %s
Modified: %s
CanModify: %s`, 
    rqs.Id, 
    rqs.Provider, 
    rqs.Market, 
    rqs.Duration,
    rqs.Backing.String(), 
    rqs.Ask.String(),
    rqs.Bid.String(),
    rqs.Quantity.String(),
    rqs.Consensus.String(),
    rqs.Spot.String(),
    rqs.Delta.String(),
    rqs.CallAsk.String(),
    rqs.CallBid.String(),
    rqs.PutAsk.String(),
    rqs.PutBid.String(),
    rqs.Modified.String(),
    rqs.CanModify.String()))
}

func QueryQuoteStatus(ctx sdk.Context, path []string, req abci.RequestQuery, keeper keeper.Keeper) (res []byte, err error) {
    id, err := strconv.Atoi(path[0])
    if err != nil {
        return nil, sdkerrors.Wrapf(mt.ErrInvalidQuote, "%d", id)
    }
    data, err := keeper.GetActiveQuote(ctx, mt.MicrotickId(id))
    if err != nil {
        return nil, sdkerrors.Wrapf(mt.ErrInvalidQuote, "fetching %d", id)
    }
    dataMarket, err := keeper.GetDataMarket(ctx, data.Market)
    if err != nil {
        return nil, sdkerrors.Wrap(mt.ErrInvalidMarket, data.Market)
    }
    
    response := ResponseQuoteStatus {
        Id: data.Id,
        Market: data.Market,
        Duration: data.DurationName,
        Provider: data.Provider,
        Backing: data.Backing,
        Spot: data.Spot,
        Consensus: dataMarket.Consensus,
        Delta: data.Spot.Amount.Sub(dataMarket.Consensus.Amount).QuoInt64(2),
        Ask: data.Ask,
        Bid: data.Bid,
        Quantity: data.Quantity,
        CallBid: data.CallBid(dataMarket.Consensus),
        CallAsk: data.CallAsk(dataMarket.Consensus),
        PutBid: data.PutBid(dataMarket.Consensus),
        PutAsk: data.PutAsk(dataMarket.Consensus),
        Modified: data.Modified,
        CanModify: data.CanModify,
    }
    
    bz, err := codec.MarshalJSONIndent(keeper.Cdc, response)
    if err != nil {
        panic("Could not marshal result to JSON")
    }
    
    return bz, nil
}