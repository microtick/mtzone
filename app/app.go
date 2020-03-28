package app

import (
	"fmt"
	"io"
	"os"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"
	dbm "github.com/tendermint/tm-db"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codecstd "github.com/cosmos/cosmos-sdk/codec/std"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/ibc"
	transfer "github.com/cosmos/cosmos-sdk/x/ibc/20-transfer"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramsproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradeclient "github.com/cosmos/cosmos-sdk/x/upgrade/client"
	
	"github.com/mjackson001/mtzone/x/microtick"
)

const appName = "MicrotickApp"

var (
	// default home directories for mtcli
	DefaultCLIHome = ""

	// default home directories for mtd
	DefaultNodeHome = ""

	// The module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration
	// and genesis verification.
	ModuleBasics = module.NewBasicManager(
		genutil.AppModuleBasic{},
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
		staking.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic(paramsclient.ProposalHandler, distr.ProposalHandler, upgradeclient.ProposalHandler),
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},
		supply.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		ibc.AppModuleBasic{},
		transfer.AppModuleBasic{},
		microtick.AppModuleBasic{},
	)

	// module account permissions
	maccPerms = map[string][]string{
		auth.FeeCollectorName:     nil,
		distr.ModuleName:          nil,
		staking.BondedPoolName:    {supply.Burner, supply.Staking},
		staking.NotBondedPoolName: {supply.Burner, supply.Staking},
		gov.ModuleName:            {supply.Burner},
		transfer.GetModuleAccountName(): {supply.Minter, supply.Burner},
		microtick.ModuleName:	   {supply.Minter, supply.Burner},
	}
)

// custom tx codec
func MakeCodec() *codec.Codec {
	var cdc = codec.New()

	ModuleBasics.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	codec.RegisterEvidences(cdc)

	return cdc
}

// Extended ABCI application
type MTApp struct {
	*bam.BaseApp
	cdc *codec.Codec

	invCheckPeriod uint

	// keys to access the substores
	keys  map[string]*sdk.KVStoreKey
	tkeys map[string]*sdk.TransientStoreKey
	
	// subspaces
	subspaces map[string]params.Subspace

	// keepers
	accountKeeper  auth.AccountKeeper
	bankKeeper     bank.Keeper
	supplyKeeper   supply.Keeper
	stakingKeeper  staking.Keeper
	slashingKeeper slashing.Keeper
	distrKeeper    distr.Keeper
	govKeeper      gov.Keeper
	crisisKeeper   crisis.Keeper
	paramsKeeper   params.Keeper
	upgradeKeeper  upgrade.Keeper
	evidenceKeeper evidence.Keeper
	ibcKeeper      ibc.Keeper
	transferKeeper transfer.Keeper
	mtKeeper       microtick.Keeper

	// the module manager
	mm *module.Manager
}

func SetAppVersion() {
	// Check for MTROOT environment set
	mtroot := os.Getenv("MTROOT")
	if mtroot == "" {
		mtroot = fmt.Sprintf("%s/.microtick", os.Getenv("HOME"))
	} else {
		// Print custom MTROOT on stderr
		fmt.Fprintf(os.Stderr, "MTROOT set to %s\n", mtroot)
	}
	//fmt.Fprintf(os.Stderr, "Using MTROOT=%s\n", mtroot)
	DefaultNodeHome = fmt.Sprintf("%s/mtd", mtroot)
	DefaultCLIHome = fmt.Sprintf("%s/mtcli", mtroot)
	
	// Check MTROOT version.lock file for correct version, if not, print a warning
	if _, err := os.Stat(mtroot); os.IsNotExist(err) {
		os.Mkdir(mtroot, os.ModePerm)
	}
	filename := fmt.Sprintf("%s/version.lock", mtroot)
	versionRead, err := os.Open(filename)
	if err != nil {
		versionWrite, err := os.Create(filename)
		if err != nil {
			panic(err)
		}
		defer func() {
			if err := versionWrite.Close(); err != nil {
				panic(err)
			}
		}()
		fmt.Fprint(versionWrite, MTAppVersion)
	} else {
		// Check version matches
		defer func() {
			if err := versionRead.Close(); err != nil {
				panic(err)
			}
		}()
		var ver string
		fmt.Fscan(versionRead, &ver)
		if ver != MTAppVersion {
			fmt.Fprintf(os.Stderr, "\nVersion mismatch\n")
			fmt.Fprintf(os.Stderr, "Executable version: %s\n", MTAppVersion)
			fmt.Fprintf(os.Stderr, "Version lock: %s\n\n", ver)
			fmt.Fprintf(os.Stderr, "This warning exists to make sure the Microtick executables are using data and config files " +
			    "generated with correct settings for the correct software version.\n\n")
			fmt.Fprintf(os.Stderr, "(remove this warning by deleting %s/version.lock or using a different root directory by " +
			    "setting the MTROOT environment variable. MTROOT is currently set to %s)\n\n", mtroot, mtroot)
			os.Exit(1)
		}
	}
	
	version.Name = "Microtick"
	version.ServerName = "mtd"
	version.ClientName = "mtcli"
	version.Version = MTAppVersion
	version.Commit = MTCommit
	version.BuildTags = fmt.Sprintf("build_host=%s;build_date=%s", 
	  MTHostBuild,
	  MTBuildDate)
}

