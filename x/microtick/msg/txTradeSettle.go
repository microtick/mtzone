package msg

import (
    "fmt"
    "time"
    
    "github.com/gogo/protobuf/proto"
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    
    mt "gitlab.com/microtick/mtzone/x/microtick/types"
    "gitlab.com/microtick/mtzone/x/microtick/keeper"
)

// The rules for settling a trade are as follows:
// If European mode is set, (mode TBD)
//     If expiration time is now or in the past
//         Anyone can settle-trade, value is paid out and difference refunded
// If American mode is set, (TBD)
//     If expiration time is in the future
//         Only the buyer can settle trade (must be signed and verified)
//     Else
//         Anyone can settle trade

func (msg TxSettleTrade) Route() string { return "microtick" }

func (msg TxSettleTrade) Type() string { return "trade_settle" }

func (msg TxSettleTrade) ValidateBasic() error {
    if msg.Requester.Empty() {
        return sdkerrors.Wrap(mt.ErrInvalidAddress, msg.Requester.String())
    }
    return nil
}

func (msg TxSettleTrade) GetSignBytes() []byte {
    return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg TxSettleTrade) GetSigners() []sdk.AccAddress {
    return []sdk.AccAddress{msg.Requester}
}

// Handler

func HandleTxSettleTrade(ctx sdk.Context, mtKeeper keeper.Keeper, params mt.MicrotickParams,
    msg TxSettleTrade) (*sdk.Result, error) {
    
    trade, err := mtKeeper.GetActiveTrade(ctx, msg.Id)
    if err != nil {
        return nil, sdkerrors.Wrapf(mt.ErrInvalidTrade, "%d", msg.Id)
    }
    
    var settleData []SettlementData
    
    // check if trade has expired
    now := ctx.BlockHeader().Time
    if now.Before(time.Unix(trade.Expiration, 0)) {
        return nil, sdkerrors.Wrap(mt.ErrTradeSettlement, "trade not expired")
    }
        
    dataMarket, err := mtKeeper.GetDataMarket(ctx, trade.Market)
    if err != nil {
        return nil, sdkerrors.Wrap(mt.ErrInvalidMarket, trade.Market)
    }
    
    // ensure we have at least one quote past its "canModify" time - this check is
    // to prevent manipulation by requiring quotes to have aged before settling a
    // trade.
    if !dataMarket.CanSettle(now) {
        return nil, sdkerrors.Wrap(mt.ErrTradeSettlement, "unconfirmed consensus")
    }
    
    settlements := trade.CalculateLegSettlements(dataMarket.Consensus)
    
    // Reward settle incentive 
    err = mtKeeper.DepositMicrotickCoin(ctx, msg.Requester, trade.SettleIncentive)
    if err != nil {
        return nil, sdkerrors.Wrap(mt.ErrTradeSettlement, "settle incentive")
    }
    
    // Commission
    commission := mt.NewMicrotickCoinFromDec(params.CommissionSettleFixed)
    err = mtKeeper.WithdrawMicrotickCoin(ctx, msg.Requester, commission)
    if err != nil {
        return nil, mt.ErrInsufficientFunds
    }
    reward, err := mtKeeper.PoolCommission(ctx, msg.Requester, commission, true, sdk.OneDec())
    if err != nil {
        return nil, err
    }
    
    if params.EuropeanOptions {
        
        // Payout and refunds
        for _, pair := range settlements {
            
            // Long
            err = mtKeeper.DepositMicrotickCoin(ctx, pair.SettleAccount, pair.Settle)
            if err != nil {
                return nil, sdkerrors.Wrap(mt.ErrTradeSettlement, "payout")
            }
            
            // Refund
            err := mtKeeper.DepositMicrotickCoin(ctx, pair.RefundAccount, pair.Refund)
            if err != nil {
                return nil, sdkerrors.Wrap(mt.ErrTradeSettlement, "refund")
            }
            
            // Adjust trade backing
            accountStatus := mtKeeper.GetAccountStatus(ctx, pair.RefundAccount)
            accountStatus.ActiveTrades.Delete(trade.Id)
            accountStatus.TradeBacking = accountStatus.TradeBacking.Sub(pair.Backing)
            mtKeeper.SetAccountStatus(ctx, pair.RefundAccount, accountStatus)
            
            accountStatus = mtKeeper.GetAccountStatus(ctx, pair.SettleAccount)
            accountStatus.ActiveTrades.Delete(trade.Id)
            mtKeeper.SetAccountStatus(ctx, pair.SettleAccount, accountStatus)
            
            settleData = append(settleData, SettlementData {
                LegId: pair.LegId,
                SettleAccount: pair.SettleAccount,
                Settle: pair.Settle,
                RefundAccount: pair.RefundAccount,
                Refund: pair.Refund,
            })
        }
        
        accountStatusTaker := mtKeeper.GetAccountStatus(ctx, trade.Taker)
        accountStatusTaker.SettleBacking = accountStatusTaker.SettleBacking.Sub(trade.SettleIncentive)
        mtKeeper.SetAccountStatus(ctx, trade.Taker, accountStatusTaker)
        mtKeeper.DeleteActiveTrade(ctx, trade.Id)
        
    } else {
        
        // American options not implemented yet
        return nil, sdkerrors.Wrap(mt.ErrGeneral, "American style option settlement not implemented yet")
        
    }
    
    data := TradeSettlementData {
        Id: trade.Id,
        Time: now.Unix(),
        Final: dataMarket.Consensus,
        Settlements: settleData,
        Incentive: trade.SettleIncentive,
        Commission: commission,
        Settler: msg.Requester,
    }
    bz, err := proto.Marshal(&data)

    var events []sdk.Event
    events = append(events, sdk.NewEvent(
        sdk.EventTypeMessage,
        sdk.NewAttribute(sdk.AttributeKeyModule, mt.ModuleKey),
    ), sdk.NewEvent(
        sdk.EventTypeMessage,
        sdk.NewAttribute(fmt.Sprintf("trade.%d", trade.Id), "event.settle"),
    ), sdk.NewEvent(
        sdk.EventTypeMessage,
        sdk.NewAttribute("commission", commission.String()),
        sdk.NewAttribute("reward", reward.String()),
    ))
    
    for _, leg := range trade.Legs {
        events = append(events, sdk.NewEvent(
            sdk.EventTypeMessage,
            sdk.NewAttribute(fmt.Sprintf("acct.%s", leg.Long), "trade.end"),
            sdk.NewAttribute(fmt.Sprintf("acct.%s", leg.Short), "trade.end"),
        ))
    }
    
    events = append(events, sdk.NewEvent(
        sdk.EventTypeMessage,
        sdk.NewAttribute(fmt.Sprintf("acct.%s", msg.Requester), "settle.finalize"),
    ))
    
    ctx.EventManager().EmitEvents(events)
    
	return &sdk.Result {
	    Data: bz,
	    Events: ctx.EventManager().ABCIEvents(),
	}, nil
}
