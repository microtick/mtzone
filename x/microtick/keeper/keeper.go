package keeper

import (
	"encoding/binary"
	"fmt"
	"strings"
	"time"
	
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/staking"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	
	mt "gitlab.com/microtick/mtzone/x/microtick/types"
)

type Keeper struct {
	Cdc codec.Marshaler
	AccountKeeper auth.AccountKeeper
	BankKeeper bank.Keeper
	DistrKeeper distribution.Keeper
	stakingKeeper staking.Keeper
	AppGlobalsKey sdk.StoreKey
	accountStatusKey sdk.StoreKey
	activeQuotesKey sdk.StoreKey
	activeTradesKey sdk.StoreKey
	marketsKey sdk.StoreKey
	durationsKey sdk.StoreKey
	paramSubspace params.Subspace
}

func NewKeeper(
  cdc codec.Marshaler, 
	accountKeeper auth.AccountKeeper, 
	bankKeeper bank.Keeper,
	distrKeeper distribution.Keeper,
	stakingKeeper staking.Keeper,
	mtAppGlobalsKey sdk.StoreKey,
	mtAccountStatusKey sdk.StoreKey,
	mtActiveQuotesKey sdk.StoreKey,
	mtActiveTradesKey sdk.StoreKey,
	mtMarketsKey sdk.StoreKey,
	mtDurationsKey sdk.StoreKey,
  paramstore params.Subspace,
) Keeper {
	if !paramstore.HasKeyTable() {
		paramstore = paramstore.WithKeyTable(mt.ParamKeyTable())
	}
	return Keeper {
		Cdc: cdc,
		AccountKeeper: accountKeeper,
		BankKeeper: bankKeeper,
		DistrKeeper: distrKeeper,
		stakingKeeper: stakingKeeper,
		AppGlobalsKey: mtAppGlobalsKey,
		accountStatusKey: mtAccountStatusKey,
		activeQuotesKey: mtActiveQuotesKey,
		activeTradesKey: mtActiveTradesKey,
		marketsKey: mtMarketsKey,
		durationsKey: mtDurationsKey,
		paramSubspace: paramstore,
	}
}

// Keeper as used here contains access methods for data structures only - business logic
// is maintained in the tx handlers

func (keeper Keeper) GetCodec() codec.Marshaler {
	return keeper.Cdc
}

type Termination struct {
	HaltTime int64 `json:"haltTime"`
}

func (keeper Keeper) GetHaltTime(ctx sdk.Context) int64 {
	store := ctx.KVStore(keeper.AppGlobalsKey)
	key := []byte("termination")
	bz := store.Get(key)
	var termination Termination
	keeper.Cdc.MustUnmarshalJSON(bz, &termination)
	return termination.HaltTime
}

// DataAccountStatus

func (k Keeper) GetAccountStatus(ctx sdk.Context, acct mt.MicrotickAccount) DataAccountStatus {
	store := ctx.KVStore(k.accountStatusKey)
	key := []byte(acct.String())
	if !store.Has(key) {
		return NewDataAccountStatus(acct)
	}
	bz := store.Get(key)
	var acctStatus DataAccountStatus
	k.Cdc.MustUnmarshalJSON(bz, &acctStatus)
	return acctStatus
}

func (k Keeper) SetAccountStatus(ctx sdk.Context, acct mt.MicrotickAccount, status DataAccountStatus) {
	store := ctx.KVStore(k.accountStatusKey)
	key := []byte(acct.String())
	status.Account = acct
	store.Set(key, k.Cdc.MustMarshalJSON(status))
}

func (k Keeper) IterateAccountStatus(ctx sdk.Context, process func(DataAccountStatus) (stop bool)) {
	store := ctx.KVStore(k.accountStatusKey)
	iter := sdk.KVStorePrefixIterator(store, nil) 
	defer iter.Close()
	for {
		if !iter.Valid() {
			return
		}
		bz := iter.Value()
		var acctStatus DataAccountStatus
		k.Cdc.MustUnmarshalJSON(bz, &acctStatus)
		if process(acctStatus) {
			return
		}
		iter.Next()
	}
}

