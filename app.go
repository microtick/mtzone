package app

import (
	"io"
	"os"

	tlog "github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/cosmos/cosmos-sdk/x/params"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/genaccounts"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	"github.com/cosmos/cosmos-sdk/x/supply"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tm-db"
	
	"github.com/mjackson001/mtzone/x/microtick"
	"github.com/mjackson001/mtzone/x/microtick/keeper"
)

const (
    appName = "Microtick"
    BondDenom = "fox"
    
    mtGlobalsKey = "MTGlobals"
    mtAccountStatusKey = "MTAccountStatus"
    mtActiveQuotesKey = "MTActiveQuotes"
    mtActiveTradesKey = "MTActiveTrades"
    mtMarketsKey = "MTMarkets"
)

var (
	DefaultCLIHome = os.ExpandEnv("$MTROOT/mtcli")
	DefaultNodeHome = os.ExpandEnv("$MTROOT/mtd")
	ModuleBasics = module.NewBasicManager(
		genaccounts.AppModuleBasic{},
		genutil.AppModuleBasic{},
		auth.AppModuleBasic{},
		bank.AppModuleBasic{},
		staking.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic(paramsclient.ProposalHandler, distr.ProposalHandler),
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},
		supply.AppModuleBasic{},
	)
	maccPerms = map[string][]string{
		auth.FeeCollectorName: nil,
		distr.ModuleName: nil,
		staking.BondedPoolName: {supply.Burner, supply.Staking},
		staking.NotBondedPoolName: {supply.Burner, supply.Staking},
		gov.ModuleName: {supply.Burner},
	}
)

func init() {
	sdk.PowerReduction = sdk.NewInt(1)
}

