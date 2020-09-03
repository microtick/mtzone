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

type ResponseTradeStatus struct {
    Id mt.MicrotickId `json:"id"`
    Market mt.MicrotickMarket `json:"market"`
    Duration mt.MicrotickDurationName `json:"duration"`
    Order mt.MicrotickOrderTypeName `json:"order"`
    Taker mt.MicrotickAccount `json:"taker"`
    Legs []ResponseTradeLeg `json:"legs"`
    Start time.Time `json:"start"`
    Expiration time.Time `json:"expiration"`
    Strike mt.MicrotickSpot `json:"strike"`
    CurrentSpot mt.MicrotickSpot `json:"currentSpot"`
    CurrentValue sdk.Dec `json:"currentValue"`
    Commission mt.MicrotickCoin `json:"commission"`
    SettleIncentive mt.MicrotickCoin `json:"settleIncentive"`
}

type ResponseTradeLeg struct {
    LegId mt.MicrotickId `json:"leg_id"`
    Type mt.MicrotickLegTypeName `json:"type"`
    Backing mt.MicrotickCoin `json:"backing"`
    Premium mt.MicrotickPremium `json:"premium"`
    Cost mt.MicrotickCoin `json:"cost"`
    Quantity mt.MicrotickQuantity `json:"quantity"`
    Long mt.MicrotickAccount `json:"long"`
    Short mt.MicrotickAccount `json:"short"`
    Quoted keeper.DataQuotedParams `json:"quoted"`
    CurrentValue sdk.Dec `json:"currentValue"`
}

func (rts ResponseTradeStatus) String() string {
    legStrings := make([]string, len(rts.Legs))
    for i := 0; i < len(rts.Legs); i++ {
        legStrings[i] = formatTradeLeg(rts.Legs[i])
    }
    return strings.TrimSpace(fmt.Sprintf(`Trade Id: %d
Market: %s
Duration: %s
Order: %s
Start: %s
Expiration: %s
Commission: %s
Settle Incentive: %s
Taker: %s
Legs: %s
Strike: %s 
Current Spot: %s
Current Value (Taker): %sdai`,
    rts.Id, 
    rts.Market, 
    rts.Duration,
    rts.Order,
    rts.Start.String(),
    rts.Expiration.String(),
    rts.Commission.String(),
    rts.SettleIncentive.String(),
    rts.Taker.String(),
    legStrings,
    rts.Strike.String(),
    rts.CurrentSpot.String(),
    rts.CurrentValue.String()))
}

func formatTradeLeg(leg ResponseTradeLeg) string {
    return fmt.Sprintf(`
    Leg: %d
        Type: %s
        Long: %s
        Short: %s
        Quantity: %s
        Backing: %s
        Premium: %s
        Cost: %s
        Quoted: %s
        CurrentValue: %s`,
        leg.LegId,
        leg.Type,
        leg.Long.String(),
        leg.Short.String(),
        leg.Quantity.String(),
        leg.Backing.String(),
        leg.Premium.String(),
        leg.Cost.String(),
        formatQuoteParams(leg.Quoted),
        leg.CurrentValue.String(),
    )
}

func formatQuoteParams(params keeper.DataQuotedParams) string {
    var final string
    if params.Final {
        final = "true"
    } else {
        final = "false"
    }
    return fmt.Sprintf(`
            Id: %d 
            Final: %s
            Premium: %s 
            Spot: %s`,
        params.Id,
        final,
        params.Premium.String(),
        params.Spot.String(),
    )
}

func QueryTradeStatus(ctx sdk.Context, path []string, req abci.RequestQuery, keeper keeper.Keeper) (res []byte, err error) {
    id, err := strconv.Atoi(path[0])
    if err != nil {
        return nil, sdkerrors.Wrapf(mt.ErrInvalidTrade, "%d", id)
    }
    data, err := keeper.GetActiveTrade(ctx, mt.MicrotickId(id))
    if err != nil {
        return nil, sdkerrors.Wrapf(mt.ErrInvalidTrade, "fetching %d", id)
    }
    dataMarket, err := keeper.GetDataMarket(ctx, data.Market)
    if err != nil {
        return nil, sdkerrors.Wrap(mt.ErrInvalidMarket, data.Market)
    }
    
    var legs []ResponseTradeLeg
    for _, leg := range data.Legs {
        legs = append(legs, ResponseTradeLeg{
            LegId: leg.LegId,
            Type: mt.MicrotickLegNameFromType(leg.Type),
            Backing: leg.Backing,
            Premium: leg.Premium,
            Cost: leg.Cost,
            Quantity: leg.Quantity,
            Long: leg.Long,
            Short: leg.Short,
            Quoted: leg.Quoted,
            CurrentValue: leg.CalculateValue(dataMarket.Consensus.Amount, data.Strike.Amount),
        })
    }
    
    response := ResponseTradeStatus {
        Id: data.Id,
        Market: data.Market,
        Duration: data.DurationName,
        Order: mt.MicrotickOrderNameFromType(data.Order),
        Taker: data.Taker,
        Legs: legs,
        Start: data.Start,
        Expiration: data.Expiration,
        Strike: data.Strike,
        CurrentSpot: dataMarket.Consensus,
        CurrentValue: data.CurrentValue(data.Taker, dataMarket.Consensus),
        Commission: data.Commission,
        SettleIncentive: data.SettleIncentive,
    }
    
    bz, err := codec.MarshalJSONIndent(keeper.Cdc, response)
    if err != nil {
        panic("Could not marshal result to JSON")
    }
    
    return bz, nil
}
