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
    Cost MicrotickCoin `json:"premium"` 
    FilledQuantity MicrotickQuantity `json:"quantity"`
    Start time.Time `json:"start"`
    Expiration time.Time `json:"expiration"`
    Strike MicrotickSpot `json:"strike"`
    CurrentSpot MicrotickSpot `json:"currentSpot"`
    CurrentValue MicrotickCoin `json:"currentValue"`
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
Filled Quantity: %s
Backing: %s
Cost: %s
Commission: %s
Counter Parties: %s
Strike: %s 
Current Spot: %s
Current Value: %s`,
    rat.Id, 
    rat.Long, 
    rat.Market, 
    MicrotickDurationNameFromDur(rat.Duration),
    MicrotickTradeTypeToString(rat.Type),
    rat.Start.String(),
    rat.Expiration.String(),
    rat.FilledQuantity.String(),
    rat.Backing.String(), 
    rat.Cost.String(),
    rat.Commission.String(),
    cpStrings,
    rat.Strike.String(),
    rat.CurrentSpot.String(),
    rat.CurrentValue.String()))
}

func formatCounterParty(cpData DataCounterParty) string {
    return fmt.Sprintf(`
    Short: %s
    Backing: %s
    Cost: %s
    FilledQuantity: %s`,
        cpData.Short.String(),
        cpData.Backing.String(),
        cpData.Cost.String(),
        cpData.FilledQuantity.String(),
    )
}

func queryTradeStatus(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
    var id int
    id, err2 := strconv.Atoi(path[0])
    if err2 != nil {
        return nil, sdk.ErrInternal("Invalid trade ID")
    }
    data, err2 := keeper.GetActiveTrade(ctx, MicrotickId(id))
    if err2 != nil {
        return nil, sdk.ErrInternal("Could not fetch trade data")
    }
    dataMarket, err3 := keeper.GetDataMarket(ctx, data.Market)
    if err3 != nil {
        return nil, sdk.ErrInternal("Could not fetch market consensus")
    }
    
    response := ResponseTradeStatus {
        Id: data.Id,
        Market: data.Market,
        Duration: data.Duration,
        Type: data.Type,
        Commission: data.Commission,
        CounterParties: data.CounterParties,
        Long: data.Long,
        Backing: data.Backing,
        Cost: data.Cost,
        FilledQuantity: data.FilledQuantity,
        Start: data.Start,
        Expiration: data.Expiration,
        Strike: data.Strike,
        CurrentSpot: dataMarket.Consensus,
        CurrentValue: data.CurrentValue(dataMarket.Consensus),
    }
    
    bz, err2 := codec.MarshalJSONIndent(keeper.cdc, response)
    if err2 != nil {
        panic("Could not marshal result to JSON")
    }
    
    return bz, nil
}
