package app

import (
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrclient "github.com/cosmos/cosmos-sdk/x/distribution/client"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	"github.com/cosmos/cosmos-sdk/x/capability"
	"github.com/cosmos/cosmos-sdk/x/ibc"
	transfer "github.com/cosmos/cosmos-sdk/x/ibc-transfer"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradeclient "github.com/cosmos/cosmos-sdk/x/upgrade/client"
	
	appparams "gitlab.com/microtick/mtzone/app/params"
	"gitlab.com/microtick/mtzone/x/microtick"
)

var (
	mbasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		genutil.AppModuleBasic{},
		bank.AppModuleBasic{},
		capability.AppModuleBasic{},
		mint.AppModuleBasic{},
		staking.AppModuleBasic{},
		slashing.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic(
			paramsclient.ProposalHandler, distrclient.ProposalHandler,
			upgradeclient.ProposalHandler, upgradeclient.CancelProposalHandler,
		),
		params.AppModuleBasic{},
		ibc.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		crisis.AppModuleBasic{},
		transfer.AppModuleBasic{},
		microtick.AppModuleBasic{},
	)
)

// ModuleBasics returns all app modules basics
func ModuleBasics() module.BasicManager {
	return mbasics
}

// MakeEncodingConfig creates an EncodingConfig for testing
func MakeEncodingConfig() appparams.EncodingConfig {
	encodingConfig := appparams.MakeEncodingConfig()
	std.RegisterLegacyAminoCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	mbasics.RegisterLegacyAminoCodec(encodingConfig.Amino)
	mbasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	return encodingConfig
}
