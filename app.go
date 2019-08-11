package app

import (
	"fmt"
	"io"
	"os"
	"io/ioutil"
	"path/filepath"
	"errors"
	"strings"
	"sort"
	"time"
	"encoding/json"
	"log"

	tlog "github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/mjackson001/mtzone/x/microtick"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	tmtypes "github.com/tendermint/tendermint/types"
)

const (
    appName = "Microtick"
    BondDenom = "fox"
)

var (
	DefaultCLIHome = os.ExpandEnv("$MTROOT/mtcli")
	DefaultNodeHome = os.ExpandEnv("$MTROOT/mtd")
)

func init() {
	sdk.PowerReduction = sdk.NewInt(1)
}

type mtApp struct {
    *bam.BaseApp
	cdc *codec.Codec

	keyMain          *sdk.KVStoreKey
	keyAccount       *sdk.KVStoreKey
	keyFeeCollection *sdk.KVStoreKey
	
	keyDistr		 *sdk.KVStoreKey
	tkeyDistr		 *sdk.TransientStoreKey
	keySlashing		 *sdk.KVStoreKey
	keyStaking		 *sdk.KVStoreKey
	tkeyStaking		 *sdk.TransientStoreKey
	keyGov			 *sdk.KVStoreKey
	
	keyParams        *sdk.KVStoreKey
	tkeyParams       *sdk.TransientStoreKey
	keyMT            microtick.MicrotickStores

	accountKeeper       auth.AccountKeeper
	bankKeeper          bank.Keeper
	feeCollectionKeeper auth.FeeCollectionKeeper
	stakingKeeper		staking.Keeper
	slashingKeeper		slashing.Keeper
	distrKeeper			distr.Keeper
	govKeeper			gov.Keeper
	paramsKeeper        params.Keeper
	mtKeeper            microtick.Keeper
}