type mtApp struct {
    *bam.BaseApp
	cdc *codec.Codec
	
	invCheckPeriod uint
	
	keys map[string]*sdk.KVStoreKey
	tkeys map[string]*sdk.TransientStoreKey

	accountKeeper       auth.AccountKeeper
	bankKeeper          bank.Keeper
	//feeCollectionKeeper auth.FeeCollectionKeeper
	supplyKeeper		supply.Keeper
	stakingKeeper		staking.Keeper
	slashingKeeper		slashing.Keeper
	distrKeeper			distr.Keeper
	govKeeper			gov.Keeper
	crisisKeeper		crisis.Keeper
	paramsKeeper        params.Keeper
	mtKeeper            keeper.MicrotickKeeper
	
	// module manager
	mm	*module.Manager
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *mtApp) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[supply.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

func NewMTApp(
	logger tlog.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool, 
	invCheckPeriod uint, baseAppOptions ...func(*bam.BaseApp),
) *mtApp {
		
    cdc := MakeCodec()

    bApp := bam.NewBaseApp(appName, logger, db, auth.DefaultTxDecoder(cdc), baseAppOptions...)
    bApp.SetCommitMultiStoreTracer(traceStore)
    bApp.SetAppVersion(MTAppVersion + " " + MTBuildDate + " @" + MTHostBuild)
    
    keys := sdk.NewKVStoreKeys(
    	bam.MainStoreKey,
    	auth.StoreKey,
    	staking.StoreKey,
    	supply.StoreKey,
    	distr.StoreKey,
    	slashing.StoreKey,
    	gov.StoreKey,
    	params.StoreKey,
    	mtGlobalsKey,
    	mtAccountStatusKey,
    	mtActiveQuotesKey,
    	mtActiveTradesKey,
    	mtMarketsKey,
    )
    
    tkeys := sdk.NewTransientStoreKeys(
    	staking.TStoreKey,
    	params.TStoreKey,
    )
    
    app := &mtApp{
        BaseApp: bApp,
        cdc:     cdc,
        
        invCheckPeriod:   invCheckPeriod,
        
        keys: keys,
        tkeys: tkeys,
    }

 	// The ParamsKeeper handles parameter storage for the application
	app.paramsKeeper = params.NewKeeper(
		app.cdc, 
		keys[params.StoreKey], 
		tkeys[params.TStoreKey], 
		params.DefaultCodespace)
		
	authSubspace := app.paramsKeeper.Subspace(auth.DefaultParamspace)	
		
	app.accountKeeper = auth.NewAccountKeeper(
		app.cdc,
		keys[auth.StoreKey],
		authSubspace,
		auth.ProtoBaseAccount,
	)
	
	app.bankKeeper = bank.NewBaseKeeper(
		app.accountKeeper,
		app.paramsKeeper.Subspace(bank.DefaultParamspace),
		bank.DefaultCodespace,
		app.ModuleAccountAddrs(),
	)
	
	app.supplyKeeper = supply.NewKeeper(
		app.cdc,
		keys[supply.StoreKey],
		app.accountKeeper,
		app.bankKeeper,
		maccPerms,
	)
	
	stakingSubspace := app.paramsKeeper.Subspace(staking.DefaultParamspace)
		
	stakingKeeper := staking.NewKeeper(
		app.cdc,
		keys[staking.StoreKey],
		tkeys[staking.TStoreKey],
		app.supplyKeeper,
		stakingSubspace,
		slashing.DefaultCodespace,
	)
	
	distrSubspace := app.paramsKeeper.Subspace(distr.DefaultParamspace)
	
	app.distrKeeper = distr.NewKeeper(
		app.cdc,
		keys[distr.StoreKey],
		distrSubspace,
		&stakingKeeper,
		app.supplyKeeper,
		distr.DefaultCodespace,
		auth.FeeCollectorName,
		app.ModuleAccountAddrs(),
	)
	
	slashingSubspace := app.paramsKeeper.Subspace(slashing.DefaultParamspace)
	
	app.slashingKeeper = slashing.NewKeeper(
		app.cdc, 
		keys[slashing.StoreKey],
		&stakingKeeper,
		slashingSubspace,
		slashing.DefaultCodespace,
	)
	
	crisisSubspace := app.paramsKeeper.Subspace(crisis.DefaultParamspace)
	
	app.crisisKeeper = crisis.NewKeeper(
		crisisSubspace,
		invCheckPeriod,
		app.supplyKeeper,
		auth.FeeCollectorName,
	)
	
	govSubspace := app.paramsKeeper.Subspace(gov.DefaultParamspace)
	
	govRouter := gov.NewRouter()
	govRouter.AddRoute(gov.RouterKey, gov.ProposalHandler).
		AddRoute(params.RouterKey, params.NewParamChangeProposalHandler(app.paramsKeeper)).
		AddRoute(distr.RouterKey, distr.NewCommunityPoolSpendProposalHandler(app.distrKeeper))
	app.govKeeper = gov.NewKeeper(
		app.cdc, 
		keys[gov.StoreKey], 
		app.paramsKeeper,
		govSubspace,
		app.supplyKeeper, 
		&stakingKeeper, 
		gov.DefaultCodespace, 
		govRouter,
	)
	
	app.mtKeeper = keeper.NewKeeper(
		app.accountKeeper,
		app.bankKeeper,
		app.distrKeeper,
		app.stakingKeeper,
		keys[mtGlobalsKey],
		keys[mtAccountStatusKey],
		keys[mtActiveQuotesKey],
		keys[mtActiveTradesKey],
		keys[mtMarketsKey],
		app.paramsKeeper.Subspace(microtick.DefaultParamspace),
		app.cdc,
	)
	
	app.stakingKeeper = *stakingKeeper.SetHooks(
		staking.NewMultiStakingHooks(
			app.distrKeeper.Hooks(),
			app.slashingKeeper.Hooks(),
		),
	)
	
	app.mm = module.NewManager(
		genaccounts.NewAppModule(app.accountKeeper),
		genutil.NewAppModule(app.accountKeeper, app.stakingKeeper, app.BaseApp.DeliverTx),
		auth.NewAppModule(app.accountKeeper),
		bank.NewAppModule(app.bankKeeper, app.accountKeeper),
		crisis.NewAppModule(&app.crisisKeeper),
		supply.NewAppModule(app.supplyKeeper, app.accountKeeper),
		distr.NewAppModule(app.distrKeeper, app.supplyKeeper),
		gov.NewAppModule(app.govKeeper, app.supplyKeeper),
		slashing.NewAppModule(app.slashingKeeper, app.stakingKeeper),
		staking.NewAppModule(app.stakingKeeper, app.distrKeeper, app.accountKeeper, app.supplyKeeper),
		microtick.NewAppModule(app.mtKeeper),
	)
	
	app.mm.SetOrderInitGenesis(
		genaccounts.ModuleName, distr.ModuleName, staking.ModuleName,
		auth.ModuleName, bank.ModuleName, slashing.ModuleName, gov.ModuleName,
		supply.ModuleName, crisis.ModuleName, genutil.ModuleName,
	)
	
	// register the crisis routes
	app.mm.RegisterInvariants(&app.crisisKeeper)
	app.mm.RegisterRoutes(app.Router(), app.QueryRouter())

	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)
	
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetAnteHandler(auth.NewAnteHandler(app.accountKeeper, app.supplyKeeper, auth.DefaultSigVerificationGasConsumer))
	app.SetEndBlocker(app.EndBlocker)

	if loadLatest {
		err := app.LoadLatestVersion(app.keys[bam.MainStoreKey])
		if err != nil {
			cmn.Exit(err.Error())
		}
	}

    return app
}

// application updates every end block
func (app *mtApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.mm.BeginBlock(ctx, req)
}

// application updates every end block
// nolint: unparam
func (app *mtApp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.mm.EndBlock(ctx, req)
}

func (app *mtApp) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState simapp.GenesisState
	app.cdc.MustUnmarshalJSON(req.AppStateBytes, &genesisState)
	
	return app.mm.InitGenesis(ctx, genesisState)
}

// MakeCodec generates the necessary codecs for Amino
func MakeCodec() *codec.Codec {
	var cdc = codec.New()
	
	ModuleBasics.RegisterCodec(cdc)
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	codec.RegisterEvidences(cdc)
	
	return cdc
}

func (app *mtApp) LoadHeight(height int64) error {
	return app.LoadVersion(height, app.keys[bam.MainStoreKey])
}
