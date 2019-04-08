package microtick 

import (
    "fmt"
    "strconv"
    "strings"
    
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
    abci "github.com/tendermint/tendermint/abci/types"
)

type ResponseQuoteStatus struct {
    Id MicrotickId `json:"id"`
    Market MicrotickMarket `json:"market"`
    Duration MicrotickDuration `json:"duration"`
    Provider MicrotickAccount `json:"provider"`
    Backing MicrotickCoin `json:"backing"`
    Spot MicrotickSpot `json:"spot"`
    Premium MicrotickPremium `json:"premium"`
    Quantity MicrotickQuantity `json:"quantity"`
    PremiumAsCall MicrotickPremium `json:"premiumAsCall"`
    PremiumAsPut MicrotickPremium `json:"premiumAsPut"`
}

func (raq ResponseQuoteStatus) String() string {
    return strings.TrimSpace(fmt.Sprintf(`Quote Id: %d
Provider: %s
Market: %s
Duration: %s
Backing: %s
Spot: %s
Premium: %s
Quantity: %s
PremiumAsCall: %s
PremiumAsPut: %s`, 
    raq.Id, 
    raq.Provider, 
    raq.Market, 
    MicrotickDurationNameFromDur(raq.Duration),
    raq.Backing.String(), 
    raq.Spot.String(),
    raq.Premium.String(),
    raq.Quantity.String(),
    raq.PremiumAsCall.String(),
    raq.PremiumAsPut.String()))
}

func queryQuoteStatus(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
    var id int
    id, err2 := strconv.Atoi(path[0])
    if err2 != nil {
        return nil, sdk.ErrInternal("Invalid quote ID")
    }
    data, err2 := keeper.GetActiveQuote(ctx, MicrotickId(id))
    if err2 != nil {
        return nil, sdk.ErrInternal("Could not fetch quote data")
    }
    dataMarket, err3 := keeper.GetDataMarket(ctx, data.Market)
    if err3 != nil {
        return nil, sdk.ErrInternal("Could not fetch market consensus")
    }
    
    response := ResponseQuoteStatus {
        Id: data.Id,
        Market: data.Market,
        Duration: data.Duration,
        Provider: data.Provider,
        Backing: data.Backing,
        Spot: data.Spot,
        Premium: data.Premium,
        Quantity: data.Quantity,
        PremiumAsCall: data.PremiumAsCall(dataMarket.Consensus),
        PremiumAsPut: data.PremiumAsPut(dataMarket.Consensus),
    }
    
    bz, err2 := codec.MarshalJSONIndent(keeper.cdc, response)
    if err2 != nil {
        panic("Could not marshal result to JSON")
    }
    
    return bz, nil
}