func NewMTApp(logger tlog.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool, 
	baseAppOptions ...func(*bam.BaseApp)) *mtApp {
		
    cdc := MakeCodec()

    bApp := bam.NewBaseApp(appName, logger, db, auth.DefaultTxDecoder(cdc), baseAppOptions...)
    bApp.SetCommitMultiStoreTracer(traceStore)
    
    var mtStores = microtick.MicrotickStores {
    	AppGlobals: sdk.NewKVStoreKey("MTGlobals"),
    	AccountStatus: sdk.NewKVStoreKey("MTAccountStatus"),
    	ActiveQuotes: sdk.NewKVStoreKey("MTActiveQuotes"),
    	ActiveTrades: sdk.NewKVStoreKey("MTActiveTrades"),
    	Markets: sdk.NewKVStoreKey("MTMarkets"),
    }

    var app = &mtApp{
        BaseApp: bApp,
        cdc:     cdc,
        
		keyMain:          sdk.NewKVStoreKey("main"),
		keyAccount:       sdk.NewKVStoreKey(auth.StoreKey),
		keyFeeCollection: sdk.NewKVStoreKey(auth.FeeStoreKey),
		
		keyDistr:         sdk.NewKVStoreKey(distr.StoreKey),
		tkeyDistr:        sdk.NewTransientStoreKey(distr.TStoreKey),
		keySlashing:      sdk.NewKVStoreKey(slashing.StoreKey),
		keyGov:           sdk.NewKVStoreKey(gov.StoreKey),
		keyStaking:       sdk.NewKVStoreKey(staking.StoreKey),
		tkeyStaking:      sdk.NewTransientStoreKey(staking.TStoreKey),
		
		keyParams:        sdk.NewKVStoreKey(params.StoreKey),
		tkeyParams:       sdk.NewTransientStoreKey(params.TStoreKey),
		
		keyMT:			  mtStores,
    }
    
 	// The ParamsKeeper handles parameter storage for the application
	app.paramsKeeper = params.NewKeeper(app.cdc, app.keyParams, app.tkeyParams)

	// The AccountKeeper handles address -> account lookups
	app.accountKeeper = auth.NewAccountKeeper(
		app.cdc,
		app.keyAccount,
		app.paramsKeeper.Subspace(auth.DefaultParamspace),
		auth.ProtoBaseAccount,
	)

	// The BankKeeper allows you perform sdk.Coins interactions
	app.bankKeeper = bank.NewBaseKeeper(
		app.accountKeeper,
		app.paramsKeeper.Subspace(bank.DefaultParamspace),
		bank.DefaultCodespace,
	)

	// The FeeCollectionKeeper collects transaction fees and renders them to the fee distribution module
	app.feeCollectionKeeper = auth.NewFeeCollectionKeeper(
		app.cdc, 
		app.keyFeeCollection,
	)
	
	stakingKeeper := staking.NewKeeper(
		app.cdc,
		app.keyStaking, app.tkeyStaking,
		app.bankKeeper, app.paramsKeeper.Subspace(staking.DefaultParamspace),
		staking.DefaultCodespace,
	)
	app.distrKeeper = distr.NewKeeper(
		app.cdc,
		app.keyDistr,
		app.paramsKeeper.Subspace(distr.DefaultParamspace),
		app.bankKeeper, &stakingKeeper, app.feeCollectionKeeper,
		distr.DefaultCodespace,
	)
	app.slashingKeeper = slashing.NewKeeper(
		app.cdc,
		app.keySlashing,
		&stakingKeeper, app.paramsKeeper.Subspace(slashing.DefaultParamspace),
		slashing.DefaultCodespace,
	)
	app.govKeeper = gov.NewKeeper(
		app.cdc,
		app.keyGov,
		app.paramsKeeper, app.paramsKeeper.Subspace(gov.DefaultParamspace), app.bankKeeper, &stakingKeeper,
		gov.DefaultCodespace,
	)
	
	app.stakingKeeper = *stakingKeeper.SetHooks(
		NewStakingHooks(app.distrKeeper.Hooks(), app.slashingKeeper.Hooks()),
	)
	
	app.mtKeeper = microtick.NewKeeper(
		app.accountKeeper,
		app.bankKeeper,
		app.feeCollectionKeeper,
		app.keyMT,
		app.cdc,
		app.paramsKeeper.Subspace(microtick.DefaultParamspace),
	)

	// The app.Router is the main transaction router where each module registers its routes
	// Register the bank and nameservice routes here
	app.Router().
		AddRoute(bank.RouterKey, NewBankHandler(app.bankKeeper, app.mtKeeper, app.cdc)).
		AddRoute(staking.RouterKey, staking.NewHandler(app.stakingKeeper)).
		AddRoute(distr.RouterKey, distr.NewHandler(app.distrKeeper)).
		AddRoute(slashing.RouterKey, slashing.NewHandler(app.slashingKeeper)).
		AddRoute(gov.RouterKey, gov.NewHandler(app.govKeeper)).
		AddRoute("microtick", microtick.NewHandler(app.mtKeeper))

	// The app.QueryRouter is the main query router where each module registers its routes
	app.QueryRouter().
		AddRoute(auth.QuerierRoute, auth.NewQuerier(app.accountKeeper)).
		AddRoute(distr.QuerierRoute, distr.NewQuerier(app.distrKeeper)).
		AddRoute(gov.QuerierRoute, gov.NewQuerier(app.govKeeper)).
		AddRoute(slashing.QuerierRoute, slashing.NewQuerier(app.slashingKeeper, app.cdc)).
		AddRoute(staking.QuerierRoute, staking.NewQuerier(app.stakingKeeper, app.cdc)).
		AddRoute("microtick", microtick.NewQuerier(app.mtKeeper))

	app.MountStores(
		app.keyMain,
		app.keyAccount,
		app.keyFeeCollection,
		app.keyStaking,
		app.tkeyStaking,
		app.keyDistr,
		app.tkeyDistr,
		app.keySlashing,
		app.keyGov,
		app.keyParams,
		app.tkeyParams,
		app.keyMT.AppGlobals,
		app.keyMT.AccountStatus,
		app.keyMT.ActiveQuotes,
		app.keyMT.ActiveTrades,
		app.keyMT.Markets,
	)
	
	app.SetInitChainer(app.initChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetAnteHandler(auth.NewAnteHandler(app.accountKeeper, app.feeCollectionKeeper))
	app.SetEndBlocker(app.EndBlocker)

	if loadLatest {
		err := app.LoadLatestVersion(app.keyMain)
		if err != nil {
			cmn.Exit(err.Error())
		}
	}

    return app
}

// GenesisState represents chain state at the start of the chain. Any initial state (account balances) are stored here.
type GenesisState struct {
	Accounts []GenesisAccount    `json:"accounts"`
	AuthData auth.GenesisState   `json:"auth"`
	BankData bank.GenesisState   `json:"bank"`
	StakingData  staking.GenesisState  `json:"staking"`
	DistrData    distr.GenesisState    `json:"distr"`
	GovData      gov.GenesisState      `json:"gov"`
	SlashingData slashing.GenesisState `json:"slashing"`	
	MicrotickData microtick.GenesisState `json:"microtick"`
	GenTxs []json.RawMessage `json:"gentxs"`
}

// NewDefaultGenesisState generates the default state
func NewDefaultGenesisState() GenesisState {
	return GenesisState{
		Accounts:     nil,
		AuthData:     auth.DefaultGenesisState(),
		BankData:     bank.DefaultGenesisState(),
		StakingData:  staking.DefaultGenesisState(),
		DistrData:    distr.DefaultGenesisState(),
		GovData:      gov.DefaultGenesisState(),
		SlashingData: slashing.DefaultGenesisState(),
		MicrotickData: microtick.DefaultGenesisState(),
	}
}

func NewGenesisState(accounts []GenesisAccount, authData auth.GenesisState,
	bankData bank.GenesisState,
	stakingData staking.GenesisState, distrData distr.GenesisState, govData gov.GenesisState, 
	slashingData slashing.GenesisState, microtickData microtick.GenesisState) GenesisState {

	return GenesisState{
		Accounts:     accounts,
		AuthData:     authData,
		BankData:     bankData,
		StakingData:  stakingData,
		DistrData:    distrData,
		GovData:      govData,
		SlashingData: slashingData,
		MicrotickData: microtickData,
	}
}

// application updates every end block
func (app *mtApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	// distribute rewards for the previous block
	distr.BeginBlocker(ctx, req, app.distrKeeper)

	// slash anyone who double signed.
	// NOTE: This should happen after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool,
	// so as to keep the CanWithdrawInvariant invariant.
	// TODO: This should really happen at EndBlocker.
	tags := slashing.BeginBlocker(ctx, req, app.slashingKeeper)

	return abci.ResponseBeginBlock{
		Tags: tags.ToKVPairs(),
	}
}

// application updates every end block
// nolint: unparam
func (app *mtApp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	tags := gov.EndBlocker(ctx, app.govKeeper)
	validatorUpdates, endBlockerTags := staking.EndBlocker(ctx, app.stakingKeeper)
	tags = append(tags, endBlockerTags...)

	//app.assertRuntimeInvariants()

	return abci.ResponseEndBlock{
		ValidatorUpdates: validatorUpdates,
		Tags:             tags,
	}
}

// initialize store from a genesis state
func (app *mtApp) initFromGenesisState(ctx sdk.Context, genesisState GenesisState) []abci.ValidatorUpdate {
	// load the accounts
	for _, gacc := range genesisState.Accounts {
		acc := gacc.ToAccount()
		acc = app.accountKeeper.NewAccount(ctx, acc) // set account number
		app.accountKeeper.SetAccount(ctx, acc)
		mtacc := gacc.ToMTAccount()
		app.mtKeeper.SetAccountStatus(ctx, acc.GetAddress(), mtacc)
	}

	// initialize distribution (must happen before staking)
	distr.InitGenesis(ctx, app.distrKeeper, genesisState.DistrData)

	// load the initial staking information
	validators, err := staking.InitGenesis(ctx, app.stakingKeeper, genesisState.StakingData)
	if err != nil {
		panic(err) // TODO find a way to do this w/o panics
	}

	// initialize module-specific stores
	auth.InitGenesis(ctx, app.accountKeeper, app.feeCollectionKeeper, genesisState.AuthData)
	bank.InitGenesis(ctx, app.bankKeeper, genesisState.BankData)
	slashing.InitGenesis(ctx, app.slashingKeeper, genesisState.SlashingData, genesisState.StakingData.Validators.ToSDKValidators())
	gov.InitGenesis(ctx, app.govKeeper, genesisState.GovData)

	// validate genesis state
	if err := ValidateGenesisState(genesisState); err != nil {
		panic(err) // TODO find a way to do this w/o panics
	}

	if len(genesisState.GenTxs) > 0 {
		for _, genTx := range genesisState.GenTxs {
			var tx auth.StdTx
			err = app.cdc.UnmarshalJSON(genTx, &tx)
			if err != nil {
				panic(err)
			}
			bz := app.cdc.MustMarshalBinaryLengthPrefixed(tx)
			res := app.BaseApp.DeliverTx(bz)
			if !res.IsOK() {
				panic(res.Log)
			}
		}

		validators = app.stakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	}
	return validators
}

// custom logic for mtd initialization
func (app *mtApp) initChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	stateJSON := req.AppStateBytes

	var genesisState GenesisState
	err := app.cdc.UnmarshalJSON(stateJSON, &genesisState)
	if err != nil {
		panic(err)
	}
	
	validators := app.initFromGenesisState(ctx, genesisState)
	
	// sanity check
	if len(req.Validators) > 0 {
		if len(req.Validators) != len(validators) {
			panic(fmt.Errorf("len(RequestInitChain.Validators) != len(validators) (%d != %d)",
				len(req.Validators), len(validators)))
		}
		sort.Sort(abci.ValidatorUpdates(req.Validators))
		sort.Sort(abci.ValidatorUpdates(validators))
		for i, val := range validators {
			if !val.Equal(req.Validators[i]) {
				panic(fmt.Errorf("validators[%d] != req.Validators[%d] ", i, i))
			}
		}
	}

	microtick.InitGenesis(ctx, app.mtKeeper, genesisState.MicrotickData)
	
	return abci.ResponseInitChain{
		Validators: validators,
	}
}

