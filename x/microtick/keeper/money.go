package keeper

import (
	"fmt"
	"os"
    sdk "github.com/cosmos/cosmos-sdk/types"
    auth "github.com/cosmos/cosmos-sdk/x/auth/types"
    mt "gitlab.com/microtick/mtzone/x/microtick/types"
)

const MTModuleAccount = "microtick"
const MTPoolName = "commissionPool"

// Commissions

func (k Keeper) PoolCommission(ctx sdk.Context, commission sdk.Dec) mt.MicrotickCoin {
	amount := mt.NewMicrotickCoinFromDec(commission)
    extCoins := k.MicrotickCoinToExtCoin(ctx, amount)
    
	store := ctx.KVStore(k.AppGlobalsKey)
	
	// Get current pool amount
	key := []byte(MTPoolName)
	var pool CommissionPool = CommissionPool {
		Pool: sdk.NewInt64DecCoin(k.GetExtTokenType(ctx), 0),
	}
	if store.Has(key) {
		bz := store.Get(key)
		k.Codec.MustUnmarshalJSON(bz, &pool)
	}
	pool.Pool = pool.Pool.Add(sdk.NewDecCoin(k.GetExtTokenType(ctx), extCoins.Amount))
	store.Set(key, k.Codec.MustMarshalJSON(&pool))
	
	return amount
}

func (k Keeper) AwardRebate(ctx sdk.Context, addr sdk.AccAddress, rebate sdk.Dec) (*sdk.Coin, error) {
	params := k.GetParams(ctx)
    coin := sdk.NewCoin(params.MintDenom, rebate.TruncateInt())
    if coin.Amount.IsPositive() {
        mintCoins := sdk.Coins{ coin }

        err := k.BankKeeper.MintCoins(ctx, MTModuleAccount, mintCoins)
        if err != nil {
	        return nil, err
        }

        err = k.BankKeeper.SendCoinsFromModuleToAccount(ctx, MTModuleAccount, addr, mintCoins)
        if err != nil {
	        return nil, err
        }
    }
    return &coin, nil
}

func (k Keeper) Sweep(ctx sdk.Context) {
	store := ctx.KVStore(k.AppGlobalsKey)
	
	// Get current pool amount
	key := []byte(MTPoolName)
	var pool CommissionPool = CommissionPool {
		Pool: sdk.NewInt64DecCoin(k.GetExtTokenType(ctx), 0),
	}
	if store.Has(key) {
		bz := store.Get(key)
		k.Codec.MustUnmarshalJSON(bz, &pool)
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
    	
    pool.Pool = sdk.NewInt64DecCoin(k.GetExtTokenType(ctx), 0)
    store.Set(key, k.Codec.MustMarshalJSON(&pool))
    
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
    	
    extCoins := k.MicrotickCoinToExtCoin(ctx, withdrawAmount)
    
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
		
	extCoins := k.MicrotickCoinToExtCoin(ctx, depositAmount)	
	
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
	extBacking := k.BankKeeper.GetBalance(ctx, addr, k.GetExtTokenType(ctx))
    backing := sdk.NewDecFromInt(extBacking.Amount).QuoInt64(int64(k.GetExtPerInt(ctx)))
	ustake := k.BankKeeper.GetBalance(ctx, addr, params.MintDenom)
    stake := sdk.NewDecFromInt(ustake.Amount).QuoInt64(1000000)
    return backing, stake
}

func (k Keeper) RefundBacking(ctx sdk.Context) {
    k.IterateAccountStatus(ctx, 
        func(acct *DataAccountStatus) (stop bool) {
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
