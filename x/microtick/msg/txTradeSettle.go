package msg

import (
    "fmt"
    "time"
    "errors"
    
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
    
    mt "github.com/mjackson001/mtzone/x/microtick/types"
    "github.com/mjackson001/mtzone/x/microtick/keeper"
)

// The rules for settling a trade are as follows:
// If European mode is set, (mode TBD)
//     If expiration time is now or in the past
//         Anyone can call settle-trade, value is paid out and difference refunded
// If American mode is set, (TBD)
//     If expiration time is in the futuer
//         Only the buyer can settle trade (must be signed and verified)
//     Else
//         Anyone can call settle trade

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
    Short mt.MicrotickAccount `json:"short"`
    Settle mt.MicrotickCoin `json:"settle"`
    Refund mt.MicrotickCoin `json:"refund"`
}

type TradeSettlementData struct {
    Id mt.MicrotickId `json:"id"`
    Time time.Time `json:"time"`
    Final mt.MicrotickSpot `json:"final"`
    Long mt.MicrotickAccount `json:"long"`
    Settle mt.MicrotickCoin `json:"settle"`
    CounterParties []SettlementData `json:"counterparties"`
    Incentive mt.MicrotickCoin `json:"incentive"`
    Commission mt.MicrotickCoin `json:"commission"`
    Settler mt.MicrotickAccount `json:"settler"`
}

func (msg TxSettleTrade) Route() string { return "microtick" }

func (msg TxSettleTrade) Type() string { return "trade_settle" }

func (msg TxSettleTrade) ValidateBasic() error {
    if msg.Requester.Empty() {
        return errors.New(fmt.Sprintf("Invalid address: %s", msg.Requester.String()))
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

func HandleTxSettleTrade(ctx sdk.Context, keeper keeper.Keeper, msg TxSettleTrade) (*sdk.Result, error) {
    params := keeper.GetParams(ctx)
    
    trade, err := keeper.GetActiveTrade(ctx, msg.Id)
    if err != nil {
        return nil, errors.New("Invalid trade ID")
    }
    
    var settleData []SettlementData
    totalPaid := sdk.NewDec(0)
    
    now := ctx.BlockHeader().Time
        
    // check if trade has expired
    if now.Before(trade.Expiration) {
        return nil, errors.New("Trade cannot be settled until expiration")
    }
        
    dataMarket, err2 := keeper.GetDataMarket(ctx, trade.Market)
    if err2 != nil {
        return nil, errors.New("Could not fetch market consensus")
    }
    
    settlements := trade.CounterPartySettlements(dataMarket.Consensus)
    
    // Incentive 
    err2 = keeper.DepositMicrotickCoin(ctx, msg.Requester, trade.SettleIncentive)
    if err2 != nil {
        return nil, errors.New("Fund mismatch (incentive)")
    }
    //fmt.Printf("Settle Incentive: %s\n", trade.SettleIncentive.String())
    
    // Commission
    commission := mt.NewMicrotickCoinFromDec(params.CommissionSettleFixed)
     
    err = keeper.WithdrawMicrotickCoin(ctx, msg.Requester, commission)
    if err != nil {
        return nil, errors.New("Insufficient funds")
    }
    
    //fmt.Printf("Settle Commission: %s\n", commission.String())
    keeper.PoolCommission(ctx, msg.Requester, commission)
    
    if params.EuropeanOptions {
        
        // Payout and refunds
        for i := 0; i < len(settlements); i++ {
            pair := settlements[i]
            
            // Long
            err2 = keeper.DepositMicrotickCoin(ctx, trade.Long, pair.Settle)
            if err2 != nil {
                return nil, errors.New("Fund mismatch (long)")
            }
            totalPaid = totalPaid.Add(pair.Settle.Amount)
            
            // Refund
            err2 := keeper.DepositMicrotickCoin(ctx, pair.RefundAddress, pair.Refund)
            if err2 != nil {
                return nil, errors.New("Fund mismatch (refund)")
            }
            
            // Adjust trade backing
            accountStatus := keeper.GetAccountStatus(ctx, pair.RefundAddress)
            accountStatus.ActiveTrades.Delete(trade.Id)
            accountStatus.TradeBacking = accountStatus.TradeBacking.Sub(pair.Backing)
            keeper.SetAccountStatus(ctx, pair.RefundAddress, accountStatus)
            
            settleData = append(settleData, SettlementData {
                Short: pair.RefundAddress,
                Settle: pair.Settle,
                Refund: pair.Refund,
            })
        }
        
        accountStatusLong := keeper.GetAccountStatus(ctx, trade.Long)
        accountStatusLong.ActiveTrades.Delete(trade.Id)
        accountStatusLong.SettleBacking = accountStatusLong.SettleBacking.Sub(trade.SettleIncentive)
        keeper.SetAccountStatus(ctx, trade.Long, accountStatusLong)
        keeper.DeleteActiveTrade(ctx, trade.Id)
        
    } else {
        
        // American options not implemented yet
        return nil, errors.New("American style option settlement not implemented yet")
        
    }
    
    data := TradeSettlementData {
        Id: trade.Id,
        Time: now,
        Final: dataMarket.Consensus,
        Long: trade.Long,
        Settle: mt.NewMicrotickCoinFromDec(totalPaid),
        CounterParties: settleData,
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
        sdk.NewAttribute(fmt.Sprintf("acct.%s", trade.Long), "settle.long"),
    ))
    
    for i := 0; i < len(trade.CounterParties); i++ {
        cp := trade.CounterParties[i]
        
        events = append(events, sdk.NewEvent(
            sdk.EventTypeMessage,
            sdk.NewAttribute(fmt.Sprintf("acct.%s", cp.Short), "settle.short"),
        ))
    }
    
    events = append(events, sdk.NewEvent(
        sdk.EventTypeMessage,
        sdk.NewAttribute(fmt.Sprintf("acct.%s", msg.Requester), "settle.finalize"),
    ))
    
	return &sdk.Result {
	    Data: bz,
	    Events: events,
	}, nil
}