// GenesisAccount defines an account initialized at genesis.
type GenesisAccount struct {
	Address       sdk.AccAddress `json:"address"`
	NumQuotes	uint32 `json:"numquotes"`
	NumTrades	uint32 `json:"numtrades"`
	Balance		sdk.DecCoin `json:"balance"`
	Sequence      uint64         `json:"sequence_number"`
	AccountNumber uint64         `json:"account_number"`

	// vesting account fields
	OriginalVesting  sdk.Coins `json:"original_vesting"`  // total vesting coins upon initialization
	DelegatedFree    sdk.Coins `json:"delegated_free"`    // delegated vested coins at time of delegation
	DelegatedVesting sdk.Coins `json:"delegated_vesting"` // delegated vesting coins at time of delegation
	StartTime        int64     `json:"start_time"`        // vesting start time (UNIX Epoch time)
	EndTime          int64     `json:"end_time"`          // vesting end time (UNIX Epoch time)
}

func NewGenesisAccount(acc *auth.BaseAccount) GenesisAccount {
	return GenesisAccount{
		Address:       acc.Address,
		NumQuotes:     0,
		NumTrades:     0,
		Balance:       sdk.NewDecCoin(microtick.TokenType, acc.Coins.AmountOf(microtick.TokenType)),
		AccountNumber: acc.AccountNumber,
		Sequence:      acc.Sequence,
	}
}

