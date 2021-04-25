package keeper

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	
	"github.com/tendermint/tendermint/libs/log"
	
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	
	mt "github.com/microtick/mtzone/x/microtick/types"
)

const (
	AppGlobalExtTokenType = "ExtTokenType"
	AppGlobalIntTokenType = "IntTokenType"
	AppGlobalExtPerInt = "ExtPerInt"
)

type Keeper struct {
	Codec codec.Marshaler
	AccountKeeper authkeeper.AccountKeeper
	BankKeeper bankkeeper.Keeper
	DistrKeeper distrkeeper.Keeper
	stakingKeeper stakingkeeper.Keeper
	AppGlobalsKey sdk.StoreKey
	accountStatusKey sdk.StoreKey
	activeQuotesKey sdk.StoreKey
	activeTradesKey sdk.StoreKey
	marketsKey sdk.StoreKey
	durationsKey sdk.StoreKey
	paramSubspace paramtypes.Subspace
}

func NewKeeper(
	cdc codec.Marshaler, 
	paramSpace paramtypes.Subspace,
	accountKeeper authkeeper.AccountKeeper, 
	bankKeeper bankkeeper.Keeper,
	distrKeeper distrkeeper.Keeper,
	stakingKeeper stakingkeeper.Keeper,
	mtAppGlobalsKey sdk.StoreKey,
	mtAccountStatusKey sdk.StoreKey,
	mtActiveQuotesKey sdk.StoreKey,
	mtActiveTradesKey sdk.StoreKey,
	mtMarketsKey sdk.StoreKey,
	mtDurationsKey sdk.StoreKey,
) Keeper {
  if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(mt.ParamKeyTable())
	}
	return Keeper {
		Codec: cdc,
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
		paramSubspace: paramSpace,
	}
}

// Logger returns a module-specific logger.
func (keeper Keeper) Logger(ctx sdk.Context) log.Logger {
  return ctx.Logger().With("module", fmt.Sprintf("x/%s", mt.ModuleName))
}

// Keeper as used here contains access methods for data structures only - business logic
// is maintained in the tx handlers

func (k Keeper) SetBackingParams(ctx sdk.Context, backingDenom string, backingRatio string) {
	params := k.GetParams(ctx)
	params.BackingDenom = backingDenom
	params.BackingRatio = backingRatio
	k.SetParams(ctx, params)
}

func (k Keeper) GetBackingParams(ctx sdk.Context) (string, int) {
	params := k.GetParams(ctx)
	ratio, _ := strconv.Atoi(params.BackingRatio)
	return params.BackingDenom, ratio
}

func (k Keeper) MicrotickCoinToExtCoin(ctx sdk.Context, mc mt.MicrotickCoin) mt.ExtCoin {
	extTokenType, extPerInt := k.GetBackingParams(ctx)
	if mc.Denom != mt.IntTokenType {
    panic(fmt.Sprintf("Not internal token type: %s", mc.Denom))
  }
  mc.Amount = mc.Amount.MulInt64(int64(extPerInt))
  extCoin, _ := mc.TruncateDecimal()
  extCoin.Denom = extTokenType
  return extCoin
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
	k.Codec.MustUnmarshalJSON(bz, &acctStatus)
	return acctStatus
}

func (k Keeper) SetAccountStatus(ctx sdk.Context, acct mt.MicrotickAccount, status DataAccountStatus) {
	store := ctx.KVStore(k.accountStatusKey)
	key := []byte(acct.String())
	status.Account = acct
	store.Set(key, k.Codec.MustMarshalJSON(&status))
}

func (k Keeper) IterateAccountStatus(ctx sdk.Context, process func(*DataAccountStatus) (stop bool)) {
	store := ctx.KVStore(k.accountStatusKey)
	iter := sdk.KVStorePrefixIterator(store, nil) 
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		bz := iter.Value()
		var acctStatus DataAccountStatus
		k.Codec.MustUnmarshalJSON(bz, &acctStatus)
		if process(&acctStatus) {
			store.Set(iter.Key(), k.Codec.MustMarshalJSON(&acctStatus))
		}
	}
}

// Durations

func (k Keeper) AddDuration(ctx sdk.Context, name mt.MicrotickDurationName, dur mt.MicrotickDuration) {
	store := ctx.KVStore(k.durationsKey)
	keyByName := []byte(fmt.Sprintf("name:%s", name))
	keyByDur := []byte(fmt.Sprintf("dur:%d", dur))
	var id InternalDuration
	id.Name = name
	id.Duration = dur
	store.Set(keyByName, k.Codec.MustMarshalJSON(&id))
	store.Set(keyByDur, k.Codec.MustMarshalJSON(&id))
}

