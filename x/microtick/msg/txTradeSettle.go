package msg

import (
    "fmt"
    "time"
    
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    
    mt "github.com/mjackson001/mtzone/x/microtick/types"
    "github.com/mjackson001/mtzone/x/microtick/keeper"
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

type TxSettleTrade struct {
    Id mt.MicrotickId
    Requester sdk.AccAddress
}

func NewTxSettleTrade(id mt.MicrotickId, requester sdk.AccAddress) TxSettleTrade {
    return TxSettleTrade {
        Id: id,
        Requester: requester,
    }
}

type SettlementData struct {
    LegId mt.MicrotickId `json:"leg_id"`
    SettleAccount mt.MicrotickAccount `json:"settle_account"`
    Settle mt.MicrotickCoin `json:"settle"`
    RefundAccount mt.MicrotickAccount `json:"refund_account"`
    Refund mt.MicrotickCoin `json:"refund"`
}

type TradeSettlementData struct {
    Id mt.MicrotickId `json:"id"`
    Time time.Time `json:"time"`
    Final mt.MicrotickSpot `json:"final"`
    Settlements []SettlementData `json:"settlements"`
    Incentive mt.MicrotickCoin `json:"incentive"`
    Commission mt.MicrotickCoin `json:"commission"`
    Settler mt.MicrotickAccount `json:"settler"`
}

func (msg TxSettleTrade) Route() string { return "microtick" }

func (msg TxSettleTrade) Type() string { return "trade_settle" }

func (msg TxSettleTrade) ValidateBasic() error {
    if msg.Requester.Empty() {
        return sdkerrors.Wrap(mt.ErrInvalidAddress, msg.Requester.String())
    }
    return nil
}

func (msg TxSettleTrade) GetSignBytes() []byte {
    return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg TxSettleTrade) GetSigners() []sdk.AccAddress {
    return []sdk.AccAddress{msg.Requester}
}

// Handler

func HandleTxSettleTrade(ctx sdk.Context, keeper keeper.Keeper, params mt.Params,
    msg TxSettleTrade) (*sdk.Result, error) {
    
    trade, err := keeper.GetActiveTrade(ctx, msg.Id)
    if err != nil {
        return nil, sdkerrors.Wrapf(mt.ErrInvalidTrade, "%d", msg.Id)
    }
    
    var settleData []SettlementData
    
    // check if trade has expired
    now := ctx.BlockHeader().Time
    if now.Before(trade.Expiration) {
        return nil, sdkerrors.Wrap(mt.ErrTradeSettlement, "trade not expired")
    }
        
    dataMarket, err := keeper.GetDataMarket(ctx, trade.Market)
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
    err = keeper.DepositMicrotickCoin(ctx, msg.Requester, trade.SettleIncentive)
    if err != nil {
        return nil, sdkerrors.Wrap(mt.ErrTradeSettlement, "settle incentive")
    }
    
    // Commission
    commission := mt.NewMicrotickCoinFromDec(params.CommissionSettleFixed)
    err = keeper.WithdrawMicrotickCoin(ctx, msg.Requester, commission)
    if err != nil {
        return nil, mt.ErrInsufficientFunds
    }
    reward, err := keeper.PoolCommission(ctx, msg.Requester, commission)
    if err != nil {
        return nil, err
    }
    
    if params.EuropeanOptions {
        
        // Payout and refunds
        for _, pair := range settlements {
            
            // Long
            err = keeper.DepositMicrotickCoin(ctx, pair.SettleAccount, pair.Settle)
            if err != nil {
                return nil, sdkerrors.Wrap(mt.ErrTradeSettlement, "payout")
            }
            
            // Refund
            err := keeper.DepositMicrotickCoin(ctx, pair.RefundAccount, pair.Refund)
            if err != nil {
                return nil, sdkerrors.Wrap(mt.ErrTradeSettlement, "refund")
            }
            
            // Adjust trade backing
            accountStatus := keeper.GetAccountStatus(ctx, pair.RefundAccount)
            accountStatus.ActiveTrades.Delete(trade.Id)
            accountStatus.TradeBacking = accountStatus.TradeBacking.Sub(pair.Backing)
            keeper.SetAccountStatus(ctx, pair.RefundAccount, accountStatus)
            
            accountStatus = keeper.GetAccountStatus(ctx, pair.SettleAccount)
            accountStatus.ActiveTrades.Delete(trade.Id)
            keeper.SetAccountStatus(ctx, pair.SettleAccount, accountStatus)
            
            settleData = append(settleData, SettlementData {
                LegId: pair.LegId,
                SettleAccount: pair.SettleAccount,
                Settle: pair.Settle,
                RefundAccount: pair.RefundAccount,
                Refund: pair.Refund,
            })
        }
        
        accountStatusTaker := keeper.GetAccountStatus(ctx, trade.Taker)
        accountStatusTaker.SettleBacking = accountStatusTaker.SettleBacking.Sub(trade.SettleIncentive)
        keeper.SetAccountStatus(ctx, trade.Taker, accountStatusTaker)
        keeper.DeleteActiveTrade(ctx, trade.Id)
        
    } else {
        
        // American options not implemented yet
        return nil, sdkerrors.Wrap(mt.ErrGeneral, "American style option settlement not implemented yet")
        
    }
    
    data := TradeSettlementData {
        Id: trade.Id,
        Time: now,
        Final: dataMarket.Consensus,
        Settlements: settleData,
        Incentive: trade.SettleIncentive,
        Commission: commission,
        Settler: msg.Requester,
    }
    bz, _ := codec.MarshalJSONIndent(ModuleCdc, data)

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
