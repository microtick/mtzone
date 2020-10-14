package microtick

import (
	"encoding/json"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

    cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/gogo/protobuf/grpc"

	mt "gitlab.com/microtick/mtzone/x/microtick/types"
	"gitlab.com/microtick/mtzone/x/microtick/msg"
	"gitlab.com/microtick/mtzone/x/microtick/keeper"
	"gitlab.com/microtick/mtzone/x/microtick/client/cli"
	"gitlab.com/microtick/mtzone/x/microtick/client/rest"
	
	abci "github.com/tendermint/tendermint/abci/types"	
)

const (
    ModuleName = "microtick"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// app module basics object
type AppModuleBasic struct{
	cdc codec.Marshaler
}

// module name
func (AppModuleBasic) Name() string {
	return ModuleName
}

func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	msg.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module's interface types
func (b AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	msg.RegisterInterfaces(registry)
}


// default genesis state
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONMarshaler) json.RawMessage {
	defaultGenesis := mt.DefaultGenesisState()
	return cdc.MustMarshalJSON(&defaultGenesis)
}

// module validate genesis
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONMarshaler, config client.TxEncodingConfig, bz json.RawMessage) error {
	var data mt.GenesisMicrotick
	err := cdc.UnmarshalJSON(bz, &data)
	if err != nil {
		return sdkerrors.Wrapf(err, "failed to unmarshal %s genesis state", ModuleName)
	}
	return mt.ValidateGenesis(data)
}

// register rest routes
func (AppModuleBasic) RegisterRESTRoutes(ctx client.Context, rtr *mux.Router) {
	rest.RegisterRoutes(ctx, rtr)
}

func (AppModuleBasic) RegisterGRPCRoutes(ctx client.Context, mux *runtime.ServeMux) {
}

// get the root query command of this module
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// get the root tx command of this module
func (AppModuleBasic) GetTxCmd() *cobra.Command { 
	return cli.GetTxCmd(ModuleName)
}

//___________________________
// app module
type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Marshaler, keeper keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         keeper,
	}
}

// module name
func (AppModule) Name() string {
	return ModuleName
}

// register invariants
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {}

// module message route name
func (am AppModule) Route() sdk.Route { 
	return sdk.NewRoute(ModuleName, NewHandler(am.keeper))
}

// module handler
func (am AppModule) NewHandler() sdk.Handler { 
	return NewHandler(am.keeper)
}

// module querier route name
func (AppModule) QuerierRoute() string {
	return ModuleName
}

func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
  return msg.NewQuerier(am.keeper)
}

// RegisterQueryService registers a GRPC query service to respond to the
// module-specific GRPC queries.
func (am AppModule) RegisterQueryService(server grpc.Server) {
	querier := msg.Querier{Keeper: am.keeper}
	msg.RegisterGRPCServer(server, querier)
}

// module init-genesis
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONMarshaler, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState mt.GenesisMicrotick
	cdc.MustUnmarshalJSON(data, &genesisState)
	InitGenesis(ctx, am.keeper, genesisState)
	return []abci.ValidatorUpdate{}
}

// module export genesis
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONMarshaler) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper)
	return cdc.MustMarshalJSON(&gs)
}

// module begin-block
func (am AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {}

// module end-block
func (am AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	EndBlocker(ctx, am.keeper)
	return []abci.ValidatorUpdate{}
}
