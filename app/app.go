package app

import (
	"fmt"
	"io"
	"os"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/capability"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/ibc"
	//ibcclient "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	//port "github.com/cosmos/cosmos-sdk/x/ibc/05-port"
	transfer "github.com/cosmos/cosmos-sdk/x/ibc-transfer"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradeclient "github.com/cosmos/cosmos-sdk/x/upgrade/client"
	
	"gitlab.com/microtick/mtzone/x/microtick"
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
		auth.AppModuleBasic{},
		genutil.AppModuleBasic{},
		bank.AppModuleBasic{},
		capability.AppModuleBasic{},
		staking.AppModuleBasic{},
		mint.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic(
			paramsclient.ProposalHandler, distr.ProposalHandler, upgradeclient.ProposalHandler,
		),
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},
		//ibc.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		//transfer.AppModuleBasic{},
		microtick.AppModuleBasic{},
	)

	// module account permissions
	maccPerms = map[string][]string{
        auth.FeeCollectorName:           nil,
        distr.ModuleName:                nil,
        mint.ModuleName:                 {auth.Minter},
        staking.BondedPoolName:          {auth.Burner, auth.Staking},
        staking.NotBondedPoolName:       {auth.Burner, auth.Staking},
        gov.ModuleName:                  {auth.Burner},
        //transfer.ModuleName:           {auth.Minter, auth.Burner},
        microtick.ModuleName:	         {auth.Minter, auth.Burner},
	}
	
	// module accounts that are allowed to receive tokens
	allowedReceivingModAcc = map[string]bool{
		distr.ModuleName: true,
	}
)