func NewGenesisAccountI(acc auth.Account) GenesisAccount {
	gacc := GenesisAccount{
		Address:       acc.GetAddress(),
		NumQuotes:     0,
		NumTrades:     0,
		Balance:       sdk.NewDecCoin(microtick.TokenType, acc.GetCoins().AmountOf(microtick.TokenType)),
		AccountNumber: acc.GetAccountNumber(),
		Sequence:      acc.GetSequence(),
	}

	vacc, ok := acc.(auth.VestingAccount)
	if ok {
		gacc.OriginalVesting = vacc.GetOriginalVesting()
		gacc.DelegatedFree = vacc.GetDelegatedFree()
		gacc.DelegatedVesting = vacc.GetDelegatedVesting()
		gacc.StartTime = vacc.GetStartTime()
		gacc.EndTime = vacc.GetEndTime()
	}

	return gacc
}

// convert GenesisAccount to auth.BaseAccount
func (ga *GenesisAccount) ToAccount() auth.Account {
	coin, _ := ga.Balance.TruncateDecimal()
	bacc := &auth.BaseAccount{
		Address:       ga.Address,
		Coins:         sdk.NewCoins(coin),
		AccountNumber: ga.AccountNumber,
		Sequence:      ga.Sequence,
	}

	if !ga.OriginalVesting.IsZero() {
		baseVestingAcc := &auth.BaseVestingAccount{
			BaseAccount:      bacc,
			OriginalVesting:  ga.OriginalVesting,
			DelegatedFree:    ga.DelegatedFree,
			DelegatedVesting: ga.DelegatedVesting,
			EndTime:          ga.EndTime,
		}

		if ga.StartTime != 0 && ga.EndTime != 0 {
			return &auth.ContinuousVestingAccount{
				BaseVestingAccount: baseVestingAcc,
				StartTime:          ga.StartTime,
			}
		} else if ga.EndTime != 0 {
			return &auth.DelayedVestingAccount{
				BaseVestingAccount: baseVestingAcc,
			}
		} else {
			panic(fmt.Sprintf("invalid genesis vesting account: %+v", ga))
		}
	}

	return bacc
}

func (ga *GenesisAccount) ToMTAccount() microtick.DataAccountStatus {
	_, change := ga.Balance.TruncateDecimal()
	return microtick.DataAccountStatus {
		Account: ga.Address,
		ActiveQuotes: microtick.NewOrderedList(),
		ActiveTrades: microtick.NewOrderedList(),
		NumQuotes: ga.NumQuotes,
		NumTrades: ga.NumTrades,
		Change: change,
		QuoteBacking: microtick.NewMicrotickCoinFromInt(0),
		TradeBacking: microtick.NewMicrotickCoinFromInt(0),
		SettleBacking: microtick.NewMicrotickCoinFromInt(0),
	}
}