func (k Keeper) DurationFromName(ctx sdk.Context, name mt.MicrotickDurationName) mt.MicrotickDuration {
	store := ctx.KVStore(k.durationsKey)
	keyByName := []byte(fmt.Sprintf("name:%s", name))
	var id InternalDuration
	if !store.Has(keyByName) {
		panic("Invalid duration")
	}
	bz := store.Get(keyByName)
	k.Codec.MustUnmarshalJSON(bz, &id)
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
	k.Codec.MustUnmarshalJSON(bz, &id)
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
	for ; iter.Valid(); iter.Next() {
		key := iter.Key()
		if strings.HasPrefix(string(key), "name:") {
		  bz := iter.Value()
		  var id InternalDuration
		  k.Codec.MustUnmarshalJSON(bz, &id)
		  if process(id.Name, id.Duration) {
			  return
		  }
		}
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
	k.Codec.MustUnmarshalJSON(bz, &dataMarket)
	return dataMarket, nil
}

func (k Keeper) AssertDataMarketHasDuration(ctx sdk.Context, market mt.MicrotickMarket, name mt.MicrotickDurationName) {
	var found bool = false
	dataMarket, _ := k.GetDataMarket(ctx, market)
	for i := range dataMarket.OrderBooks {
		// test
		if dataMarket.OrderBooks[i].Name == name {
			found = true
		}
	}
	if found {
		return
	} else {
		seconds := k.DurationFromName(ctx, name)
		// insert
		orderBooks := make([]DataOrderBook, 0)
		var added bool = false
	  for i := range dataMarket.OrderBooks {
	  	curSeconds := k.DurationFromName(ctx, dataMarket.OrderBooks[i].Name)
	  	if seconds < curSeconds && !added {
	  		orderBooks = append(orderBooks, NewOrderBook(name))
	  		added = true
	  	}
	  	orderBooks = append(orderBooks, dataMarket.OrderBooks[i])
	  }
	  if !added {
	  	orderBooks = append(orderBooks, NewOrderBook(name))
	  }
		dataMarket.OrderBooks = orderBooks
		k.SetDataMarket(ctx, dataMarket)
	}
}

func (k Keeper) SetDataMarket(ctx sdk.Context, dataMarket DataMarket) {
	store := ctx.KVStore(k.marketsKey)
	key := []byte(dataMarket.Market)
	store.Set(key, k.Codec.MustMarshalJSON(&dataMarket))
}

func (k Keeper) IterateMarkets(ctx sdk.Context, process func(*DataMarket) (stop bool)) {
	store := ctx.KVStore(k.marketsKey)
	iter := sdk.KVStorePrefixIterator(store, nil) 
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		bz := iter.Value()
		var market DataMarket
		k.Codec.MustUnmarshalJSON(bz, &market)
		if process(&market) {
			store.Set(iter.Key(), k.Codec.MustMarshalJSON(&market))
		}
	}	
}

// DataActiveQuote

func (k Keeper) GetNextActiveQuoteId(ctx sdk.Context) mt.MicrotickId {
	store := ctx.KVStore(k.activeQuotesKey)
	key := []byte("nextQuoteId")
	var id mt.MicrotickId
	if !store.Has(key) {
		id = 1
	} else {
		val := store.Get(key)
		id = binary.LittleEndian.Uint32(val)
		id++
	}
	return id
}

func (k Keeper) CommitQuoteId(ctx sdk.Context, id mt.MicrotickId) {
	store := ctx.KVStore(k.activeQuotesKey)
	key := []byte("nextQuoteId")
	val := make([]byte, 4)
	binary.LittleEndian.PutUint32(val, id)
	store.Set(key, val)
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
	k.Codec.MustUnmarshalJSON(bz, &activeQuote)
	return activeQuote, nil
}

func (k Keeper) SetActiveQuote(ctx sdk.Context, active DataActiveQuote) {
	store := ctx.KVStore(k.activeQuotesKey)
	key := make([]byte, 4)
	binary.LittleEndian.PutUint32(key, active.Id)
	store.Set(key, k.Codec.MustMarshalJSON(&active))
}

func (k Keeper) DeleteActiveQuote(ctx sdk.Context, id mt.MicrotickId) {
	store := ctx.KVStore(k.activeQuotesKey)
	key := make([]byte, 4)
	binary.LittleEndian.PutUint32(key, id)
	store.Delete(key)
}

func (k Keeper) IterateQuotes(ctx sdk.Context, process func(DataActiveQuote) (stop bool)) {
	store := ctx.KVStore(k.activeQuotesKey)
	iter := sdk.KVStorePrefixIterator(store, nil) 
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		if string(iter.Key()) != "nextQuoteId" {
		  bz := iter.Value()
		  var activeQuote DataActiveQuote
		  k.Codec.MustUnmarshalJSON(bz, &activeQuote)
		  if process(activeQuote) {
			  store.Delete(iter.Key())
		  }
		}
	}	
}

