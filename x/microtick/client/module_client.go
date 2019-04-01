package client

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
	amino "github.com/tendermint/go-amino"
	
    "github.com/mjackson001/mtzone/x/microtick/client/cli"
)

// ModuleClient exports all client functionality from this module
type ModuleClient struct {
	storeKey string
	cdc      *amino.Codec
}

func NewModuleClient(storeKey string, cdc *amino.Codec) ModuleClient {
	return ModuleClient{storeKey, cdc}
}

// GetQueryCmd returns the cli query commands for this module
func (mc ModuleClient) GetQueryCmd() *cobra.Command {
	namesvcQueryCmd := &cobra.Command{
		Use:   "microtick",
		Short: "Querying commands for the microtick module",
	}

	namesvcQueryCmd.AddCommand(client.GetCommands(
		cli.GetCmdAccountStatus(mc.storeKey, mc.cdc),
		cli.GetCmdAccountActive(mc.storeKey, mc.cdc),
		cli.GetCmdMarketStatus(mc.storeKey, mc.cdc),
	)...)

	return namesvcQueryCmd
}

func (mc ModuleClient) GetTxCmd() *cobra.Command {
	mtTxCmd := &cobra.Command{
		Use:   "microtick",
		Short: "Microtick transactions subcommands",
	}

	mtTxCmd.AddCommand(client.PostCommands(
		cli.GetCmdCreateMarket(mc.cdc),
		cli.GetCmdCreateQuote(mc.cdc),
	)...)

	return mtTxCmd
}