func ValidateGenesisState(genesisState GenesisState) error {
	if err := validateGenesisStateAccounts(genesisState.Accounts); err != nil {
		return err
	}
	
	// skip stakingData validation as genesis is created from txs
	if len(genesisState.GenTxs) > 0 {
		return nil
	}

	if err := auth.ValidateGenesis(genesisState.AuthData); err != nil {
		return err
	}
	if err := bank.ValidateGenesis(genesisState.BankData); err != nil {
		return err
	}
	if err := staking.ValidateGenesis(genesisState.StakingData); err != nil {
		return err
	}
	if err := distr.ValidateGenesis(genesisState.DistrData); err != nil {
		return err
	}
	if err := gov.ValidateGenesis(genesisState.GovData); err != nil {
		return err
	}

	return slashing.ValidateGenesis(genesisState.SlashingData)
}

// Create the core parameters for genesis initialization
// note that the pubkey input is this machines pubkey
func AppGenState(cdc *codec.Codec, genDoc tmtypes.GenesisDoc, appGenTxs []json.RawMessage) (
	genesisState GenesisState, err error) {

	if err = cdc.UnmarshalJSON(genDoc.AppState, &genesisState); err != nil {
		return genesisState, err
	}

	// if there are no gen txs to be processed, return the default empty state
	if len(appGenTxs) == 0 {
		return genesisState, errors.New("there must be at least one genesis tx")
	}

	stakingData := genesisState.StakingData
	for i, genTx := range appGenTxs {
		var tx auth.StdTx
		if err := cdc.UnmarshalJSON(genTx, &tx); err != nil {
			return genesisState, err
		}

		msgs := tx.GetMsgs()
		if len(msgs) != 1 {
			return genesisState, errors.New(
				"must provide genesis StdTx with exactly 1 CreateValidator message")
		}

		if _, ok := msgs[0].(staking.MsgCreateValidator); !ok {
			return genesisState, fmt.Errorf(
				"Genesis transaction %v does not contain a MsgCreateValidator", i)
		}
	}

	for _, acc := range genesisState.Accounts {
		coin, _ := acc.Balance.TruncateDecimal()
		// Will always be fox
		//if coin.Denom == genesisState.StakingData.Params.BondDenom {
			stakingData.Pool.NotBondedTokens = stakingData.Pool.NotBondedTokens.
				Add(coin.Amount) // increase the supply
		//}
	}

	genesisState.StakingData = stakingData
	genesisState.GenTxs = appGenTxs

	return genesisState, nil
}

// AppGenState but with JSON
func AppGenStateJSON(cdc *codec.Codec, genDoc tmtypes.GenesisDoc, appGenTxs []json.RawMessage) (
	appState json.RawMessage, err error) {
	// create the final app state
	genesisState, err := AppGenState(cdc, genDoc, appGenTxs)
	if err != nil {
		return nil, err
	}
	return codec.MarshalJSONIndent(cdc, genesisState)
}

// validateGenesisStateAccounts performs validation of genesis accounts. It
// ensures that there are no duplicate accounts in the genesis state and any
// provided vesting accounts are valid.
func validateGenesisStateAccounts(accs []GenesisAccount) error {
	addrMap := make(map[string]bool, len(accs))
	for _, acc := range accs {
		addrStr := acc.Address.String()

		// disallow any duplicate accounts
		if _, ok := addrMap[addrStr]; ok {
			return fmt.Errorf("duplicate account found in genesis state; address: %s", addrStr)
		}

		// validate any vesting fields
		if !acc.OriginalVesting.IsZero() {
			if acc.EndTime == 0 {
				return fmt.Errorf("missing end time for vesting account; address: %s", addrStr)
			}

			if acc.StartTime >= acc.EndTime {
				return fmt.Errorf(
					"vesting start time must before end time; address: %s, start: %s, end: %s",
					addrStr,
					time.Unix(acc.StartTime, 0).UTC().Format(time.RFC3339),
					time.Unix(acc.EndTime, 0).UTC().Format(time.RFC3339),
				)
			}
		}

		addrMap[addrStr] = true
	}

	return nil
}

func (app *mtApp) ExportAppStateAndValidators(forZeroHeight bool, jailWhiteList []string) (
	appState json.RawMessage, validators []tmtypes.GenesisValidator, err error) {
		
	ctx := app.NewContext(true, abci.Header{ Height: app.LastBlockHeight() })
	
	if forZeroHeight {
		app.prepForZeroHeightGenesis(ctx, jailWhiteList)
	}
	
	// iterate to get the accounts
	
	accounts := []GenesisAccount{}
	appendAccountsFn := func(acc auth.Account) (stop bool) {
		account := NewGenesisAccountI(acc)
		status := app.mtKeeper.GetAccountStatus(ctx, acc.GetAddress())
		account.Balance = account.Balance.Add(status.Change).Add(status.QuoteBacking).
			Add(status.TradeBacking).Add(status.SettleBacking)
		account.NumQuotes = status.NumQuotes
		account.NumTrades = status.NumTrades
		accounts = append(accounts, account)
		return false
	}
	app.accountKeeper.IterateAccounts(ctx, appendAccountsFn)
	
	genState := NewGenesisState(
		accounts,
		auth.ExportGenesis(ctx, app.accountKeeper, app.feeCollectionKeeper),
		bank.ExportGenesis(ctx, app.bankKeeper),
		staking.ExportGenesis(ctx, app.stakingKeeper),
		distr.ExportGenesis(ctx, app.distrKeeper),
		gov.ExportGenesis(ctx, app.govKeeper),
		slashing.ExportGenesis(ctx, app.slashingKeeper),
		microtick.ExportGenesis(ctx, app.mtKeeper),
	)

	appState, err = codec.MarshalJSONIndent(app.cdc, genState)
	if err != nil {
		return nil, nil, err
	}
	validators = staking.WriteValidators(ctx, app.stakingKeeper)
	return appState, validators, err
}