// DataActiveTrade

func (k Keeper) GetNextActiveTradeId(ctx sdk.Context) mt.MicrotickId {
	store := ctx.KVStore(k.activeTradesKey)
	key := []byte("nextTradeId")
	var id mt.MicrotickId
	if !store.Has(key) {
		id = 1
	} else {
		val := store.Get(key)
		id = binary.LittleEndian.Uint32(val)
		id++
	}
	return id
}

func (k Keeper) CommitTradeId(ctx sdk.Context, id mt.MicrotickId) {
	store := ctx.KVStore(k.activeTradesKey)
	key := []byte("nextTradeId")
	val := make([]byte, 4)
	binary.LittleEndian.PutUint32(val, id)
	store.Set(key, val)
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
	k.Codec.MustUnmarshalJSON(bz, &activeTrade)
	return activeTrade, nil
}

func (k Keeper) SetActiveTrade(ctx sdk.Context, active DataActiveTrade) {
	store := ctx.KVStore(k.activeTradesKey)
	key := make([]byte, 4)
	binary.LittleEndian.PutUint32(key, active.Id)
	store.Set(key, k.Codec.MustMarshalJSON(&active))
}

func (k Keeper) DeleteActiveTrade(ctx sdk.Context, id mt.MicrotickId) {
	store := ctx.KVStore(k.activeTradesKey)
	key := make([]byte, 4)
	binary.LittleEndian.PutUint32(key, id)
	store.Delete(key)
}

func (k Keeper) IterateTrades(ctx sdk.Context, process func(DataActiveTrade) (stop bool)) {
	store := ctx.KVStore(k.activeTradesKey)
	iter := sdk.KVStorePrefixIterator(store, nil) 
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		if string(iter.Key()) != "nextTradeId" {
		  bz := iter.Value()
		  var activeTrade DataActiveTrade
		  k.Codec.MustUnmarshalJSON(bz, &activeTrade)
		  if process(activeTrade) {
			  store.Delete(iter.Key())
		  }
		}
	}	
}

// SetParams sets the module's parameters.
func (keeper Keeper) SetParams(ctx sdk.Context, params mt.MicrotickParams) {
	keeper.paramSubspace.SetParamSet(ctx, &params)
}

// GetParams gets the module's parameters.
func (keeper Keeper) GetParams(ctx sdk.Context) (params mt.MicrotickParams) {
	keeper.paramSubspace.GetParamSet(ctx, &params)
	return params
}

// Clear markets, quotes, trades

func (keeper Keeper) ClearMarkets(ctx sdk.Context) {
	// Clear all quotes
	keeper.IterateQuotes(ctx, 
	  func(quote DataActiveQuote) bool {
	   	return true
	  },
	)
	
	// Clear all trades
	keeper.IterateTrades(ctx, 
	  func(trade DataActiveTrade) bool {
	  	return true
	  },
	)
	
	// Update markets
	keeper.IterateMarkets(ctx, 
		func(market *DataMarket) bool {
			market.TotalBacking = mt.NewMicrotickCoinFromInt(0)
			market.TotalSpots = sdk.ZeroDec()
			market.TotalWeight = mt.NewMicrotickQuantityFromInt(0)
			for i := 0; i < len(market.OrderBooks); i++ {
				market.OrderBooks[i].CallAsks = NewOrderedList()
				market.OrderBooks[i].CallBids = NewOrderedList()
				market.OrderBooks[i].PutAsks = NewOrderedList()
				market.OrderBooks[i].PutBids = NewOrderedList()
				market.OrderBooks[i].SumBacking = mt.NewMicrotickCoinFromInt(0)
				market.OrderBooks[i].SumWeight = mt.NewMicrotickQuantityFromInt(0)
			}
			return true
		},
	)
	
	// Update accounts
	keeper.IterateAccountStatus(ctx, 
	  func(account *DataAccountStatus) bool {
	  	account.ActiveQuotes = NewOrderedList()
	  	account.ActiveTrades = NewOrderedList()
	  	keeper.DepositMicrotickCoin(ctx, account.Account, account.QuoteBacking)
	  	keeper.DepositMicrotickCoin(ctx, account.Account, account.TradeBacking)
	  	keeper.DepositMicrotickCoin(ctx, account.Account, account.SettleBacking)
	  	account.QuoteBacking = mt.NewMicrotickCoinFromInt(0)
	  	account.TradeBacking = mt.NewMicrotickCoinFromInt(0)
	  	account.SettleBacking = mt.NewMicrotickCoinFromInt(0)
	  	return true
	  },
	)
}
