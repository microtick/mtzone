package msg 

import (
    "context"
    
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    
    mt "github.com/mjackson001/mtzone/x/microtick/types"
)

func (querier Querier) Trade(c context.Context, req *QueryTradeRequest) (*QueryTradeResponse, error) {
    ctx := sdk.UnwrapSDKContext(c)
    id := req.Id
    data, err := querier.Keeper.GetActiveTrade(ctx, mt.MicrotickId(id))
    if err != nil {
        return nil, sdkerrors.Wrapf(mt.ErrInvalidTrade, "fetching %d", id)
    }
    dataMarket, err := querier.Keeper.GetDataMarket(ctx, data.Market)
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
                Spot: leg.Quoted.Spot,
            },
            CurrentValue: leg.CalculateValue(dataMarket.Consensus.Amount, data.Strike.Amount),
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
        CurrentValue: data.CurrentValue(data.Taker, dataMarket.Consensus),
    }
    
    return &response, nil
}