// prepare for fresh start at zero height
func (app *mtApp) prepForZeroHeightGenesis(ctx sdk.Context, jailWhiteList []string) {
	applyWhiteList := false

	//Check if there is a whitelist
	if len(jailWhiteList) > 0 {
		applyWhiteList = true
	}

	whiteListMap := make(map[string]bool)

	for _, addr := range jailWhiteList {
		_, err := sdk.ValAddressFromBech32(addr)
		if err != nil {
			log.Fatal(err)
		}
		whiteListMap[addr] = true
	}

	/* Just to be safe, assert the invariants on current state. */
	//app.assertRuntimeInvariantsOnContext(ctx)

	/* Handle fee distribution state. */

	// withdraw all validator commission
	app.stakingKeeper.IterateValidators(ctx, func(_ int64, val sdk.Validator) (stop bool) {
		_, _ = app.distrKeeper.WithdrawValidatorCommission(ctx, val.GetOperator())
		return false
	})

	// withdraw all delegator rewards
	dels := app.stakingKeeper.GetAllDelegations(ctx)
	for _, delegation := range dels {
		_, _ = app.distrKeeper.WithdrawDelegationRewards(ctx, delegation.DelegatorAddress, delegation.ValidatorAddress)
	}

	// clear validator slash events
	app.distrKeeper.DeleteAllValidatorSlashEvents(ctx)

	// clear validator historical rewards
	app.distrKeeper.DeleteAllValidatorHistoricalRewards(ctx)

	// set context height to zero
	height := ctx.BlockHeight()
	ctx = ctx.WithBlockHeight(0)

	// reinitialize all validators
	app.stakingKeeper.IterateValidators(ctx, func(_ int64, val sdk.Validator) (stop bool) {

		// donate any unwithdrawn outstanding reward fraction tokens to the community pool
		scraps := app.distrKeeper.GetValidatorOutstandingRewards(ctx, val.GetOperator())
		feePool := app.distrKeeper.GetFeePool(ctx)
		feePool.CommunityPool = feePool.CommunityPool.Add(scraps)
		app.distrKeeper.SetFeePool(ctx, feePool)

		app.distrKeeper.Hooks().AfterValidatorCreated(ctx, val.GetOperator())
		return false
	})

	// reinitialize all delegations
	for _, del := range dels {
		app.distrKeeper.Hooks().BeforeDelegationCreated(ctx, del.DelegatorAddress, del.ValidatorAddress)
		app.distrKeeper.Hooks().AfterDelegationModified(ctx, del.DelegatorAddress, del.ValidatorAddress)
	}

	// reset context height
	ctx = ctx.WithBlockHeight(height)

	/* Handle staking state. */

	// iterate through redelegations, reset creation height
	app.stakingKeeper.IterateRedelegations(ctx, func(_ int64, red staking.Redelegation) (stop bool) {
		for i := range red.Entries {
			red.Entries[i].CreationHeight = 0
		}
		app.stakingKeeper.SetRedelegation(ctx, red)
		return false
	})

	// iterate through unbonding delegations, reset creation height
	app.stakingKeeper.IterateUnbondingDelegations(ctx, func(_ int64, ubd staking.UnbondingDelegation) (stop bool) {
		for i := range ubd.Entries {
			ubd.Entries[i].CreationHeight = 0
		}
		app.stakingKeeper.SetUnbondingDelegation(ctx, ubd)
		return false
	})

	// Iterate through validators by power descending, reset bond heights, and
	// update bond intra-tx counters.
	store := ctx.KVStore(app.keyStaking)
	iter := sdk.KVStoreReversePrefixIterator(store, staking.ValidatorsKey)
	counter := int16(0)

	var valConsAddrs []sdk.ConsAddress
	for ; iter.Valid(); iter.Next() {
		addr := sdk.ValAddress(iter.Key()[1:])
		validator, found := app.stakingKeeper.GetValidator(ctx, addr)
		if !found {
			panic("expected validator, not found")
		}

		validator.UnbondingHeight = 0
		valConsAddrs = append(valConsAddrs, validator.ConsAddress())
		if applyWhiteList && !whiteListMap[addr.String()] {
			validator.Jailed = true
		}

		app.stakingKeeper.SetValidator(ctx, validator)
		counter++
	}

	iter.Close()

	_ = app.stakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)

	/* Handle slashing state. */

	// reset start height on signing infos
	app.slashingKeeper.IterateValidatorSigningInfos(
		ctx,
		func(addr sdk.ConsAddress, info slashing.ValidatorSigningInfo) (stop bool) {
			info.StartHeight = 0
			app.slashingKeeper.SetValidatorSigningInfo(ctx, addr, info)
			return false
		},
	)
}

