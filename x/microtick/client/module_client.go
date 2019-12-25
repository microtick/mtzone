package client

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(moduleName string, cdc *codec.Codec) *cobra.Command {
	mtQueryCmd := &cobra.Command{
		Use:   "microtick",
		Short: "Querying commands for the microtick module",
	}

	mtQueryCmd.AddCommand(client.GetCommands(
		GetCmdAccountStatus(moduleName, cdc),
		GetCmdMarketStatus(moduleName, cdc),
		GetCmdMarketConsensus(moduleName, cdc),
		GetCmdOrderBook(moduleName, cdc),
		GetCmdActiveQuote(moduleName, cdc),
		GetCmdActiveTrade(moduleName, cdc),
	)...)

	return mtQueryCmd
}

func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	mtTxCmd := &cobra.Command{
		Use:   "microtick",
		Short: "Microtick transactions subcommands",
	}

	mtTxCmd.AddCommand(client.PostCommands(
		GetCmdCreateMarket(cdc),
		GetCmdCreateQuote(cdc),
		GetCmdCancelQuote(cdc),
		GetCmdUpdateQuote(cdc),
		GetCmdDepositQuote(cdc),
		GetCmdWithdrawQuote(cdc),
		GetCmdMarketTrade(cdc),
		GetCmdLimitTrade(cdc),
		GetCmdSettleTrade(cdc),
	)...)

	return mtTxCmd
}
