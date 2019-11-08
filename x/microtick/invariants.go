package microtick

import (
	//"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	//"github.com/cosmos/cosmos-sdk/x/auth"
	//staking "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// SupplyInvariants checks that the total supply reflects all held not-bonded tokens, bonded tokens, and unbonding delegations
// nolint: unparam
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(ModuleName, "constant-supply", ConstantSupply(k))
}

func ConstantSupply(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		/*
		// This needs to be entirely rewritten
		pool := k.distrKeeper.GetPool(ctx)

		loose := sdk.ZeroDec()
		bonded := sdk.ZeroDec()
		am.IterateAccounts(ctx, func(acc auth.Account) bool {
		    account := mtk.GetTotalBalance(ctx, acc.GetAddress())
		    status := mtk.GetAccountStatus(ctx, acc.GetAddress())
		    balance := account.Add(status.QuoteBacking).
			   Add(status.TradeBacking).Add(status.SettleBacking)
			//fmt.Printf("Balance: %s\n", balance.String())
			//fmt.Printf("  Initial: %s\n", account.String())
			//fmt.Printf("  Quote Backing: %s\n", status.QuoteBacking.String())
			//fmt.Printf("  Trade Backing: %s\n", status.TradeBacking.String())
			//fmt.Printf("  Settle Backing: %s\n", status.SettleBacking.String())
			loose = loose.Add(balance.Amount)
			return false
		})
		//fmt.Printf("Loose tokens - account balance total: %s\n", loose)
		k.IterateUnbondingDelegations(ctx, func(_ int64, ubd staking.UnbondingDelegation) bool {
			for _, entry := range ubd.Entries {
				loose = loose.Add(entry.Balance.ToDec())
			}
			return false
		})
		//fmt.Printf("Loose tokens - unbonding delegations: %s\n", loose)
		k.IterateValidators(ctx, func(_ int64, validator sdk.Validator) bool {
			switch validator.GetStatus() {
			case sdk.Bonded:
				bonded = bonded.Add(validator.GetBondedTokens().ToDec())
			case sdk.Unbonding, sdk.Unbonded:
				loose = loose.Add(validator.GetTokens().ToDec())
			}
			// add yet-to-be-withdrawn
			amount := d.GetValidatorOutstandingRewardsCoins(ctx, validator.GetOperator()).AmountOf(k.BondDenom(ctx))
			//fmt.Printf("  outstanding reward: %s\n", amount)
			loose = loose.Add(amount)
			return false
		})
		
		//fmt.Printf("Loose tokens - outstanding rewards: %s\n", loose)

		// add outstanding fees
		loose = loose.Add(f.GetCollectedFees(ctx).AmountOf(k.BondDenom(ctx)).ToDec())
		//fmt.Printf("Loose tokens - collected fees: %s\n", loose)

		// add community pool
		loose = loose.Add(d.GetFeePoolCommunityCoins(ctx).AmountOf(k.BondDenom(ctx)))
		//fmt.Printf("Loose tokens - community pool: %s\n", loose)
		
		// add fractional community pool
		frac := mtk.FractionalCommission(ctx)
		loose = loose.Add(frac.Amount)
		//fmt.Printf("Loose tokens - fractional community pool: %s\n", loose)

		// Not-bonded tokens should equal coin supply plus unbonding delegations
		// plus tokens on unbonded validators
		if !pool.NotBondedTokens.ToDec().Equal(loose) {
			return fmt.Errorf("loose token invariance:\n"+
				"\tpool.NotBondedTokens: %v\n"+
				"\tsum of account tokens: %v", pool.NotBondedTokens, loose)
		}

		// Bonded tokens should equal sum of tokens with bonded validators
		if !pool.BondedTokens.ToDec().Equal(bonded) {
			return fmt.Errorf("bonded token invariance:\n"+
				"\tpool.BondedTokens: %v\n"+
				"\tsum of account tokens: %v", pool.BondedTokens, bonded)
		}

		*/
		return "", false
	}
}