// Durations

type InternalDuration struct {
	Name mt.MicrotickDurationName;
	Duration mt.MicrotickDuration;
}

func (k Keeper) AddDuration(ctx sdk.Context, name mt.MicrotickDurationName, dur mt.MicrotickDuration) {
	store := ctx.KVStore(k.durationsKey)
	keyByName := []byte(fmt.Sprintf("name:%s", name))
	keyByDur := []byte(fmt.Sprintf("dur:%d", dur))
	var id InternalDuration
	id.Name = name
	id.Duration = dur
	store.Set(keyByName, k.Cdc.MustMarshalJSON(id))
	store.Set(keyByDur, k.Cdc.MustMarshalJSON(id))
}

func (k Keeper) DurationFromName(ctx sdk.Context, name mt.MicrotickDurationName) mt.MicrotickDuration {
	store := ctx.KVStore(k.durationsKey)
	keyByName := []byte(fmt.Sprintf("name:%s", name))
	var id InternalDuration
	if !store.Has(keyByName) {
		panic("Invalid duration")
	}
	bz := store.Get(keyByName)
	k.Cdc.MustUnmarshalJSON(bz, &id)
	return id.Duration
}

func (k Keeper) NameFromDuration(ctx sdk.Context, dur mt.MicrotickDuration) mt.MicrotickDurationName {
	store := ctx.KVStore(k.durationsKey)
	keyByDur := []byte(fmt.Sprintf("dur:%d", dur))
	var id InternalDuration
	if !store.Has(keyByDur) {
		panic("Invalid duration")
	}
	bz := store.Get(keyByDur)
	k.Cdc.MustUnmarshalJSON(bz, &id)
	return id.Name
}

func (k Keeper) ValidDurationName(ctx sdk.Context, name mt.MicrotickDurationName) bool {
	store := ctx.KVStore(k.durationsKey)
	keyByName := []byte(fmt.Sprintf("name:%s", name))
	return store.Has(keyByName)
}

func (k Keeper) IterateDurations(ctx sdk.Context, process func(mt.MicrotickDurationName, mt.MicrotickDuration) (stop bool)) {
	store := ctx.KVStore(k.durationsKey)
	iter := sdk.KVStorePrefixIterator(store, nil) 
	defer iter.Close()
	for {
		if !iter.Valid() {
			return
		}
		key := iter.Key()
		if strings.HasPrefix(string(key), "name:") {
		  bz := iter.Value()
		  var id InternalDuration
		  k.Cdc.MustUnmarshalJSON(bz, &id)
		  if process(id.Name, id.Duration) {
			  return
		  }
		}
		iter.Next()
	}	
}


// DataMarket

func (k Keeper) HasDataMarket(ctx sdk.Context, market mt.MicrotickMarket) bool {
	store := ctx.KVStore(k.marketsKey)
	key := []byte(market)
	return store.Has(key)
}

func (k Keeper) GetDataMarket(ctx sdk.Context, market mt.MicrotickMarket) (DataMarket, error) {
	store := ctx.KVStore(k.marketsKey)
	key := []byte(market)
	var dataMarket DataMarket
	if !store.Has(key) {
		return dataMarket, sdkerrors.Wrap(mt.ErrInvalidMarket, market)
	}
	bz := store.Get(key)
	k.Cdc.MustUnmarshalJSON(bz, &dataMarket)
	return dataMarket, nil
}

func (k Keeper) SetDataMarket(ctx sdk.Context, dataMarket DataMarket) {
	store := ctx.KVStore(k.marketsKey)
	key := []byte(dataMarket.Market)
	store.Set(key, k.Cdc.MustMarshalJSON(dataMarket))
}