// NewMTApp returns a reference to an initialized MTApp.
func NewMTApp(logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool,
	invCheckPeriod uint, skipUpgradeHeights map[int64]bool, home string, 
	baseAppOptions ...func(*bam.BaseApp)) *MTApp {
		
	cdc := codecstd.MakeCodec(ModuleBasics)
	appCodec := codecstd.NewAppCodec(cdc)

	bApp := bam.NewBaseApp(appName, logger, db, auth.DefaultTxDecoder(cdc), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)

	keys := sdk.NewKVStoreKeys(
		bam.MainStoreKey, auth.StoreKey, bank.StoreKey, staking.StoreKey,
		supply.StoreKey, distr.StoreKey, slashing.StoreKey,
		gov.StoreKey, params.StoreKey, ibc.StoreKey, 
		transfer.StoreKey, evidence.StoreKey, upgrade.StoreKey,
		microtick.GlobalsKey,
		microtick.AccountStatusKey,
		microtick.ActiveQuotesKey,
		microtick.ActiveTradesKey,
		microtick.MarketsKey,
	)
	tkeys := sdk.NewTransientStoreKeys(staking.TStoreKey, params.TStoreKey)

	app := &MTApp{
		BaseApp:        bApp,
		cdc:            cdc,
		invCheckPeriod: invCheckPeriod,
		keys:           keys,
		tkeys:          tkeys,
		subspaces:			make(map[string]params.Subspace),
	}

	// init params keeper and subspaces
	app.paramsKeeper = params.NewKeeper(appCodec, keys[params.StoreKey], tkeys[params.TStoreKey])
	app.subspaces[auth.ModuleName] = app.paramsKeeper.Subspace(auth.DefaultParamspace)
	app.subspaces[bank.ModuleName] = app.paramsKeeper.Subspace(bank.DefaultParamspace)
	app.subspaces[staking.ModuleName] = app.paramsKeeper.Subspace(staking.DefaultParamspace)
	app.subspaces[distr.ModuleName] = app.paramsKeeper.Subspace(distr.DefaultParamspace)
	app.subspaces[slashing.ModuleName] = app.paramsKeeper.Subspace(slashing.DefaultParamspace)
	app.subspaces[gov.ModuleName] = app.paramsKeeper.Subspace(gov.DefaultParamspace).WithKeyTable(gov.ParamKeyTable())
	app.subspaces[crisis.ModuleName] = app.paramsKeeper.Subspace(crisis.DefaultParamspace)
	app.subspaces[evidence.ModuleName] = app.paramsKeeper.Subspace(evidence.DefaultParamspace)
	app.subspaces[microtick.ModuleName] = app.paramsKeeper.Subspace(microtick.DefaultParamspace)

	// add keepers
	app.accountKeeper = auth.NewAccountKeeper(
		appCodec, keys[auth.StoreKey], app.subspaces[auth.ModuleName], auth.ProtoBaseAccount,
	)
	app.bankKeeper = bank.NewBaseKeeper(
		appCodec, keys[bank.StoreKey], app.accountKeeper, app.subspaces[bank.ModuleName], app.ModuleAccountAddrs(),
	)
	app.supplyKeeper = supply.NewKeeper(
		appCodec, keys[supply.StoreKey], app.accountKeeper, app.bankKeeper, maccPerms,
	)
	stakingKeeper := staking.NewKeeper(
		appCodec, keys[staking.StoreKey], app.bankKeeper, app.supplyKeeper, app.subspaces[staking.ModuleName],
	)
	app.distrKeeper = distr.NewKeeper(
		appCodec, keys[distr.StoreKey], app.subspaces[distr.ModuleName], app.bankKeeper, &stakingKeeper, 
		app.supplyKeeper, auth.FeeCollectorName, app.ModuleAccountAddrs(),
	)
	app.slashingKeeper = slashing.NewKeeper(
		appCodec, keys[slashing.StoreKey], &stakingKeeper, app.subspaces[slashing.ModuleName],
	)

	app.crisisKeeper = crisis.NewKeeper(
		app.subspaces[crisis.ModuleName], invCheckPeriod, app.supplyKeeper, auth.FeeCollectorName,
	)
	app.upgradeKeeper = upgrade.NewKeeper(skipUpgradeHeights, keys[upgrade.StoreKey], appCodec, home)

	// create evidence keeper with evidence router
	evidenceKeeper := evidence.NewKeeper(
		appCodec, keys[evidence.StoreKey], app.subspaces[evidence.ModuleName], &stakingKeeper, app.slashingKeeper,
	)
	evidenceRouter := evidence.NewRouter()

	// TODO: register evidence routes
	evidenceKeeper.SetRouter(evidenceRouter)

	app.evidenceKeeper = *evidenceKeeper

	// register the proposal types
	govRouter := gov.NewRouter()
	govRouter.AddRoute(gov.RouterKey, gov.ProposalHandler).
		AddRoute(paramsproposal.RouterKey, params.NewParamChangeProposalHandler(app.paramsKeeper)).
		AddRoute(distr.RouterKey, distr.NewCommunityPoolSpendProposalHandler(app.distrKeeper)).
		AddRoute(upgrade.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(app.upgradeKeeper))
	app.govKeeper = gov.NewKeeper(
		appCodec, keys[gov.StoreKey], app.subspaces[gov.ModuleName],
		app.supplyKeeper, &stakingKeeper, govRouter,
	)

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	app.stakingKeeper = *stakingKeeper.SetHooks(
		staking.NewMultiStakingHooks(app.distrKeeper.Hooks(), app.slashingKeeper.Hooks()),
	)

	app.ibcKeeper = ibc.NewKeeper(app.cdc, keys[ibc.StoreKey], stakingKeeper)

	transferCapKey := app.ibcKeeper.PortKeeper.BindPort(bank.ModuleName)
	app.transferKeeper = transfer.NewKeeper(
		app.cdc, keys[transfer.StoreKey], transferCapKey,
		app.ibcKeeper.ChannelKeeper, app.bankKeeper, app.supplyKeeper,
	)
	
	app.mtKeeper = microtick.NewKeeper(
		app.accountKeeper, app.bankKeeper, app.distrKeeper, app.stakingKeeper,
		app.supplyKeeper,
		keys[microtick.GlobalsKey],
		keys[microtick.AccountStatusKey],
		keys[microtick.ActiveQuotesKey],
		keys[microtick.ActiveTradesKey],
		keys[microtick.MarketsKey],
		app.subspaces[microtick.ModuleName], app.cdc,
	)

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.
	app.mm = module.NewManager(
		microtick.NewAppModule(app.mtKeeper),
		genutil.NewAppModule(app.accountKeeper, app.stakingKeeper, app.BaseApp.DeliverTx),
		auth.NewAppModule(app.accountKeeper, app.supplyKeeper),
		bank.NewAppModule(app.bankKeeper, app.accountKeeper),
		crisis.NewAppModule(&app.crisisKeeper),
		supply.NewAppModule(app.supplyKeeper, app.bankKeeper, app.accountKeeper),
		distr.NewAppModule(app.distrKeeper, app.accountKeeper, app.bankKeeper, app.supplyKeeper, app.stakingKeeper),
		gov.NewAppModule(app.govKeeper, app.accountKeeper, app.bankKeeper, app.supplyKeeper),
		slashing.NewAppModule(app.slashingKeeper, app.accountKeeper, app.bankKeeper, app.stakingKeeper),
		staking.NewAppModule(app.stakingKeeper, app.accountKeeper, app.bankKeeper, app.supplyKeeper),
		upgrade.NewAppModule(app.upgradeKeeper),
		evidence.NewAppModule(app.evidenceKeeper),
		ibc.NewAppModule(app.ibcKeeper),
		transfer.NewAppModule(app.transferKeeper),
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	app.mm.SetOrderBeginBlockers(upgrade.ModuleName, distr.ModuleName, slashing.ModuleName, staking.ModuleName)

	app.mm.SetOrderEndBlockers(crisis.ModuleName, gov.ModuleName, staking.ModuleName,
		microtick.ModuleName)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	app.mm.SetOrderInitGenesis(
		microtick.ModuleName,
		distr.ModuleName, staking.ModuleName,
		auth.ModuleName, bank.ModuleName, slashing.ModuleName, gov.ModuleName,
		supply.ModuleName, crisis.ModuleName, genutil.ModuleName, evidence.ModuleName,
	)
	
	//app.distrKeeper.SetBaseProposerReward(ctx, sdk.ZeroDec())

	app.mm.RegisterInvariants(&app.crisisKeeper)
	app.mm.RegisterRoutes(app.Router(), app.QueryRouter())

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetAnteHandler(auth.NewAnteHandler(app.accountKeeper, app.supplyKeeper, app.ibcKeeper, auth.DefaultSigVerificationGasConsumer))
	app.SetEndBlocker(app.EndBlocker)

	if loadLatest {
		err := app.LoadLatestVersion(app.keys[bam.MainStoreKey])
		if err != nil {
			tmos.Exit(err.Error())
		}
	}

	return app
}

// application updates every begin block
func (app *MTApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.mm.BeginBlock(ctx, req)
}

// application updates every end block
func (app *MTApp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.mm.EndBlock(ctx, req)
}

// application update at chain initialization
func (app *MTApp) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState simapp.GenesisState
	app.cdc.MustUnmarshalJSON(req.AppStateBytes, &genesisState)

	res := app.mm.InitGenesis(ctx, app.cdc, genesisState)
	
	// Set Historical infos in InitChain to ignore genesis params
	stakingParams := staking.DefaultParams()
	stakingParams.HistoricalEntries = 1000
	app.stakingKeeper.SetParams(ctx, stakingParams)

	return res
}

// load a particular height
func (app *MTApp) LoadHeight(height int64) error {
	return app.LoadVersion(height, app.keys[bam.MainStoreKey])
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *MTApp) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[supply.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}
