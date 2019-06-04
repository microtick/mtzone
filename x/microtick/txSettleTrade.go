package microtick

import (
    "fmt"
    "time"
    
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
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
    Id MicrotickId
    Requester sdk.AccAddress
}

func NewTxSettleTrade(id MicrotickId, requester sdk.AccAddress) TxSettleTrade {
    return TxSettleTrade {
        Id: id,
        Requester: requester,
    }
}

type SettlementData struct {
    Short MicrotickAccount `json:"short"`
    Settle MicrotickCoin `json:"settle"`
    Refund MicrotickCoin `json:"refund"`
    Balance MicrotickCoin `json:"balance"`
}

type TradeSettlementData struct {
    Originator string `json:"originator"`
    Id MicrotickId `json:"id"`
    Time time.Time `json:"time"`
    Final MicrotickSpot `json:"final"`
    Long MicrotickAccount `json:"long"`
    Settle MicrotickCoin `json:"settle"`
    CounterParties []SettlementData `json:"counterparties"`
    Incentive MicrotickCoin `json:"incentive"`
    Commission MicrotickCoin `json:"commission"`
    Balance MicrotickCoin `json:"balance"`
    Settler MicrotickAccount `json:"settler"`
    SettlerBalance MicrotickCoin `json:"settlerBalance"`
}

func (msg TxSettleTrade) Route() string { return "microtick" }

func (msg TxSettleTrade) Type() string { return "settle-trade" }

func (msg TxSettleTrade) ValidateBasic() sdk.Error {
    if msg.Requester.Empty() {
        return sdk.ErrInvalidAddress(msg.Requester.String())
    }
    return nil
}

func (msg TxSettleTrade) GetSignBytes() []byte {
    return sdk.MustSortJSON(msgCdc.MustMarshalJSON(msg))
}

func (msg TxSettleTrade) GetSigners() []sdk.AccAddress {
    return []sdk.AccAddress{msg.Requester}
}

// Handler

func handleTxSettleTrade(ctx sdk.Context, keeper Keeper, msg TxSettleTrade) sdk.Result {
    params := keeper.GetParams(ctx)
    
    trade, err := keeper.GetActiveTrade(ctx, msg.Id)
    if err != nil {
        return sdk.ErrInternal("Invalid trade ID").Result()
    }
    
    var settleData []SettlementData
    totalPaid := sdk.NewDec(0)
    
    now := ctx.BlockHeader().Time
        
    // check if trade has expired
    if now.Before(trade.Expiration) {
        return sdk.ErrInternal("Trade cannot be settled until expiration").Result()
    }
        
    dataMarket, err2 := keeper.GetDataMarket(ctx, trade.Market)
    if err2 != nil {
        return sdk.ErrInternal("Could not fetch market consensus").Result()
    }
    
    settlements := trade.CounterPartySettlements(dataMarket.Consensus)
    
    // Incentive 
    keeper.DepositMicrotickCoin(ctx, msg.Requester, trade.SettleIncentive)
    fmt.Printf("Settle Incentive: %s\n", trade.SettleIncentive.String())
    
    // Commission
    commission := NewMicrotickCoinFromDec(params.CommissionSettleFixed)
    keeper.WithdrawMicrotickCoin(ctx, msg.Requester, commission)
    fmt.Printf("Settle Commission: %s\n", commission.String())
    keeper.PoolCommission(ctx, commission)
    
    if params.EuropeanOptions {
        
        // Payout and refunds
        for i := 0; i < len(settlements); i++ {
            pair := settlements[i]
            
            // Long
            keeper.DepositMicrotickCoin(ctx, trade.Long, pair.Settle)
            totalPaid = totalPaid.Add(pair.Settle.Amount)
            
            // Refund
            keeper.DepositMicrotickCoin(ctx, pair.RefundAddress, pair.Refund)
            
            // Adjust trade backing
            accountStatus := keeper.GetAccountStatus(ctx, pair.RefundAddress)
            accountStatus.ActiveTrades.Delete(trade.Id)
            accountStatus.TradeBacking = accountStatus.TradeBacking.Sub(pair.Backing)
            keeper.SetAccountStatus(ctx, pair.RefundAddress, accountStatus)
            
            settleData = append(settleData, SettlementData {
                Short: pair.RefundAddress,
                Settle: pair.Settle,
                Refund: pair.Refund,
                Balance: keeper.GetTotalBalance(ctx, pair.RefundAddress),
            })
        }
        
        accountStatusLong := keeper.GetAccountStatus(ctx, trade.Long)
        accountStatusLong.ActiveTrades.Delete(trade.Id)
        keeper.SetAccountStatus(ctx, trade.Long, accountStatusLong)
        keeper.DeleteActiveTrade(ctx, trade.Id)
        
    } else {
        
        // American options not implemented yet
        return sdk.ErrInternal("American style option settlement not implemented yet").Result()
        
    }
    
    tags := sdk.NewTags(
        fmt.Sprintf("trade.%d", trade.Id), "settle",
        fmt.Sprintf("acct.%s", trade.Long), "settle.long",
    )
    
    var found bool = false
    if trade.Long.Equals(msg.Requester) {
        found = true
    }
    
    for i := 0; i < len(trade.CounterParties); i++ {
        cp := trade.CounterParties[i]
        
        tags = tags.AppendTag(fmt.Sprintf("acct.%s", cp.Short), "settle.short")
        if cp.Short.Equals(msg.Requester) {
            found = true
        }
    }
    
    if !found {
        // msg.Requester is not a long or short account for this trade
        tags = tags.AppendTag(fmt.Sprintf("acct.%s", msg.Requester), "settle.finalize")
    }
    
    data := TradeSettlementData {
        Originator: "settleTrade",
        Id: trade.Id,
        Time: now,
        Final: dataMarket.Consensus,
        Long: trade.Long,
        Settle: NewMicrotickCoinFromDec(totalPaid),
        CounterParties: settleData,
        Incentive: trade.SettleIncentive,
        Commission: commission,
        Balance: keeper.GetTotalBalance(ctx, trade.Long),
        Settler: msg.Requester,
        SettlerBalance: keeper.GetTotalBalance(ctx, msg.Requester),
    }
    bz, _ := codec.MarshalJSONIndent(keeper.cdc, data)
    
	return sdk.Result {
	    Data: bz,
	    Tags: tags,
	}
}
