package microtick 

import (
    "fmt"
    "strconv"
    "strings"
    
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
    abci "github.com/tendermint/tendermint/abci/types"
)

type ResponseActiveQuote struct {
    Id MicrotickId `json:"id"`
    Market MicrotickMarket `json:"market"`
    Duration MicrotickDuration `json:"duration"`
    Provider MicrotickAccount `json:"provider"`
    Backing sdk.Coins `json:"backing"`
    Spot MicrotickSpot `json:"spot"`
    Premium MicrotickPremium `json:"premium"`
    Quantity MicrotickQuantity `json:"quantity"`
}

func (raq ResponseActiveQuote) String() string {
    return strings.TrimSpace(fmt.Sprintf(`Id: %s
Market: %s
Duration: %s
Provider: %s
Backing: %s
Spot: %s
Premium: %s
Quantity: %s`, 
    fmt.Sprintf("%d", raq.Id), 
    raq.Market, 
    fmt.Sprintf("%d", raq.Duration),
    raq.Provider, raq.Backing.String(), 
    FormatMicrotickSpot(raq.Spot),
    FormatMicrotickPremium(raq.Premium),
    FormatMicrotickQuantity(raq.Quantity)))
}

func queryActiveQuote(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
    var id int
    id, err2 := strconv.Atoi(path[0])
    if err2 != nil {
        panic("invalid quote id")
    }
    data, err2 := keeper.GetActiveQuote(ctx, MicrotickId(id))
    if err2 != nil {
        panic("could not fetch quote data")
    }
    
    response := ResponseActiveQuote {
        Id: data.Id,
        Market: data.Market,
        Duration: data.Duration,
        Provider: data.Provider,
        Backing: data.Backing,
        Spot: data.Spot,
        Premium: data.Premium,
        Quantity: data.Quantity,
    }
    
    bz, err2 := codec.MarshalJSONIndent(keeper.cdc, response)
    if err2 != nil {
        panic("could not marshal result to JSON")
    }
    
    return bz, nil
}