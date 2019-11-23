package keeper

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	mt "github.com/mjackson001/mtzone/x/microtick/types"
)

func MicrotickPoolInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		// Module account balance (Int)
		mtAcct := k.supplyKeeper.GetModuleAccount(ctx, MTModuleAccount)
		coins := mtAcct.GetCoins().AmountOf(mt.TokenType)
		
		// Commission pool (sdk.DecCoin)
		store := ctx.KVStore(k.AppGlobalsKey)
		key := []byte("commissionPool")
		var pool mt.MicrotickCoin = mt.NewMicrotickCoinFromInt(0)
		if store.Has(key) {
			bz := store.Get(key)
			k.cdc.MustUnmarshalBinaryBare(bz, &pool)
		}
		
		// Account change balance (sdk.DecCoin)
		var sum mt.MicrotickCoin = mt.NewMicrotickCoinFromInt(0)
		k.IterateAccountStatus(ctx,
			func(acct DataAccountStatus) (stop bool) {
				sum = sum.Add(acct.Change)
				sum = sum.Add(acct.QuoteBacking)
				sum = sum.Add(acct.TradeBacking)
				sum = sum.Add(acct.SettleBacking)
				return false
			},
		)
		
		sum = sum.Add(pool)
		
		if !sum.IsEqual(sdk.NewDecCoin(mt.TokenType, coins)) {
			return fmt.Sprintf("microtick commission pool invariance:\n\tmodule coins: %v\n\tsum of balances: %v\n", coins, sum), true
		}

		return "", false
	}
}