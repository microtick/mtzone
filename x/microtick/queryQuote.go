package microtick 

import (
    "fmt"
    "strconv"
    "strings"
    "time"
    
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
    abci "github.com/tendermint/tendermint/abci/types"
)

type ResponseQuoteStatus struct {
    Id MicrotickId `json:"id"`
    Market MicrotickMarket `json:"market"`
    Duration MicrotickDurationName `json:"duration"`
    Provider MicrotickAccount `json:"provider"`
    Backing MicrotickCoin `json:"backing"`
    Spot MicrotickSpot `json:"spot"`
    Premium MicrotickPremium `json:"premium"`
    Quantity MicrotickQuantity `json:"quantity"`
    PremiumAsCall MicrotickPremium `json:"premiumAsCall"`
    PremiumAsPut MicrotickPremium `json:"premiumAsPut"`
    Modified time.Time `json:"modified"`
    CanModify time.Time `json:"canModify"`
}

func (rqs ResponseQuoteStatus) String() string {
    return strings.TrimSpace(fmt.Sprintf(`Quote Id: %d
Provider: %s
Market: %s
Duration: %s
Backing: %s
Spot: %s
Premium: %s
Quantity: %s
PremiumAsCall: %s
PremiumAsPut: %s
Modified: %s
CanModify: %s`, 
    rqs.Id, 
    rqs.Provider, 
    rqs.Market, 
    rqs.Duration,
    rqs.Backing.String(), 
    rqs.Spot.String(),
    rqs.Premium.String(),
    rqs.Quantity.String(),
    rqs.PremiumAsCall.String(),
    rqs.PremiumAsPut.String(),
    rqs.Modified.String(),
    rqs.CanModify.String()))
}

func queryQuoteStatus(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
    var id int
    id, err2 := strconv.Atoi(path[0])
    if err2 != nil {
        return nil, sdk.ErrInternal(fmt.Sprintf("Invalid quote ID: %s", err2))
    }
    data, err2 := keeper.GetActiveQuote(ctx, MicrotickId(id))
    if err2 != nil {
        return nil, sdk.ErrInternal(fmt.Sprintf("Could not fetch quote data: %s", err2))
    }
    dataMarket, err3 := keeper.GetDataMarket(ctx, data.Market)
    if err3 != nil {
        return nil, sdk.ErrInternal(fmt.Sprintf("Could not fetch market consensus: %s", err3))
    }
    
    response := ResponseQuoteStatus {
        Id: data.Id,
        Market: data.Market,
        Duration: MicrotickDurationNameFromDur(data.Duration),
        Provider: data.Provider,
        Backing: data.Backing,
        Spot: data.Spot,
        Premium: data.Premium,
        Quantity: data.Quantity,
        PremiumAsCall: data.PremiumAsCall(dataMarket.Consensus),
        PremiumAsPut: data.PremiumAsPut(dataMarket.Consensus),
        Modified: data.Modified,
        CanModify: data.CanModify,
    }
    
    bz, err2 := codec.MarshalJSONIndent(keeper.cdc, response)
    if err2 != nil {
        panic("Could not marshal result to JSON")
    }
    
    return bz, nil
}