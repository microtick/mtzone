package keeper

import (
	"fmt"
    sdk "github.com/cosmos/cosmos-sdk/types"
    auth "github.com/cosmos/cosmos-sdk/x/auth/types"
    mt "github.com/mjackson001/mtzone/x/microtick/types"
)

const MTModuleAccount = "microtick"
const MTPoolName = "commissionPool"

// Commissions

func (k Keeper) PoolCommission(ctx sdk.Context, addr sdk.AccAddress, amount mt.MicrotickCoin) {
    extCoins := mt.MicrotickCoinToExtCoin(amount)
    
	store := ctx.KVStore(k.AppGlobalsKey)
	
	// Get current pool amount
	key := []byte(MTPoolName)
	var pool sdk.DecCoin = sdk.NewInt64DecCoin(mt.ExtTokenType, 0)
	if store.Has(key) {
		bz := store.Get(key)
		k.Cdc.MustUnmarshalBinaryBare(bz, &pool)
	}
	pool = pool.Add(sdk.NewDecCoin(mt.ExtTokenType, extCoins.Amount))
	
    //fmt.Printf("Add Pool Commission: requested %s actual %s pool %s\n", amount.String(), extCoins.String(), pool.String())
	
	store.Set(key, k.Cdc.MustMarshalBinaryBare(pool))
}

func (k Keeper) Sweep(ctx sdk.Context) {
	store := ctx.KVStore(k.AppGlobalsKey)
	
	// Get current pool amount
	key := []byte(MTPoolName)
	var pool sdk.DecCoin = sdk.NewInt64DecCoin(mt.ExtTokenType, 0)
	if store.Has(key) {
		bz := store.Get(key)
		k.Cdc.MustUnmarshalBinaryBare(bz, &pool)
	}
	coin, _ := pool.TruncateDecimal()
	
    //fmt.Printf("Sweep: %s %s\n", pool.String(), coin.String())
    k.supplyKeeper.SendCoinsFromModuleToModule(ctx, MTModuleAccount, 
    	auth.FeeCollectorName, sdk.Coins{coin})
    	
    pool = sdk.NewInt64DecCoin(mt.ExtTokenType, 0)
    store.Set(key, k.Cdc.MustMarshalBinaryBare(pool))
    
    ctx.EventManager().EmitEvent(
        sdk.NewEvent(
            sdk.EventTypeMessage,
            sdk.NewAttribute("mtm.Commissions", fmt.Sprintf("%s", coin)),
        ),
    )
}

// Account balances

func (k Keeper) WithdrawMicrotickCoin(ctx sdk.Context, account sdk.AccAddress, 
    withdrawAmount mt.MicrotickCoin) error {
    	
    extCoins := mt.MicrotickCoinToExtCoin(withdrawAmount)
    
    //fmt.Printf("Withdraw account %s: %s\n", account.String(), extCoins.String())
    
	if extCoins.Amount.IsPositive() {
		err := k.supplyKeeper.SendCoinsFromAccountToModule(ctx, account, MTModuleAccount, sdk.Coins{extCoins})
		if err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) DepositMicrotickCoin(ctx sdk.Context, account sdk.AccAddress,
	depositAmount mt.MicrotickCoin) error {
		
	extCoins := mt.MicrotickCoinToExtCoin(depositAmount)	
	
	//fmt.Printf("Requested: %s\n", depositAmount.String())
    //fmt.Printf("Deposit account %s: %s\n", account.String(), extCoins.String())
	
	if extCoins.Amount.IsPositive() {
		err := k.supplyKeeper.SendCoinsFromModuleToAccount(ctx, MTModuleAccount, account, sdk.Coins{extCoins})
		if err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) GetTotalBalance(ctx sdk.Context, addr sdk.AccAddress) mt.MicrotickCoin {
	coins := k.CoinKeeper.GetBalance(ctx, addr, mt.ExtTokenType)
    balance := mt.ExtCoinToMicrotickCoin(coins)
    return balance
}

func (k Keeper) RefundBacking(ctx sdk.Context) {
    k.IterateAccountStatus(ctx, 
        func(acct DataAccountStatus) (stop bool) {
        	k.DepositMicrotickCoin(ctx, acct.Account, acct.QuoteBacking)
        	k.DepositMicrotickCoin(ctx, acct.Account, acct.TradeBacking)
        	k.DepositMicrotickCoin(ctx, acct.Account, acct.SettleBacking)
            return false
        },
    )
}