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

type Keeper struct {
	AccountKeeper auth.AccountKeeper
	CoinKeeper bank.Keeper
	DistrKeeper distribution.Keeper
	stakingKeeper staking.Keeper
	AppGlobalsKey sdk.StoreKey
	accountStatusKey sdk.StoreKey
	activeQuotesKey sdk.StoreKey
	activeTradesKey sdk.StoreKey
	marketsKey sdk.StoreKey
	paramSubspace params.Subspace
	cdc *codec.Codec
}

func NewKeeper(
	accountKeeper auth.AccountKeeper, 
	coinKeeper bank.Keeper,
	distrKeeper distribution.Keeper,
	stakingKeeper staking.Keeper,
	mtAppGlobalsKey sdk.StoreKey,
	mtAccountStatusKey sdk.StoreKey,
	mtActiveQuotesKey sdk.StoreKey,
	mtActiveTradesKey sdk.StoreKey,
	mtMarketsKey sdk.StoreKey,
    paramstore params.Subspace,
    cdc *codec.Codec,
) Keeper {
	return Keeper {
		AccountKeeper: accountKeeper,
		CoinKeeper: coinKeeper,
		DistrKeeper: distrKeeper,
		stakingKeeper: stakingKeeper,
		AppGlobalsKey: mtAppGlobalsKey,
		accountStatusKey: mtAccountStatusKey,
		activeQuotesKey: mtActiveQuotesKey,
		activeTradesKey: mtActiveTradesKey,
		marketsKey: mtMarketsKey,
		paramSubspace: paramstore.WithKeyTable(mt.ParamKeyTable()),
		cdc: cdc,
	}
}

// Keeper as used here contains access methods for data structures only - business logic
// is maintained in the tx handlers

func (keeper Keeper) GetCodec() *codec.Codec {
	return keeper.cdc
}

// SetParams sets the module's parameters.
func (keeper Keeper) SetParams(ctx sdk.Context, params mt.Params) {
	keeper.paramSubspace.SetParamSet(ctx, &params)
}

// GetParams gets the auth module's parameters.
func (keeper Keeper) GetParams(ctx sdk.Context) (params mt.Params) {
	keeper.paramSubspace.GetParamSet(ctx, &params)
	return
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
	k.cdc.MustUnmarshalBinaryBare(bz, &acctStatus)
	return acctStatus
}

func (k Keeper) SetAccountStatus(ctx sdk.Context, acct mt.MicrotickAccount, status DataAccountStatus) {
	store := ctx.KVStore(k.accountStatusKey)
	key := []byte(acct.String())
	status.Account = acct
	store.Set(key, k.cdc.MustMarshalBinaryBare(status))
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
		return dataMarket, errors.New(fmt.Sprintf("No such market: {%s}", market))
	}
	bz := store.Get(key)
	k.cdc.MustUnmarshalBinaryBare(bz, &dataMarket)
	return dataMarket, nil
}

func (k Keeper) SetDataMarket(ctx sdk.Context, dataMarket DataMarket) {
	store := ctx.KVStore(k.marketsKey)
	key := []byte(dataMarket.Market)
	store.Set(key, k.cdc.MustMarshalBinaryBare(dataMarket))
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
		return activeQuote, errors.New(fmt.Sprintf("No such quote ID: {%i}", id))
	}
	bz := store.Get(key)
	k.cdc.MustUnmarshalBinaryBare(bz, &activeQuote)
	return activeQuote, nil
}

func (k Keeper) SetActiveQuote(ctx sdk.Context, active DataActiveQuote) {
	store := ctx.KVStore(k.activeQuotesKey)
	key := make([]byte, 4)
	binary.LittleEndian.PutUint32(key, active.Id)
	store.Set(key, k.cdc.MustMarshalBinaryBare(active))
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
		return activeTrade, errors.New(fmt.Sprintf("No such trade ID: {%i}", id))
	}
	bz := store.Get(key)
	k.cdc.MustUnmarshalBinaryBare(bz, &activeTrade)
	return activeTrade, nil
}

func (k Keeper) SetActiveTrade(ctx sdk.Context, active DataActiveTrade) {
	store := ctx.KVStore(k.activeTradesKey)
	key := make([]byte, 4)
	binary.LittleEndian.PutUint32(key, active.Id)
	store.Set(key, k.cdc.MustMarshalBinaryBare(active))
}

func (k Keeper) DeleteActiveTrade(ctx sdk.Context, id mt.MicrotickId) {
	store := ctx.KVStore(k.activeTradesKey)
	key := make([]byte, 4)
	binary.LittleEndian.PutUint32(key, id)
	store.Delete(key)
}
