package microtick

import (
    "encoding/json"
    "time"
    
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

func (msg TxSettleTrade) Route() string { return "microtick" }

func (msg TxSettleTrade) Type() string { return "settle-trade" }

func (msg TxSettleTrade) ValidateBasic() sdk.Error {
    if msg.Requester.Empty() {
        return sdk.ErrInvalidAddress(msg.Requester.String())
    }
    return nil
}

func (msg TxSettleTrade) GetSignBytes() []byte {
    b, err := json.Marshal(msg)
    if err != nil {
        panic(err)
    }
    return sdk.MustSortJSON(b)
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
    
    if params.EuropeanOptions {
        now := time.Now()
        
        // check if trade has expired
        if now.Before(trade.Expiration) {
            return sdk.ErrInternal("Trade cannot be settled until expiration").Result()
        }
        
        dataMarket, err2 := keeper.GetDataMarket(ctx, trade.Market)
        if err2 != nil {
            return sdk.ErrInternal("Could not fetch market consensus").Result()
        }
        
        settlements := trade.CounterPartySettlements(dataMarket.Consensus)
        
        // Payout and refunds
        for i := 0; i < len(settlements); i++ {
            pair := settlements[i]
            keeper.DepositMicrotickCoin(ctx, trade.Long, pair.Settle)
            keeper.DepositMicrotickCoin(ctx, pair.RefundAddress, pair.Refund)
            
            // Adjust trade backing
            accountStatus := keeper.GetAccountStatus(ctx, pair.RefundAddress)
            accountStatus.ActiveTrades.Delete(trade.Id)
            accountStatus.TradeBacking = accountStatus.TradeBacking.Minus(pair.Backing)
            keeper.SetAccountStatus(ctx, pair.RefundAddress, accountStatus)
        }
        
        accountStatusLong := keeper.GetAccountStatus(ctx, trade.Long)
        accountStatusLong.ActiveTrades.Delete(trade.Id)
        keeper.SetAccountStatus(ctx, trade.Long, accountStatusLong)
        keeper.DeleteActiveTrade(ctx, trade.Id)
        
    } else {
        
        // American options not implemented yet
        return sdk.ErrInternal("American style option settlement not implemented yet").Result()
        
    }
    
	return sdk.Result {}
}
