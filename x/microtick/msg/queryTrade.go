package msg 

import (
    "fmt"
    "strconv"
    "strings"
    "time"
    
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
    abci "github.com/tendermint/tendermint/abci/types"
    
    mt "github.com/mjackson001/mtzone/x/microtick/types"
    "github.com/mjackson001/mtzone/x/microtick/keeper"
)

type ResponseTradeStatus struct {
    Id mt.MicrotickId `json:"id"`
    Market mt.MicrotickMarket `json:"market"`
    Duration mt.MicrotickDurationName `json:"duration"`
    Type mt.MicrotickTradeTypeName `json:"type"`
    CounterParties []keeper.DataCounterParty `json:"counterParties"`
    Long mt.MicrotickAccount `json:"long"`
    Backing mt.MicrotickCoin `json:"backing"`
    Cost mt.MicrotickCoin `json:"premium"` 
    FilledQuantity mt.MicrotickQuantity `json:"quantity"`
    Start time.Time `json:"start"`
    Expiration time.Time `json:"expiration"`
    Strike mt.MicrotickSpot `json:"strike"`
    CurrentSpot mt.MicrotickSpot `json:"currentSpot"`
    CurrentValue mt.MicrotickCoin `json:"currentValue"`
    Commission mt.MicrotickCoin `json:"commission"`
    SettleIncentive mt.MicrotickCoin `json:"settleIncentive"`
    Balance mt.MicrotickCoin `json:"balance"`
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
Balance: %s
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
    rts.Balance.String(),
    cpStrings,
    rts.Strike.String(),
    rts.CurrentSpot.String(),
    rts.CurrentValue.String()))
}

func formatCounterParty(cpData keeper.DataCounterParty) string {
    return fmt.Sprintf(`
    Short: %s
        Quoted: %s
        Backing: %s
        Cost: %s
        Filled Quantity: %s
        Balance: %s`,
        cpData.Short.String(),
        formatQuoteParams(cpData.Quoted),
        cpData.Backing.String(),
        cpData.Cost.String(),
        cpData.FilledQuantity.String(),
        cpData.Balance.String(),
    )
}

func formatQuoteParams(params keeper.DataQuoteParams) string {
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

func queryTradeStatus(ctx sdk.Context, path []string, req abci.RequestQuery, keeper keeper.MicrotickKeeper) (res []byte, err sdk.Error) {
    var id int
    id, err2 := strconv.Atoi(path[0])
    if err2 != nil {
        return nil, sdk.ErrInternal(fmt.Sprintf("Invalid trade ID: %s", err2))
    }
    data, err2 := keeper.GetActiveTrade(ctx, mt.MicrotickId(id))
    if err2 != nil {
        return nil, sdk.ErrInternal(fmt.Sprintf("Could not fetch trade data: %s", err2))
    }
    dataMarket, err3 := keeper.GetDataMarket(ctx, data.Market)
    if err3 != nil {
        return nil, sdk.ErrInternal(fmt.Sprintf("Could not fetch market consensus: %s", err3))
    }
    
    response := ResponseTradeStatus {
        Id: data.Id,
        Market: data.Market,
        Duration: data.Duration,
        Type: mt.MicrotickTradeNameFromType(data.Type),
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
        Balance: data.Balance,
    }
    
    bz, err2 := codec.MarshalJSONIndent(ModuleCdc, response)
    if err2 != nil {
        panic("Could not marshal result to JSON")
    }
    
    return bz, nil
}
