package keeper

import (
	"encoding/binary"
	"errors"
	"fmt"
	
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/staking"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	
	mt "github.com/mjackson001/mtzone/x/microtick/types"
)

type MicrotickKeeper struct {
	accountKeeper auth.AccountKeeper
	coinKeeper bank.Keeper
	distrKeeper distribution.Keeper
	stakingKeeper staking.Keeper
	appGlobalsKey sdk.StoreKey
	accountStatusKey sdk.StoreKey
	activeQuotesKey sdk.StoreKey
	activeTradesKey sdk.StoreKey
	marketsKey sdk.StoreKey
	cdc *codec.Codec 
	paramSubspace params.Subspace
}

func NewKeeper(accountKeeper auth.AccountKeeper, 
	coinKeeper bank.Keeper,
	distrKeeper distribution.Keeper,
	stakingKeeper staking.Keeper,
	mtAppGlobalsKey sdk.StoreKey,
	mtAccountStatusKey sdk.StoreKey,
	mtActiveQuotesKey sdk.StoreKey,
	mtActiveTradesKey sdk.StoreKey,
	mtMarketsKey sdk.StoreKey,
    cdc *codec.Codec, 
    paramstore params.Subspace,
) MicrotickKeeper {
	return MicrotickKeeper {
		accountKeeper: accountKeeper,
		coinKeeper: coinKeeper,
		distrKeeper: distrKeeper,
		stakingKeeper: stakingKeeper,
		appGlobalsKey: mtAppGlobalsKey,
		accountStatusKey: mtAccountStatusKey,
		activeQuotesKey: mtActiveQuotesKey,
		activeTradesKey: mtActiveTradesKey,
		marketsKey: mtMarketsKey,
		cdc: cdc,
		paramSubspace: paramstore.WithKeyTable(mt.ParamKeyTable()),
	}
}

// Keeper as used here contains access methods for data structures only - business logic
// is maintained in the tx handlers

// SetParams sets the module's parameters.
func (keeper MicrotickKeeper) SetParams(ctx sdk.Context, params mt.Params) {
	keeper.paramSubspace.SetParamSet(ctx, &params)
}

// GetParams gets the auth module's parameters.
func (keeper MicrotickKeeper) GetParams(ctx sdk.Context) (params mt.Params) {
	keeper.paramSubspace.GetParamSet(ctx, &params)
	return
}

// DataAccountStatus

func (k MicrotickKeeper) GetAccountStatus(ctx sdk.Context, acct mt.MicrotickAccount) DataAccountStatus {
	store := ctx.KVStore(k.accountStatusKey)
	key := []byte(acct.String())
	if !store.Has(key) {
		return NewDataAccountStatus(acct)
	}
	bz := store.Get(key)
	var acctStatus DataAccountStatus
	k.cdc.MustUnmarshalBinaryBare(bz, &acctStatus)
	return acctStatus
}

func (k MicrotickKeeper) SetAccountStatus(ctx sdk.Context, acct mt.MicrotickAccount, status DataAccountStatus) {
	store := ctx.KVStore(k.accountStatusKey)
	key := []byte(acct.String())
	status.Account = acct
	store.Set(key, k.cdc.MustMarshalBinaryBare(status))
}

// DataMarket

func (k MicrotickKeeper) HasDataMarket(ctx sdk.Context, market mt.MicrotickMarket) bool {
	store := ctx.KVStore(k.marketsKey)
	key := []byte(market)
	return store.Has(key)
}

func (k MicrotickKeeper) GetDataMarket(ctx sdk.Context, market mt.MicrotickMarket) (DataMarket, error) {
	store := ctx.KVStore(k.marketsKey)
	key := []byte(market)
	var dataMarket DataMarket
	if !store.Has(key) {
		return dataMarket, errors.New(fmt.Sprintf("No such market: {%s}", market))
	}
	bz := store.Get(key)
	k.cdc.MustUnmarshalBinaryBare(bz, &dataMarket)
	return dataMarket, nil
}

func (k MicrotickKeeper) SetDataMarket(ctx sdk.Context, dataMarket DataMarket) {
	store := ctx.KVStore(k.marketsKey)
	key := []byte(dataMarket.Market)
	store.Set(key, k.cdc.MustMarshalBinaryBare(dataMarket))
}

// DataActiveQuote

func (k MicrotickKeeper) GetNextActiveQuoteId(ctx sdk.Context) mt.MicrotickId {
	store := ctx.KVStore(k.activeQuotesKey)
	key := []byte("nextQuoteId")
	var id mt.MicrotickId
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

func (k MicrotickKeeper) GetActiveQuote(ctx sdk.Context, id mt.MicrotickId) (DataActiveQuote, error) {
	store := ctx.KVStore(k.activeQuotesKey)
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

func (k MicrotickKeeper) SetActiveQuote(ctx sdk.Context, active DataActiveQuote) {
	store := ctx.KVStore(k.activeQuotesKey)
	key := make([]byte, 4)
	binary.LittleEndian.PutUint32(key, active.Id)
	store.Set(key, k.cdc.MustMarshalBinaryBare(active))
}

func (k MicrotickKeeper) DeleteActiveQuote(ctx sdk.Context, id mt.MicrotickId) {
	store := ctx.KVStore(k.activeQuotesKey)
	key := make([]byte, 4)
	binary.LittleEndian.PutUint32(key, id)
	store.Delete(key)
}

// DataActiveTrade

func (k MicrotickKeeper) GetNextActiveTradeId(ctx sdk.Context) mt.MicrotickId {
	store := ctx.KVStore(k.activeQuotesKey)
	key := []byte("nextTradeId")
	var id mt.MicrotickId
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

func (k MicrotickKeeper) GetActiveTrade(ctx sdk.Context, id mt.MicrotickId) (DataActiveTrade, error) {
	store := ctx.KVStore(k.activeTradesKey)
	key := make([]byte, 4)
	var activeTrade DataActiveTrade
	binary.LittleEndian.PutUint32(key, id)
	if !store.Has(key) {
		return activeTrade, errors.New(fmt.Sprintf("No such trade ID: {%i}", id))
	}
	bz := store.Get(key)
	k.cdc.MustUnmarshalBinaryBare(bz, &activeTrade)
	return activeTrade, nil
}

func (k MicrotickKeeper) SetActiveTrade(ctx sdk.Context, active DataActiveTrade) {
	store := ctx.KVStore(k.activeTradesKey)
	key := make([]byte, 4)
	binary.LittleEndian.PutUint32(key, active.Id)
	store.Set(key, k.cdc.MustMarshalBinaryBare(active))
}

func (k MicrotickKeeper) DeleteActiveTrade(ctx sdk.Context, id mt.MicrotickId) {
	store := ctx.KVStore(k.activeTradesKey)
	key := make([]byte, 4)
	binary.LittleEndian.PutUint32(key, id)
	store.Delete(key)
}
