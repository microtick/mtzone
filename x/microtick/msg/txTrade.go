package msg

import (
    "github.com/gogo/protobuf/proto"
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    
    mt "gitlab.com/microtick/mtzone/x/microtick/types"
    "gitlab.com/microtick/mtzone/x/microtick/keeper"
)

func (msg TxMarketTrade) Route() string { return "microtick" }

func (msg TxMarketTrade) Type() string { return "trade" }

func (msg TxMarketTrade) ValidateBasic() error {
    if msg.OrderType != mt.MicrotickOrderBuyCall &&
        msg.OrderType != mt.MicrotickOrderSellCall &&
        msg.OrderType != mt.MicrotickOrderBuyPut &&
        msg.OrderType != mt.MicrotickOrderSellPut &&
        msg.OrderType != mt.MicrotickOrderBuySyn &&
        msg.OrderType != mt.MicrotickOrderSellSyn {
        return sdkerrors.Wrap(mt.ErrInvalidOrderType, msg.OrderType)
    }
    if msg.Market == "" {
        return sdkerrors.Wrap(mt.ErrMissingParam, "market")
    }
    if msg.Taker.Empty() {
        return sdkerrors.Wrap(mt.ErrInvalidAddress, msg.Taker.String())
    }
    return nil
}

func (msg TxMarketTrade) GetSignBytes() []byte {
    return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg TxMarketTrade) GetSigners() []sdk.AccAddress {
    return []sdk.AccAddress{ msg.Taker }
}

// Handler

func HandleTxMarketTrade(ctx sdk.Context, mtKeeper keeper.Keeper, params mt.MicrotickParams,
    msg TxMarketTrade) (*sdk.Result, error) {
        
    quantity := mt.NewMicrotickQuantityFromString(msg.Quantity)
     
    if !mtKeeper.HasDataMarket(ctx, msg.Market) {
        return nil, sdkerrors.Wrap(mt.ErrInvalidMarket, msg.Market)
    }
    
    if !mtKeeper.ValidDurationName(ctx, msg.Duration) {
        return nil, sdkerrors.Wrapf(mt.ErrInvalidDuration, "%s", msg.Duration)
    }
    
    // Step 1 - Obtain the strike spot price and create trade struct
    market, err := mtKeeper.GetDataMarket(ctx, msg.Market)
    if err != nil {
        return nil, mt.ErrInvalidMarket
    }
    commission := mt.NewMicrotickCoinFromDec(params.CommissionTradeFixed)
    settleIncentive := mt.NewMicrotickCoinFromDec(params.SettleIncentive)
    now := ctx.BlockHeader().Time
    seconds := mtKeeper.DurationFromName(ctx, msg.Duration)
    trade := keeper.NewDataActiveTrade(now, msg.Market, msg.Duration, seconds,
        msg.OrderType, msg.Taker, market.Consensus, commission, settleIncentive)
        
    matcher := keeper.NewMatcher(trade, func (id mt.MicrotickId) keeper.DataActiveQuote {
        quote, err := mtKeeper.GetActiveQuote(ctx, id)
        if err != nil {
            // This function should always be called with an active quote
            panic("Invalid quote ID")
        }
        return quote
    })
        
    // Step 2 - Compute premium for quantity requested
    if msg.OrderType == mt.MicrotickOrderBuyCall || msg.OrderType == mt.MicrotickOrderSellCall ||
        msg.OrderType == mt.MicrotickOrderBuyPut || msg.OrderType == mt.MicrotickOrderSellPut {
            
        err = matcher.MatchByQuantity(mtKeeper, &market, msg.OrderType, quantity)
        if err != nil {
            return nil, err
        }
    } else {
        syntheticBook := mtKeeper.GetSyntheticBook(ctx, &market, msg.Duration, &msg.Taker)
        err = matcher.MatchSynthetic(mtKeeper, &syntheticBook, &market, quantity)
        if err != nil {
            return nil, err
        }
    }
    
    if matcher.HasQuantity {
        
        // Deduct commission and settle incentive
        total := trade.Commission.Add(settleIncentive)
        err = mtKeeper.WithdrawMicrotickCoin(ctx, msg.Taker, total)
        if err != nil {
            return nil, mt.ErrInsufficientFunds
        }
        reward, err := mtKeeper.PoolCommission(ctx, msg.Taker, trade.Commission, true, sdk.OneDec())
        if err != nil {
            return nil, err
        }
    
        // Step 4 - Finalize trade 
        matcher.Trade.Id = mtKeeper.GetNextActiveTradeId(ctx)
        
        err = matcher.AssignCounterparties(ctx, mtKeeper, &market)
        if err != nil {
            return nil, err
        }
        
        accountStatus := mtKeeper.GetAccountStatus(ctx, msg.Taker)
        accountStatus.SettleBacking = accountStatus.SettleBacking.Add(settleIncentive)
        accountStatus.PlacedTrades++
        mtKeeper.SetAccountStatus(ctx, msg.Taker, accountStatus)
        
        // Save
        mtKeeper.CommitTradeId(ctx, matcher.Trade.Id)
        mtKeeper.SetDataMarket(ctx, market)
        mtKeeper.SetActiveTrade(ctx, matcher.Trade)
        
        // Data
        data := MarketTradeData {
            Consensus: market.Consensus,
            Time: now.Unix(),
            Trade: matcher.Trade,
            Commission: commission,
            Reward: *reward,
        }
        bz, err := proto.Marshal(&data)
        
        var events []sdk.Event
        events = append(events, sdk.NewEvent(
            sdk.EventTypeMessage,
            sdk.NewAttribute(sdk.AttributeKeyModule, mt.ModuleKey),
        ))
        
        ctx.EventManager().EmitEvents(events)
            
        return &sdk.Result {
            Data: bz,
            Events: ctx.EventManager().ABCIEvents(),
        }, nil
        
    }
    
    // No liquidity available
    return nil, sdkerrors.Wrap(mt.ErrTradeMatch, "no liquidity available")
}
