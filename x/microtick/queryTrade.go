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

type ResponseTradeStatus struct {
    Id MicrotickId `json:"id"`
    Market MicrotickMarket `json:"market"`
    Duration MicrotickDuration `json:"duration"`
    Type MicrotickTradeType `json:"type"`
    Commission MicrotickCoin `json:"commission"`
    CounterParties []DataCounterParty `json:"counterParties"`
    Long MicrotickAccount `json:"long"`
    Backing MicrotickCoin `json:"backing"`
    Premium MicrotickCoin `json:"premium"` 
    RequestedQuantity MicrotickQuantity `json:"requestedQuantity"`
    FilledQuantity MicrotickQuantity `json:"quantity"`
    Start time.Time `json:"start"`
    Expiration time.Time `json:"expiration"`
    Strike MicrotickSpot `json:"strike"`
}

func (rat ResponseTradeStatus) String() string {
    cpStrings := make([]string, len(rat.CounterParties))
    for i := 0; i < len(rat.CounterParties); i++ {
        cpStrings[i] = formatCounterParty(rat.CounterParties[i])
    }
    return strings.TrimSpace(fmt.Sprintf(`Trade Id: %d
Long: %s
Market: %s
Duration: %s
Type: %s
Start: %s
Expiration: %s
Strike: %s 
Requested Quantity: %s
Filled Quantity: %s
Backing: %s
Premium: %s
Commission: %s
Counter Parties: %s`,
    rat.Id, 
    rat.Long, 
    rat.Market, 
    MicrotickDurationNameFromDur(rat.Duration),
    MicrotickTradeTypeToString(rat.Type),
    rat.Start.String(),
    rat.Expiration.String(),
    rat.Strike.String(),
    rat.RequestedQuantity.String(),
    rat.FilledQuantity.String(),
    rat.Backing.String(), 
    rat.Premium.String(),
    rat.Commission.String(),
    cpStrings))
}

func formatCounterParty(cpData DataCounterParty) string {
    return fmt.Sprintf(`
    Short: %s
    Backing: %s
    Paid Premium: %s
    FilledQuantity: %s`,
        cpData.Short.String(),
        cpData.Backing.String(),
        cpData.PaidPremium.String(),
        cpData.FilledQuantity.String(),
    )
}

func queryTradeStatus(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
    var id int
    id, err2 := strconv.Atoi(path[0])
    if err2 != nil {
        panic("invalid trade id")
    }
    data, err2 := keeper.GetActiveTrade(ctx, MicrotickId(id))
    if err2 != nil {
        panic("could not fetch trade data")
    }
    //dataMarket, err3 := keeper.GetDataMarket(ctx, data.Market)
    //if err3 != nil {
        //panic("could not fetch market consensus")
    //}
    
    response := ResponseTradeStatus {
        Id: data.Id,
        Market: data.Market,
        Duration: data.Duration,
        Type: data.Type,
        Commission: data.Commission,
        CounterParties: data.CounterParties,
        Long: data.Long,
        Backing: data.Backing,
        Premium: data.Premium,
        RequestedQuantity: data.RequestedQuantity,
        FilledQuantity: data.FilledQuantity,
        Start: data.Start,
        Expiration: data.Expiration,
        Strike: data.Strike,
    }
    
    bz, err2 := codec.MarshalJSONIndent(keeper.cdc, response)
    if err2 != nil {
        panic("could not marshal result to JSON")
    }
    
    return bz, nil
}
