package keeper

import (
	"fmt"
	"os"
    sdk "github.com/cosmos/cosmos-sdk/types"
    auth "github.com/cosmos/cosmos-sdk/x/auth/types"
    mt "github.com/mjackson001/mtzone/x/microtick/types"
)

const MTModuleAccount = "microtick"
const MTPoolName = "commissionPool"

type CommissionPool struct {
	Pool sdk.DecCoin `json:"pool"`
}

// Commissions

func (k Keeper) PoolCommission(ctx sdk.Context, addr sdk.AccAddress, amount mt.MicrotickCoin, doRebate bool) (*sdk.Coin, error) {
	params := k.GetParams(ctx)
    extCoins := mt.MicrotickCoinToExtCoin(amount)
    
	store := ctx.KVStore(k.AppGlobalsKey)
	
	// Get current pool amount
	key := []byte(MTPoolName)
	var pool CommissionPool = CommissionPool {
		Pool: sdk.NewInt64DecCoin(mt.ExtTokenType, 0),
	}
	if store.Has(key) {
		bz := store.Get(key)
		k.Cdc.MustUnmarshalJSON(bz, &pool)
	}
	pool.Pool = pool.Pool.Add(sdk.NewDecCoin(mt.ExtTokenType, extCoins.Amount))
	store.Set(key, k.Cdc.MustMarshalJSON(pool))
	
    // Mint stake and award to commission payer
    if doRebate {
        rebate := sdk.NewCoin(params.MintDenom, params.MintRatio.MulInt(extCoins.Amount).TruncateInt())
        mintCoins := sdk.Coins{ rebate }
    
        err := k.BankKeeper.MintCoins(ctx, MTModuleAccount, mintCoins)
        if err != nil {
    	    return nil, err
        }
    
	    err = k.BankKeeper.SendCoinsFromModuleToAccount(ctx, MTModuleAccount, addr, mintCoins)
	    if err != nil {
		    return nil, err
	    }
	    
	    return &rebate, nil
    }
	return nil, nil
}

func (k Keeper) Sweep(ctx sdk.Context) {
	store := ctx.KVStore(k.AppGlobalsKey)
	
	// Get current pool amount
	key := []byte(MTPoolName)
	var pool CommissionPool = CommissionPool {
		Pool: sdk.NewInt64DecCoin(mt.ExtTokenType, 0),
	}
	if store.Has(key) {
		bz := store.Get(key)
		k.Cdc.MustUnmarshalJSON(bz, &pool)
	}
	coin, _ := pool.Pool.TruncateDecimal()
	
	if coin.Amount.IsPositive() {
        //fmt.Printf("Sweep: %s %s\n", pool.String(), coin.String())
        err := k.BankKeeper.SendCoinsFromModuleToModule(ctx, MTModuleAccount, 
    	    auth.FeeCollectorName, sdk.Coins{coin})
        if err != nil {
    	    panic(fmt.Sprintf("Could not sweep fees: %s", coin.String()))
        }
	}
    	
    pool.Pool = sdk.NewInt64DecCoin(mt.ExtTokenType, 0)
    store.Set(key, k.Cdc.MustMarshalJSON(pool))
    
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
		err := k.BankKeeper.SendCoinsFromAccountToModule(ctx, account, MTModuleAccount, sdk.Coins{extCoins})
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
		err := k.BankKeeper.SendCoinsFromModuleToAccount(ctx, MTModuleAccount, account, sdk.Coins{extCoins})
		if err != nil {
			return err
		}
	}
	return nil
}

func (k Keeper) GetTotalBalance(ctx sdk.Context, addr sdk.AccAddress) (sdk.Dec, sdk.Dec) {
	params := k.GetParams(ctx)
	udai := k.BankKeeper.GetBalance(ctx, addr, mt.ExtTokenType)
	utick := k.BankKeeper.GetBalance(ctx, addr, params.MintDenom)
    return sdk.NewDecFromInt(udai.Amount).QuoInt64(1000000), sdk.NewDecFromInt(utick.Amount).QuoInt64(1000000)
}

func (k Keeper) RefundBacking(ctx sdk.Context) {
    k.IterateAccountStatus(ctx, 
        func(acct DataAccountStatus) (stop bool) {
        	err := k.DepositMicrotickCoin(ctx, acct.Account, acct.QuoteBacking)
        	if err != nil {
        		fmt.Fprintf(os.Stderr, "Could not refund quote backing for account: %s\n", acct.Account.String())
        	}
        	err = k.DepositMicrotickCoin(ctx, acct.Account, acct.TradeBacking)
        	if err != nil {
        		fmt.Fprintf(os.Stderr, "Could not refund trade backing for account: %s\n", acct.Account.String())
        	}
        	err = k.DepositMicrotickCoin(ctx, acct.Account, acct.SettleBacking)
        	if err != nil {
        		fmt.Fprintf(os.Stderr, "Could not refund settle backing for account: %s\n", acct.Account.String())
        	}
            return false
        },
    )
}