// Extended ABCI application
type MTApp struct {
	*baseapp.BaseApp
	cdc *codec.Codec

	invCheckPeriod uint

	// keys to access the substores
	keys    map[string]*sdk.KVStoreKey
	tkeys   map[string]*sdk.TransientStoreKey
	memKeys map[string]*sdk.MemoryStoreKey
	
	// subspaces
	subspaces map[string]params.Subspace

	// keepers
	accountKeeper    auth.AccountKeeper
	bankKeeper       bank.Keeper
	capabilityKeeper *capability.Keeper
	stakingKeeper    staking.Keeper
	slashingKeeper   slashing.Keeper
	mintKeeper		   mint.Keeper
	distrKeeper      distr.Keeper
	govKeeper        gov.Keeper
	crisisKeeper     crisis.Keeper
	upgradeKeeper    upgrade.Keeper
	paramsKeeper     params.Keeper
	ibcKeeper        *ibc.Keeper
	evidenceKeeper   evidence.Keeper
	transferKeeper   transfer.Keeper
	mtKeeper         microtick.Keeper
	
	// make scoped keepers public for test purposes
	scopedIBCKeeper      capability.ScopedKeeper
	scopedTransferKeeper capability.ScopedKeeper

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
func NewMTApp(
	logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool,
	invCheckPeriod uint, skipUpgradeHeights map[int64]bool, home string, 
	baseAppOptions ...func(*baseapp.BaseApp),
) *MTApp {
		
	appCodec, cdc := MakeCodecs()

	bApp := baseapp.NewBaseApp(appName, logger, db, auth.DefaultTxDecoder(cdc), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)

	keys := sdk.NewKVStoreKeys(
		auth.StoreKey, bank.StoreKey, staking.StoreKey,
		mint.StoreKey, distr.StoreKey, slashing.StoreKey,
		gov.StoreKey, params.StoreKey, ibc.StoreKey, upgrade.StoreKey,
		evidence.StoreKey, transfer.StoreKey, capability.StoreKey,
		microtick.GlobalsKey,
		microtick.AccountStatusKey,
		microtick.ActiveQuotesKey,
		microtick.ActiveTradesKey,
		microtick.MarketsKey,
		microtick.DurationsKey,
	)
	tkeys := sdk.NewTransientStoreKeys(params.TStoreKey)
	memKeys := sdk.NewMemoryStoreKeys(capability.MemStoreKey)

	app := &MTApp{
		BaseApp:        bApp,
		cdc:            cdc,
		invCheckPeriod: invCheckPeriod,
		keys:           keys,
		tkeys:          tkeys,
		memKeys:		    memKeys,
		subspaces:		  make(map[string]params.Subspace),
	}

	// init params keeper and subspaces
	app.paramsKeeper = params.NewKeeper(appCodec, keys[params.StoreKey], tkeys[params.TStoreKey])
	app.subspaces[auth.ModuleName] = app.paramsKeeper.Subspace(auth.DefaultParamspace)
	app.subspaces[bank.ModuleName] = app.paramsKeeper.Subspace(bank.DefaultParamspace)
	app.subspaces[staking.ModuleName] = app.paramsKeeper.Subspace(staking.DefaultParamspace)
	app.subspaces[mint.ModuleName] = app.paramsKeeper.Subspace(mint.DefaultParamspace)
	app.subspaces[distr.ModuleName] = app.paramsKeeper.Subspace(distr.DefaultParamspace)
	app.subspaces[slashing.ModuleName] = app.paramsKeeper.Subspace(slashing.DefaultParamspace)
	app.subspaces[gov.ModuleName] = app.paramsKeeper.Subspace(gov.DefaultParamspace).WithKeyTable(gov.ParamKeyTable())
	app.subspaces[crisis.ModuleName] = app.paramsKeeper.Subspace(crisis.DefaultParamspace)
	app.subspaces[microtick.ModuleName] = app.paramsKeeper.Subspace(microtick.DefaultParamspace)
	
	bApp.SetParamStore(app.paramsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(std.ConsensusParamsKeyTable()))
	
	// add capability keeper and ScopeToModule for ibc module
	app.capabilityKeeper = capability.NewKeeper(appCodec, keys[capability.StoreKey], memKeys[capability.MemStoreKey])
	scopedIBCKeeper := app.capabilityKeeper.ScopeToModule(ibc.ModuleName)
	scopedTransferKeeper := app.capabilityKeeper.ScopeToModule(transfer.ModuleName)	

	// add keepers
	app.accountKeeper = auth.NewAccountKeeper(
		appCodec, keys[auth.StoreKey], app.subspaces[auth.ModuleName], auth.ProtoBaseAccount, maccPerms,
	)
	app.bankKeeper = bank.NewBaseKeeper(
		appCodec, keys[bank.StoreKey], app.accountKeeper, app.subspaces[bank.ModuleName], app.BlacklistedAccAddrs(),
	)
	stakingKeeper := staking.NewKeeper(
		appCodec, keys[staking.StoreKey], app.accountKeeper, app.bankKeeper, app.subspaces[staking.ModuleName],
	)
	app.mintKeeper = mint.NewKeeper(
		appCodec, keys[mint.StoreKey], app.subspaces[mint.ModuleName], &stakingKeeper, 
		app.accountKeeper, app.bankKeeper, auth.FeeCollectorName,
	)
	app.distrKeeper = distr.NewKeeper(
		appCodec, keys[distr.StoreKey], app.subspaces[distr.ModuleName], app.accountKeeper, app.bankKeeper, 
		&stakingKeeper, auth.FeeCollectorName, app.ModuleAccountAddrs(),
	)
	app.slashingKeeper = slashing.NewKeeper(
		appCodec, keys[slashing.StoreKey], &stakingKeeper, app.subspaces[slashing.ModuleName],
	)
	app.crisisKeeper = crisis.NewKeeper(
		app.subspaces[crisis.ModuleName], invCheckPeriod, app.bankKeeper, auth.FeeCollectorName,
	)
	app.upgradeKeeper = upgrade.NewKeeper(skipUpgradeHeights, keys[upgrade.StoreKey], appCodec, home)
	
	// register the proposal types
	govRouter := gov.NewRouter()
	govRouter.AddRoute(gov.RouterKey, gov.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(app.paramsKeeper)).
		AddRoute(distr.RouterKey, distr.NewCommunityPoolSpendProposalHandler(app.distrKeeper)).
		AddRoute(upgrade.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(app.upgradeKeeper))
	app.govKeeper = gov.NewKeeper(
		appCodec, keys[gov.StoreKey], app.subspaces[gov.ModuleName], app.accountKeeper, app.bankKeeper,
		&stakingKeeper, govRouter,
	)	

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	app.stakingKeeper = *stakingKeeper.SetHooks(
		staking.NewMultiStakingHooks(app.distrKeeper.Hooks(), app.slashingKeeper.Hooks()),
	)

  // Create IBC Keeper
	app.ibcKeeper = ibc.NewKeeper(
		app.cdc, appCodec, keys[ibc.StoreKey], stakingKeeper, scopedIBCKeeper,
	)
	
	// Create Transfer Keepers
	//app.transferKeeper = transfer.NewKeeper(
		//appCodec, keys[transfer.StoreKey],
		//app.ibcKeeper.ChannelKeeper, &app.ibcKeeper.PortKeeper,
		//app.accountKeeper, app.bankKeeper,
		//scopedTransferKeeper,
	//)
	//transferModule := transfer.NewAppModule(app.transferKeeper)

	// Create static IBC router, add transfer route, then set and seal it
	//ibcRouter := port.NewRouter()
	//ibcRouter.AddRoute(transfer.ModuleName, transferModule)
	//app.ibcKeeper.SetRouter(ibcRouter)

	// create evidence keeper with router
	evidenceKeeper := evidence.NewKeeper(
		appCodec, keys[evidence.StoreKey], &app.stakingKeeper, app.slashingKeeper,
	)
	evidenceRouter := evidence.NewRouter()
    //.AddRoute(ibcclient.RouterKey, ibcclient.HandlerClientMisbehaviour(app.ibcKeeper.ClientKeeper))

	evidenceKeeper.SetRouter(evidenceRouter)
	app.evidenceKeeper = *evidenceKeeper
	
	app.mtKeeper = microtick.NewKeeper(
		appCodec,
		app.accountKeeper, app.bankKeeper, app.distrKeeper, app.stakingKeeper,
		keys[microtick.GlobalsKey],
		keys[microtick.AccountStatusKey],
		keys[microtick.ActiveQuotesKey],
		keys[microtick.ActiveTradesKey],
		keys[microtick.MarketsKey],
		keys[microtick.DurationsKey],
		app.subspaces[microtick.ModuleName],
	)

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.
	app.mm = module.NewManager(
		microtick.NewAppModule(app.mtKeeper),
		genutil.NewAppModule(app.accountKeeper, app.stakingKeeper, app.BaseApp.DeliverTx),
		auth.NewAppModule(appCodec, app.accountKeeper),
		bank.NewAppModule(appCodec, app.bankKeeper, app.accountKeeper),
		capability.NewAppModule(appCodec, *app.capabilityKeeper),
		crisis.NewAppModule(&app.crisisKeeper),
		gov.NewAppModule(appCodec, app.govKeeper, app.accountKeeper, app.bankKeeper),
		mint.NewAppModule(appCodec, app.mintKeeper, app.accountKeeper),
		slashing.NewAppModule(appCodec, app.slashingKeeper, app.accountKeeper, app.bankKeeper, app.stakingKeeper),
		distr.NewAppModule(appCodec, app.distrKeeper, app.accountKeeper, app.bankKeeper, app.stakingKeeper),
		staking.NewAppModule(appCodec, app.stakingKeeper, app.accountKeeper, app.bankKeeper),
		upgrade.NewAppModule(app.upgradeKeeper),
		evidence.NewAppModule(app.evidenceKeeper),
		//ibc.NewAppModule(app.ibcKeeper),
		params.NewAppModule(app.paramsKeeper),
		//transferModule,
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	app.mm.SetOrderBeginBlockers(
		upgrade.ModuleName, mint.ModuleName, distr.ModuleName, slashing.ModuleName, 
		evidence.ModuleName, staking.ModuleName, //ibc.ModuleName,
	)

	app.mm.SetOrderEndBlockers(
		crisis.ModuleName, gov.ModuleName, staking.ModuleName,
		microtick.ModuleName,
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	app.mm.SetOrderInitGenesis(
		microtick.ModuleName,
		capability.ModuleName, auth.ModuleName, distr.ModuleName, staking.ModuleName, bank.ModuleName,
		slashing.ModuleName, gov.ModuleName, mint.ModuleName, crisis.ModuleName,
		//ibc.ModuleName, genutil.ModuleName, evidence.ModuleName, transfer.ModuleName,
		genutil.ModuleName, evidence.ModuleName,
	)
	
	//app.distrKeeper.SetBaseProposerReward(ctx, sdk.ZeroDec())

	app.mm.RegisterInvariants(&app.crisisKeeper)
	app.mm.RegisterRoutes(app.Router(), app.QueryRouter())

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)
	app.MountMemoryStores(memKeys)

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetAnteHandler(
		auth.NewAnteHandler(
			app.accountKeeper, app.bankKeeper, *app.ibcKeeper, 
			ante.DefaultSigVerificationGasConsumer,
		),
	)
	app.SetEndBlocker(app.EndBlocker)

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			tmos.Exit(err.Error())
		}
	
	  // Initialize and seal the capability keeper so all persistent capabilities
	  // are loaded in-memory and prevent any further modules from creating scoped
	  // sub-keepers.
	  // This must be done during creation of baseapp rather than in InitChain so
	  // that in-memory capabilities get regenerated on app restart
	  ctx := app.BaseApp.NewUncachedContext(true, abci.Header{})
	  app.capabilityKeeper.InitializeAndSeal(ctx)
	}

	app.scopedIBCKeeper = scopedIBCKeeper
	app.scopedTransferKeeper = scopedTransferKeeper

	return app
}