// MakeCodec generates the necessary codecs for Amino
func MakeCodec() *codec.Codec {
	var cdc = codec.New()
	auth.RegisterCodec(cdc)
	bank.RegisterCodec(cdc)
	staking.RegisterCodec(cdc)
	distr.RegisterCodec(cdc)
	slashing.RegisterCodec(cdc)
	gov.RegisterCodec(cdc)
	microtick.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	return cdc
}

// CollectStdTxs processes and validates application's genesis StdTxs and returns
// the list of appGenTxs, and persistent peers required to generate genesis.json.
func CollectStdTxs(cdc *codec.Codec, moniker string, genTxsDir string, genDoc tmtypes.GenesisDoc) (
	appGenTxs []auth.StdTx, persistentPeers string, err error) {

	var fos []os.FileInfo
	fos, err = ioutil.ReadDir(genTxsDir)
	if err != nil {
		return appGenTxs, persistentPeers, err
	}

	// prepare a map of all accounts in genesis state to then validate
	// against the validators addresses
	var appState GenesisState
	if err := cdc.UnmarshalJSON(genDoc.AppState, &appState); err != nil {
		return appGenTxs, persistentPeers, err
	}

	addrMap := make(map[string]GenesisAccount, len(appState.Accounts))
	for i := 0; i < len(appState.Accounts); i++ {
		acc := appState.Accounts[i]
		addrMap[acc.Address.String()] = acc
	}

	// addresses and IPs (and port) validator server info
	var addressesIPs []string

	for _, fo := range fos {
		filename := filepath.Join(genTxsDir, fo.Name())
		if !fo.IsDir() && (filepath.Ext(filename) != ".json") {
			continue
		}

		// get the genStdTx
		var jsonRawTx []byte
		if jsonRawTx, err = ioutil.ReadFile(filename); err != nil {
			return appGenTxs, persistentPeers, err
		}
		var genStdTx auth.StdTx
		if err = cdc.UnmarshalJSON(jsonRawTx, &genStdTx); err != nil {
			return appGenTxs, persistentPeers, err
		}
		appGenTxs = append(appGenTxs, genStdTx)

		// the memo flag is used to store
		// the ip and node-id, for example this may be:
		// "528fd3df22b31f4969b05652bfe8f0fe921321d5@192.168.2.37:26656"
		nodeAddrIP := genStdTx.GetMemo()
		if len(nodeAddrIP) == 0 {
			return appGenTxs, persistentPeers, fmt.Errorf(
				"couldn't find node's address and IP in %s", fo.Name())
		}

		// genesis transactions must be single-message
		msgs := genStdTx.GetMsgs()
		if len(msgs) != 1 {

			return appGenTxs, persistentPeers, errors.New(
				"each genesis transaction must provide a single genesis message")
		}

		msg := msgs[0].(staking.MsgCreateValidator)
		// validate delegator and validator addresses and funds against the accounts in the state
		delAddr := msg.DelegatorAddress.String()
		valAddr := sdk.AccAddress(msg.ValidatorAddress).String()

		delAcc, delOk := addrMap[delAddr]
		_, valOk := addrMap[valAddr]

		accsNotInGenesis := []string{}
		if !delOk {
			accsNotInGenesis = append(accsNotInGenesis, delAddr)
		}
		if !valOk {
			accsNotInGenesis = append(accsNotInGenesis, valAddr)
		}
		if len(accsNotInGenesis) != 0 {
			return appGenTxs, persistentPeers, fmt.Errorf(
				"account(s) %v not in genesis.json: %+v", strings.Join(accsNotInGenesis, " "), addrMap)
		}

		amount, _ := delAcc.Balance.TruncateDecimal()
		if amount.Amount.LT(msg.Value.Amount) {
			return appGenTxs, persistentPeers, fmt.Errorf(
				"insufficient fund for delegation %v: %v < %v",
				delAcc.Address, amount.Amount, msg.Value.Amount,
			)
		}

		// exclude itself from persistent peers
		if msg.Description.Moniker != moniker {
			addressesIPs = append(addressesIPs, nodeAddrIP)
		}
	}

	sort.Strings(addressesIPs)
	persistentPeers = strings.Join(addressesIPs, ",")

	return appGenTxs, persistentPeers, nil
}