func (k Keeper) IterateMarkets(ctx sdk.Context, process func(DataMarket) (stop bool)) {
	store := ctx.KVStore(k.marketsKey)
	iter := sdk.KVStorePrefixIterator(store, nil) 
	defer iter.Close()
	for {
		if !iter.Valid() {
			return
		}
		bz := iter.Value()
		var market DataMarket
		k.Cdc.MustUnmarshalJSON(bz, &market)
		if process(market) {
			return
		}
		iter.Next()
	}	
}

// DataActiveQuote

func (k Keeper) GetNextActiveQuoteId(ctx sdk.Context) mt.MicrotickId {
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

func (k Keeper) GetActiveQuote(ctx sdk.Context, id mt.MicrotickId) (DataActiveQuote, error) {
	store := ctx.KVStore(k.activeQuotesKey)
	key := make([]byte, 4)
	var activeQuote DataActiveQuote
	binary.LittleEndian.PutUint32(key, id)
	if !store.Has(key) {
		return activeQuote, sdkerrors.Wrapf(mt.ErrInvalidQuote, "%i", id)
	}
	bz := store.Get(key)
	k.Cdc.MustUnmarshalJSON(bz, &activeQuote)
	return activeQuote, nil
}

func (k Keeper) SetActiveQuote(ctx sdk.Context, active DataActiveQuote) {
	store := ctx.KVStore(k.activeQuotesKey)
	key := make([]byte, 4)
	binary.LittleEndian.PutUint32(key, active.Id)
	store.Set(key, k.Cdc.MustMarshalJSON(active))
}

func (k Keeper) DeleteActiveQuote(ctx sdk.Context, id mt.MicrotickId) {
	store := ctx.KVStore(k.activeQuotesKey)
	key := make([]byte, 4)
	binary.LittleEndian.PutUint32(key, id)
	store.Delete(key)
}

// DataActiveTrade

func (k Keeper) GetNextActiveTradeId(ctx sdk.Context) mt.MicrotickId {
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

func (k Keeper) GetActiveTrade(ctx sdk.Context, id mt.MicrotickId) (DataActiveTrade, error) {
	store := ctx.KVStore(k.activeTradesKey)
	key := make([]byte, 4)
	var activeTrade DataActiveTrade
	binary.LittleEndian.PutUint32(key, id)
	if !store.Has(key) {
		return activeTrade, sdkerrors.Wrapf(mt.ErrInvalidTrade, "%i", id)
	}
	bz := store.Get(key)
	k.Cdc.MustUnmarshalJSON(bz, &activeTrade)
	return activeTrade, nil
}

func (k Keeper) SetActiveTrade(ctx sdk.Context, active DataActiveTrade) {
	store := ctx.KVStore(k.activeTradesKey)
	key := make([]byte, 4)
	binary.LittleEndian.PutUint32(key, active.Id)
	store.Set(key, k.Cdc.MustMarshalJSON(active))
}

func (k Keeper) DeleteActiveTrade(ctx sdk.Context, id mt.MicrotickId) {
	store := ctx.KVStore(k.activeTradesKey)
	key := make([]byte, 4)
	binary.LittleEndian.PutUint32(key, id)
	store.Delete(key)
}

// SetParams sets the module's parameters.
func (keeper Keeper) SetParams(ctx sdk.Context, params mt.Params) {
	keeper.paramSubspace.SetParamSet(ctx, &params)
	
	haltTime, _ := time.Parse(mt.TimeFormat, params.HaltTime)
	termination := Termination {
		HaltTime: haltTime.Unix(),
	}
	store := ctx.KVStore(keeper.AppGlobalsKey)
	key := []byte("termination")
	store.Set(key, keeper.Cdc.MustMarshalJSON(termination))
}

// GetParams gets the module's parameters.
func (keeper Keeper) GetParams(ctx sdk.Context) (params mt.Params) {
	keeper.paramSubspace.GetParamSet(ctx, &params)
	return params
}
