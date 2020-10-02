package cli

import (
    "fmt"
    
    "github.com/cosmos/cosmos-sdk/codec"
    "github.com/cosmos/cosmos-sdk/client/flags"
    "github.com/cosmos/cosmos-sdk/client/context"
    "github.com/spf13/cobra"
    
    "gitlab.com/microtick/mtzone/x/microtick/msg"
)

func GetQueryCmd(moduleName string, cdc *codec.Codec) *cobra.Command {
	mtQueryCmd := &cobra.Command{
		Use:   "microtick",
		Short: "Querying commands for the microtick module",
	}

	mtQueryCmd.AddCommand(flags.GetCommands(
		GetCmdAccountStatus(moduleName, cdc),
		GetCmdMarketStatus(moduleName, cdc),
		GetCmdMarketConsensus(moduleName, cdc),
		GetCmdOrderBook(moduleName, cdc),
		GetCmdActiveQuote(moduleName, cdc),
		GetCmdActiveTrade(moduleName, cdc),
	)...)

	return mtQueryCmd
}

func GetCmdAccountStatus(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "account [acct]",
		Short: "Query account full details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			acct := args[0]

			res, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/account/%s", queryRoute, acct))
			if err != nil {
				fmt.Printf("No such account: %s \n", string(acct))
				return nil
			}

			var out msg.ResponseAccountStatus
			cdc.MustUnmarshalJSON(res, &out)
			
			if cliCtx.OutputFormat == "text" {
				fmt.Println(out.String())
			} else {
				cliCtx.PrintOutput(out)
			}
			return nil
		},
	}
}

func GetCmdMarketStatus(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "market [name]",
		Short: "Query market status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			market := args[0]

			res, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/market/%s", queryRoute, market))
			if err != nil {
				fmt.Printf("No such market: %s \n", string(market))
				return nil
			}

			var out msg.ResponseMarketStatus
			cdc.MustUnmarshalJSON(res, &out)
			
			if cliCtx.OutputFormat == "text" {
				fmt.Println(out.String())
			} else {
				cliCtx.PrintOutput(out)
			}
			return nil
		},
	}
}

func GetCmdMarketConsensus(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "consensus [market]",
		Short: "Query market consensus",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			market := args[0]

			res, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/consensus/%s", queryRoute, market))
			if err != nil {
				fmt.Printf("No such market: %s \n", string(market))
				return nil
			}

			var out msg.ResponseMarketConsensus
			cdc.MustUnmarshalJSON(res, &out)
			
			if cliCtx.OutputFormat == "text" {
				fmt.Println(out.String())
			} else {
				cliCtx.PrintOutput(out)
			}
			return nil
		},
	}
}

func GetCmdOrderBook(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "orderbook [market] [dur]",
		Short: "Query market orderbook",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			market := args[0]
			dur := args[1]

			res, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/orderbook/%s/%s", queryRoute, market, dur))
			if err != nil {
				fmt.Printf("No such orderbook: %s %s\n", market, dur)
				return nil
			}

			var out msg.ResponseOrderBook
			cdc.MustUnmarshalJSON(res, &out)
			
			if cliCtx.OutputFormat == "text" {
				fmt.Println(out.String())
			} else {
				cliCtx.PrintOutput(out)
			}
			return nil
		},
	}
}

func GetCmdActiveQuote(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "quote [id]",
		Short: "Query quote",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			id := args[0]

			res, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/quote/%s", queryRoute, id))
			if err != nil {
				fmt.Printf("No such quote: %s \n", string(id))
				return nil
			}

			var out msg.ResponseQuoteStatus
			cdc.MustUnmarshalJSON(res, &out)
			
			if cliCtx.OutputFormat == "text" {
				fmt.Println(out.String())
			} else {
				cliCtx.PrintOutput(out)
			}
			return nil
		},
	}
}

func GetCmdActiveTrade(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "trade [id]",
		Short: "Query trade",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			id := args[0]

			res, _, err := cliCtx.Query(fmt.Sprintf("custom/%s/trade/%s", queryRoute, id))
			if err != nil {
				fmt.Printf("No such trade: %s \n", string(id))
				return nil
			}

			var out msg.ResponseTradeStatus
			cdc.MustUnmarshalJSON(res, &out)
			
			if cliCtx.OutputFormat == "text" {
				fmt.Println(out.String())
			} else {
				cliCtx.PrintOutput(out)
			}
			return nil
		},
	}
}
