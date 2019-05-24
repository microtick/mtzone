package microtick

import (
	"fmt"
    sdk "github.com/cosmos/cosmos-sdk/types"
)

// Commissions

func (k Keeper) PoolCommission(ctx sdk.Context, amount MicrotickCoin) {
	store := ctx.KVStore(k.storeKeys.AppGlobals)
	key := []byte("commissionPool")
	
	var pool MicrotickCoin = NewMicrotickCoinFromInt(0)
	
	if store.Has(key) {
		bz := store.Get(key)
		k.cdc.MustUnmarshalBinaryBare(bz, &pool)
	}
	
	pool = pool.Add(amount)
	fmt.Printf("Pool: %s\n", pool.String())
	whole, pool := pool.TruncateDecimal()
	
	fmt.Printf("Amount: %s\n", amount.String())
	if whole.IsPositive() {
		fmt.Printf("Adding commission: %s\n", whole.String())
		k.feeKeeper.AddCollectedFees(ctx, sdk.Coins{whole})
	}
	
	store.Set(key, k.cdc.MustMarshalBinaryBare(pool))
}

// Account balances

func (k Keeper) WithdrawMicrotickCoin(ctx sdk.Context, account sdk.AccAddress, 
    withdrawAmount MicrotickCoin) {
	accountStatus := k.GetAccountStatus(ctx, account)
	
    if (accountStatus.Change.IsGTE(withdrawAmount)) {
        // handle without needing from the coin balance
        accountStatus.Change = accountStatus.Change.Sub(withdrawAmount)
    } else {
        neededAmount := withdrawAmount.Sub(accountStatus.Change)
	
	    // Load total coin balance + change into DecCoin
	    var amt sdk.Coin
	    var change sdk.DecCoin
	    amt, change = neededAmount.TruncateDecimal()
	    
	    if change.IsPositive() {
	        amt = amt.Add(sdk.NewInt64Coin(TokenType, 1))
	        accountStatus.Change = sdk.NewDecCoinFromDec(TokenType, sdk.OneDec()).Sub(change)
	    } else {
	        accountStatus.Change = NewMicrotickCoinFromInt(0)
	    }
	    
	    _, _, err := k.coinKeeper.SubtractCoins(ctx, account, sdk.Coins{amt})
	    if err != nil {
	        panic("Not enough funds")
	    }
	
	    k.SetAccountStatus(ctx, account, accountStatus)
    }
}

func (k Keeper) DepositMicrotickCoin(ctx sdk.Context, account sdk.AccAddress,
	depositAmount MicrotickCoin) {
	accountStatus := k.GetAccountStatus(ctx, account)
	
	totalDecCoin := accountStatus.Change.Add(depositAmount)
	
	var amt sdk.Coin
	var change sdk.DecCoin
	amt, change = totalDecCoin.TruncateDecimal()
	
	if amt.IsPositive() {
		_, _, err := k.coinKeeper.AddCoins(ctx, account, sdk.Coins{amt})
		if err != nil {
			panic("Deposit failed")
		}
	}
	
	accountStatus.Change = change
	k.SetAccountStatus(ctx, account, accountStatus)
}