func (app *MTApp) Name() string { return app.BaseApp.Name() }

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
	return app.mm.InitGenesis(ctx, app.cdc, genesisState)
}

// load a particular height
func (app *MTApp) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *MTApp) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[auth.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// BlacklistedAccAddrs returns all the app's module account addresses black listed for receiving tokens.
func (app *MTApp) BlacklistedAccAddrs() map[string]bool {
	blacklistedAddrs := make(map[string]bool)
	for acc := range maccPerms {
		blacklistedAddrs[auth.NewModuleAddress(acc).String()] = !allowedReceivingModAcc[acc]
	}

	return blacklistedAddrs
}

// Codec returns MTApp's codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *MTApp) Codec() *codec.Codec {
	return app.cdc
}

// MakeCodecs constructs the *std.Codec and *codec.Codec instances used by
// GaiaApp.
func MakeCodecs() (*std.Codec, *codec.Codec) {
	cdc := std.MakeCodec(ModuleBasics)
	interfaceRegistry := cdctypes.NewInterfaceRegistry()
	appCodec := std.NewAppCodec(cdc, interfaceRegistry)

	sdk.RegisterInterfaces(interfaceRegistry)
	ModuleBasics.RegisterInterfaceModules(interfaceRegistry)

	return appCodec, cdc
}

// GetMaccPerms returns a copy of the module account permissions
func GetMaccPerms() map[string][]string {
	dupMaccPerms := make(map[string][]string)
	for k, v := range maccPerms {
		dupMaccPerms[k] = v
	}
	return dupMaccPerms
}
