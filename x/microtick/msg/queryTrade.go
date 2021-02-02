package msg 

import (
    "context"
    
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    
    "gitlab.com/microtick/mtzone/x/microtick/keeper"
    mt "gitlab.com/microtick/mtzone/x/microtick/types"
)

func (querier Querier) Trade(c context.Context, req *QueryTradeRequest) (*QueryTradeResponse, error) {
    ctx := sdk.UnwrapSDKContext(c)
    return baseQueryTrade(ctx, querier.Keeper, req)
}

func baseQueryTrade(ctx sdk.Context, keeper keeper.Keeper, req *QueryTradeRequest) (*QueryTradeResponse, error) {
    id := req.Id
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
            Quoted: ResponseQuotedParams {
                Id: leg.Quoted.Id,
                Premium: leg.Quoted.Premium,
                RemainBacking: leg.Quoted.RemainBacking.String(),
                Spot: leg.Quoted.Spot,
            },
            CurrentValue: leg.CalculateValue(dataMarket.Consensus.Amount, data.Strike.Amount).String(),
        })
    }
    
    response := QueryTradeResponse {
        Id: data.Id,
        Market: data.Market,
        Duration: data.Duration,
        Order: mt.MicrotickOrderNameFromType(data.Order),
        Taker: data.Taker,
        Quantity: data.Quantity,
        Legs: legs,
        Start: data.Start,
        Expiration: data.Expiration,
        Strike: data.Strike,
        Commission: data.Commission,
        SettleIncentive: data.SettleIncentive,
        Consensus: dataMarket.Consensus,
        CurrentValue: data.CurrentValue(data.Taker, dataMarket.Consensus).String(),
    }
    
    return &response, nil
}
