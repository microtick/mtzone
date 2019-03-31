package microtick

import (
	"encoding/binary"
	"errors"
	"fmt"
	
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/bank"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type MicrotickStores struct {
	AccountStatus sdk.StoreKey
	ActiveQuotes sdk.StoreKey
	ActiveTrades sdk.StoreKey
	Markets sdk.StoreKey
}

type Keeper struct {
	coinKeeper bank.Keeper
	storeKeys MicrotickStores
	cdc *codec.Codec 
}

func NewKeeper(coinKeeper bank.Keeper, storeKeys MicrotickStores, cdc *codec.Codec) Keeper {
	return Keeper {
		coinKeeper: coinKeeper,
		storeKeys:   storeKeys,
		cdc:        cdc,
	}
}

// Keeper as used here contains access methods for data structures only - business logic
// is maintained in the tx handlers

// DataAccountStatus

func (k Keeper) GetAccountStatus(ctx sdk.Context, acct string) DataAccountStatus {
	store := ctx.KVStore(k.storeKeys.AccountStatus)
	key := []byte(acct)
	if !store.Has(key) {
		return NewDataAccountStatus(acct)
	}
	bz := store.Get(key)
	var acctStatus DataAccountStatus
	k.cdc.MustUnmarshalBinaryBare(bz, &acctStatus)
	return acctStatus
}

func (k Keeper) SetAccountStatus(ctx sdk.Context, acct string, status DataAccountStatus) {
	store := ctx.KVStore(k.storeKeys.AccountStatus)
	key := []byte(acct)
	status.Account = acct
	store.Set(key, k.cdc.MustMarshalBinaryBare(status))
}

// DataActiveQuote

func (k Keeper) GetNextActiveQuoteId(ctx sdk.Context) MicrotickId {
	store := ctx.KVStore(k.storeKeys.ActiveQuotes)
	key := []byte("nextQuoteId")
	var id MicrotickId
	var val []byte
	if !store.Has(key) {
		val = make([]byte, 4)
		id = 1
	} else {
		val = store.Get(key)
		id = binary.LittleEndian.Uint32(val)
		id++
	}
	binary.LittleEndian.PutUint32(val, id)
	store.Set(key, val)
	return id
}

func (k Keeper) GetActiveQuote(ctx sdk.Context, id MicrotickId) (DataActiveQuote, error) {
	store := ctx.KVStore(k.storeKeys.ActiveQuotes)
	key := make([]byte, 4)
	var activeQuote DataActiveQuote
	binary.LittleEndian.PutUint32(key, id)
	if !store.Has(key) {
		return activeQuote, errors.New(fmt.Sprintf("No such quote ID: {%i}", id))
	}
	bz := store.Get(key)
	k.cdc.MustUnmarshalBinaryBare(bz, &activeQuote)
	return activeQuote, nil
}

func (k Keeper) SetActiveQuote(ctx sdk.Context, active DataActiveQuote) {
	store := ctx.KVStore(k.storeKeys.ActiveQuotes)
	key := make([]byte, 4)
	binary.LittleEndian.PutUint32(key, active.Id)
	store.Set(key, k.cdc.MustMarshalBinaryBare(active))
}
  
  
// DataMarket

func (k Keeper) HasDataMarket(ctx sdk.Context, market MicrotickMarket) bool {
	store := ctx.KVStore(k.storeKeys.Markets)
	key := []byte(market)
	return store.Has(key)
}

func (k Keeper) GetDataMarket(ctx sdk.Context, market MicrotickMarket) (DataMarket, error) {
	store := ctx.KVStore(k.storeKeys.Markets)
	key := []byte(market)
	var dataMarket DataMarket
	if !store.Has(key) {
		return dataMarket, errors.New(fmt.Sprintf("No such market: {%s}", market))
	}
	bz := store.Get(key)
	k.cdc.MustUnmarshalBinaryBare(bz, &dataMarket)
	return dataMarket, nil
}

func (k Keeper) SetDataMarket(ctx sdk.Context, dataMarket DataMarket) {
	store := ctx.KVStore(k.storeKeys.Markets)
	key := []byte(dataMarket.Market)
	store.Set(key, k.cdc.MustMarshalBinaryBare(dataMarket))
}
