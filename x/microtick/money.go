package microtick

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
)

// Money

func (k Keeper) WithdrawDecCoin(ctx sdk.Context, account sdk.AccAddress, 
    withdrawAmount MicrotickCoin) {
	accountStatus := k.GetAccountStatus(ctx, account)
	
    if (accountStatus.Change.IsGTE(withdrawAmount)) {
        // handle without needing from the coin balance
        accountStatus.Change = accountStatus.Change.Minus(withdrawAmount)
    } else {
        neededAmount := withdrawAmount.Minus(accountStatus.Change)
	
	    // Load total coin balance + change into DecCoin
	    var amt sdk.Coin
	    var change sdk.DecCoin
	    amt, change = neededAmount.TruncateDecimal()
	    
	    if (change.IsPositive()) {
	        amt = amt.Plus(sdk.NewInt64Coin(TokenType, 1))
	        accountStatus.Change = sdk.NewDecCoinFromDec(TokenType, sdk.OneDec()).Minus(change)
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

func (k Keeper) DepositDecCoin(ctx sdk.Context, account sdk.AccAddress,
	depositAmount MicrotickCoin) {
	accountStatus := k.GetAccountStatus(ctx, account)
	
	totalDecCoin := accountStatus.Change.Plus(depositAmount)
	
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
