package msg

import (
    "github.com/gogo/protobuf/proto"
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    
    mt "github.com/microtick/mtzone/x/microtick/types"
    "github.com/microtick/mtzone/x/microtick/keeper"
)

func (msg TxPickTrade) Route() string { return "microtick" }

func (msg TxPickTrade) Type() string { return "pick" }

func (msg TxPickTrade) ValidateBasic() error {
    if msg.OrderType != mt.MicrotickOrderBuyCall &&
        msg.OrderType != mt.MicrotickOrderSellCall &&
        msg.OrderType != mt.MicrotickOrderBuyPut &&
        msg.OrderType != mt.MicrotickOrderSellPut {
        return sdkerrors.Wrap(mt.ErrInvalidOrderType, msg.OrderType)
    }
    if msg.Taker.Empty() {
        return sdkerrors.Wrap(mt.ErrInvalidAddress, msg.Taker.String())
    }
    return nil
}

func (msg TxPickTrade) GetSignBytes() []byte {
    return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg TxPickTrade) GetSigners() []sdk.AccAddress {
    return []sdk.AccAddress{ msg.Taker }
}

// Handler

func HandleTxPickTrade(ctx sdk.Context, mtKeeper keeper.Keeper, params mt.MicrotickParams,
    msg TxPickTrade) (*sdk.Result, error) {
    
    quote, err := mtKeeper.GetActiveQuote(ctx, msg.Id)
    if err != nil {
        return nil, sdkerrors.Wrapf(mt.ErrInvalidQuote, "%d", msg.Id)
    }
    
    // Step 1 - Obtain the strike spot price and create trade struct
    market, err := mtKeeper.GetDataMarket(ctx, quote.Market)
    if err != nil {
        return nil, sdkerrors.Wrap(mt.ErrInvalidMarket, quote.Market)
    }
    
    if quote.Provider.Equals(msg.Taker) {
        return nil, sdkerrors.Wrap(mt.ErrTradeMatch, "already owner")
    }
    
    commission := mtKeeper.PoolCommission(ctx, params.CommissionTradeFixed)
    settleIncentive := mt.NewMicrotickCoinFromDec(params.SettleIncentive)
    now := ctx.BlockHeader().Time
    durName := mtKeeper.NameFromDuration(ctx, quote.Duration)
    trade := keeper.NewDataActiveTrade(now, quote.Market, durName, mtKeeper.DurationFromName(ctx, durName),
        msg.OrderType, msg.Taker, market.Consensus, commission, settleIncentive)
        
    matcher := keeper.NewMatcher(trade, nil)
    
    // Step 2 - Compute premium and cost
    matcher.MatchQuote(mtKeeper, msg.OrderType, quote)
    
    if matcher.HasQuantity {
        
        // Deduct commission and settle incentive
        total := trade.Commission.Add(settleIncentive)
        err = mtKeeper.WithdrawMicrotickCoin(ctx, msg.Taker, total)
        if err != nil {
            return nil, mt.ErrInsufficientFunds
        }
        reward, err := mtKeeper.AwardRebate(ctx, msg.Taker, params.MintRewardTradeFixed)
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
        data := PickTradeData {
            Market: quote.Market,
            Duration: quote.DurationName,
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
