package keeper

import (
	"fmt"
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/x/auth/types"
    
    mt "github.com/mjackson001/mtzone/x/microtick/types"
)

var MTModuleAccount = "microtick"

// Commissions

func (k Keeper) PoolCommission(ctx sdk.Context, addr sdk.AccAddress, amount mt.MicrotickCoin) {
	store := ctx.KVStore(k.AppGlobalsKey)
	key := []byte("commissionPool")
	
	var pool mt.MicrotickCoin = mt.NewMicrotickCoinFromInt(0)
	
	if store.Has(key) {
		bz := store.Get(key)
		k.cdc.MustUnmarshalBinaryBare(bz, &pool)
	}
	
	pool = pool.Add(amount)
	fmt.Printf("Pool: %s\n", pool.String())
	whole, pool := pool.TruncateDecimal()
	
	//fmt.Printf("Amount: %s\n", amount.String())
	if whole.IsPositive() {
		fmt.Printf("Adding commission: %s\n", whole.String())
		k.supplyKeeper.SendCoinsFromAccountToModule(ctx, addr, 
			types.FeeCollectorName, sdk.Coins{whole})
		//k.feeKeeper.AddCollectedFees(ctx, sdk.Coins{whole})
	}
	
	store.Set(key, k.cdc.MustMarshalBinaryBare(pool))
}

func (k Keeper) FractionalCommission(ctx sdk.Context) mt.MicrotickCoin {
	store := ctx.KVStore(k.AppGlobalsKey)
	key := []byte("commissionPool")
	
	var pool mt.MicrotickCoin = mt.NewMicrotickCoinFromInt(0)
	
	if store.Has(key) {
		bz := store.Get(key)
		k.cdc.MustUnmarshalBinaryBare(bz, &pool)
	}
	
	return pool
}

// Account balances

func (k Keeper) WithdrawMicrotickCoin(ctx sdk.Context, account sdk.AccAddress, 
    withdrawAmount mt.MicrotickCoin) {
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
	        amt = amt.Add(sdk.NewInt64Coin(mt.TokenType, 1))
	        accountStatus.Change = sdk.NewDecCoinFromDec(mt.TokenType, sdk.OneDec()).Sub(change)
	    } else {
	        accountStatus.Change = mt.NewMicrotickCoinFromInt(0)
	    }
	    
	    err := k.supplyKeeper.SendCoinsFromAccountToModule(ctx,
	        account, MTModuleAccount, sdk.Coins{amt})
	    if err != nil {
	        panic(err)
	    }
	    
    }
	k.SetAccountStatus(ctx, account, accountStatus)
}

func (k Keeper) DepositMicrotickCoin(ctx sdk.Context, account sdk.AccAddress,
	depositAmount mt.MicrotickCoin) {
	accountStatus := k.GetAccountStatus(ctx, account)
	
	totalDecCoin := accountStatus.Change.Add(depositAmount)
	
	var amt sdk.Coin
	var change sdk.DecCoin
	amt, change = totalDecCoin.TruncateDecimal()
	
	if amt.IsPositive() {
		err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx,
		    MTModuleAccount, account, sdk.Coins{amt})
		if err != nil {
			panic("Deposit failed")
		}
	}
	
	accountStatus.Change = change
	k.SetAccountStatus(ctx, account, accountStatus)
}

func (k Keeper) GetTotalBalance(ctx sdk.Context, addr sdk.AccAddress) mt.MicrotickCoin {
	status := k.GetAccountStatus(ctx, addr)
	coins := k.CoinKeeper.GetCoins(ctx, addr)
    balance := status.Change
    for i := 0; i < len(coins); i++ {
        if coins[i].Denom == mt.TokenType {
            balance = balance.Add(mt.NewMicrotickCoinFromInt(coins[i].Amount.Int64()))
        }
    }
    return balance
}

func (k Keeper) RefundBacking(ctx sdk.Context) {
    k.IterateAccountStatus(ctx, 
        func(acct DataAccountStatus) (stop bool) {
        	k.DepositMicrotickCoin(ctx, acct.Account, acct.QuoteBacking)
        	k.DepositMicrotickCoin(ctx, acct.Account, acct.TradeBacking)
            return false
        },
    )
}
