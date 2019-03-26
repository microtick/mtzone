package microtick

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/bank"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Keeper struct {
	coinKeeper bank.Keeper
	storeKey  sdk.StoreKey 
	cdc *codec.Codec 
}

func NewKeeper(coinKeeper bank.Keeper, storeKey sdk.StoreKey, cdc *codec.Codec) Keeper {
	return Keeper{
		coinKeeper: coinKeeper,
		storeKey:   storeKey,
		cdc:        cdc,
	}
}

func (k Keeper) GetAccountStatus(ctx sdk.Context, acct string) AccountInfo {
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(acct)) {
		return NewAccountInfo(acct)
	}
	bz := store.Get([]byte(acct))
	var acctInfo AccountInfo
	k.cdc.MustUnmarshalBinaryBare(bz, &acctInfo)
	return acctInfo
}