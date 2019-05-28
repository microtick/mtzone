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
    Duration MicrotickDurationName `json:"duration"`
    Type MicrotickTradeTypeName `json:"type"`
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
    Commission MicrotickCoin `json:"commission"`
    SettleIncentive MicrotickCoin `json:"settleIncentive"`
}

func (rts ResponseTradeStatus) String() string {
    cpStrings := make([]string, len(rts.CounterParties))
    for i := 0; i < len(rts.CounterParties); i++ {
        cpStrings[i] = formatCounterParty(rts.CounterParties[i])
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
Settle Incentive: %s
Counterparties: %s
Strike: %s 
Current Spot: %s
Current Value: %s`,
    rts.Id, 
    rts.Long, 
    rts.Market, 
    rts.Duration,
    rts.Type,
    rts.Start.String(),
    rts.Expiration.String(),
    rts.FilledQuantity.String(),
    rts.Backing.String(), 
    rts.Cost.String(),
    rts.Commission.String(),
    rts.SettleIncentive.String(),
    cpStrings,
    rts.Strike.String(),
    rts.CurrentSpot.String(),
    rts.CurrentValue.String()))
}

func formatCounterParty(cpData DataCounterParty) string {
    return fmt.Sprintf(`
    Short: %s
        Quoted: %s
        Backing: %s
        Cost: %s
        Filled Quantity: %s`,
        cpData.Short.String(),
        formatQuoteParams(cpData.Quoted),
        cpData.Backing.String(),
        cpData.Cost.String(),
        cpData.FilledQuantity.String(),
    )
}

func formatQuoteParams(params DataQuoteParams) string {
    return fmt.Sprintf(`
            Id: %d 
            Premium: %s 
            Quantity: %s 
            Spot: %s`,
        params.Id,
        params.Premium.String(),
        params.Quantity.String(),
        params.Spot.String(),
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
        Type: MicrotickTradeNameFromType(data.Type),
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
        Commission: data.Commission,
        SettleIncentive: data.SettleIncentive,
    }
    
    bz, err2 := codec.MarshalJSONIndent(keeper.cdc, response)
    if err2 != nil {
        panic("Could not marshal result to JSON")
    }
    
    return bz, nil
}