//______________________________________________________________________________________________

// Here we append Microtick specific tags to the MsgSend Tx

type msgSendData struct {
	Time time.Time `json:"time"`
	Originator string `json:"originator"`
	Transfer sdk.Coins `json:"transfer"`
	BalanceFrom microtick.MicrotickCoin `json:"balanceFrom"`
	BalanceTo microtick.MicrotickCoin `json:"balanceTo"`
}

func NewBankHandler(bankKeeper bank.Keeper, mtk microtick.Keeper, cdc *codec.Codec) sdk.Handler {
	bankHandler := bank.NewHandler(bankKeeper)
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		result := bankHandler(ctx, msg)
		switch msg := msg.(type) {
		case bank.MsgSend:
			fromBalance := mtk.GetTotalBalance(ctx, msg.FromAddress)
			toBalance := mtk.GetTotalBalance(ctx, msg.ToAddress)
			tags := sdk.NewTags(
				fmt.Sprintf("acct.%s", msg.FromAddress), "account.withdraw",
				fmt.Sprintf("acct.%s", msg.ToAddress), "account.deposit",
			)
			result.Tags = result.Tags.AppendTags(tags)
			data := msgSendData {
				Time: ctx.BlockHeader().Time,
				Originator: "send",
				Transfer: msg.Amount,
				BalanceFrom: fromBalance,
				BalanceTo: toBalance,
			}
			bz, _ := codec.MarshalJSONIndent(cdc, data)
			result.Data = bz
		}
		return result
	}
}

// load a particular height
func (app *mtApp) LoadHeight(height int64) error {
	return app.LoadVersion(height, app.keyMain)
}

//______________________________________________________________________________________________

var _ sdk.StakingHooks = StakingHooks{}

// StakingHooks contains combined distribution and slashing hooks needed for the
// staking module.
type StakingHooks struct {
	dh distr.Hooks
	sh slashing.Hooks
}

func NewStakingHooks(dh distr.Hooks, sh slashing.Hooks) StakingHooks {
	return StakingHooks{dh, sh}
}

// nolint
func (h StakingHooks) AfterValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress) {
	h.dh.AfterValidatorCreated(ctx, valAddr)
	h.sh.AfterValidatorCreated(ctx, valAddr)
}
func (h StakingHooks) BeforeValidatorModified(ctx sdk.Context, valAddr sdk.ValAddress) {
	h.dh.BeforeValidatorModified(ctx, valAddr)
	h.sh.BeforeValidatorModified(ctx, valAddr)
}
func (h StakingHooks) AfterValidatorRemoved(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
	h.dh.AfterValidatorRemoved(ctx, consAddr, valAddr)
	h.sh.AfterValidatorRemoved(ctx, consAddr, valAddr)
}
func (h StakingHooks) AfterValidatorBonded(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
	h.dh.AfterValidatorBonded(ctx, consAddr, valAddr)
	h.sh.AfterValidatorBonded(ctx, consAddr, valAddr)
}
func (h StakingHooks) AfterValidatorBeginUnbonding(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) {
	h.dh.AfterValidatorBeginUnbonding(ctx, consAddr, valAddr)
	h.sh.AfterValidatorBeginUnbonding(ctx, consAddr, valAddr)
}
func (h StakingHooks) BeforeDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	h.dh.BeforeDelegationCreated(ctx, delAddr, valAddr)
	h.sh.BeforeDelegationCreated(ctx, delAddr, valAddr)
}
func (h StakingHooks) BeforeDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	h.dh.BeforeDelegationSharesModified(ctx, delAddr, valAddr)
	h.sh.BeforeDelegationSharesModified(ctx, delAddr, valAddr)
}
func (h StakingHooks) BeforeDelegationRemoved(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	h.dh.BeforeDelegationRemoved(ctx, delAddr, valAddr)
	h.sh.BeforeDelegationRemoved(ctx, delAddr, valAddr)
}
func (h StakingHooks) AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) {
	h.dh.AfterDelegationModified(ctx, delAddr, valAddr)
	h.sh.AfterDelegationModified(ctx, delAddr, valAddr)
}
func (h StakingHooks) BeforeValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec) {
	h.dh.BeforeValidatorSlashed(ctx, valAddr, fraction)
	h.sh.BeforeValidatorSlashed(ctx, valAddr, fraction)
}
