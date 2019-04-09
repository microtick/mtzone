package cli

import (
    "fmt"
    
    "github.com/cosmos/cosmos-sdk/codec"
    "github.com/cosmos/cosmos-sdk/client/context"
    "github.com/spf13/cobra"
    "github.com/mjackson001/mtzone/x/microtick"
)

func GetCmdAccountStatus(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "account [acct]",
		Short: "account acct",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			acct := args[0]

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/account/%s", queryRoute, acct), nil)
			if err != nil {
				fmt.Printf("No such account: %s \n", string(acct))
				return nil
			}

			var out microtick.ResponseAccountStatus
			cdc.MustUnmarshalJSON(res, &out)
			//fmt.Println(out.String())
			return cliCtx.PrintOutput(out)
		},
	}
}

func GetCmdMarketStatus(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "market [name]",
		Short: "market name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			market := args[0]

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/market/%s", queryRoute, market), nil)
			if err != nil {
				fmt.Printf("No such market: %s \n", string(market))
				return nil
			}

			var out microtick.ResponseMarketStatus
			cdc.MustUnmarshalJSON(res, &out)
			//fmt.Println(out.String())
			return cliCtx.PrintOutput(out)
		},
	}
}

func GetCmdMarketConsensus(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "consensus [market]",
		Short: "consensus market",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			market := args[0]

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/consensus/%s", queryRoute, market), nil)
			if err != nil {
				fmt.Printf("No such market: %s \n", string(market))
				return nil
			}

			var out microtick.ResponseMarketConsensus
			cdc.MustUnmarshalJSON(res, &out)
			//fmt.Println(out.String())
			return cliCtx.PrintOutput(out)
		},
	}
}

func GetCmdOrderBook(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "orderbook [market] [dur]",
		Short: "orderbook market dur",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			market := args[0]
			dur := args[1]

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/orderbook/%s/%s", queryRoute, market, dur), nil)
			if err != nil {
				fmt.Printf("No such orderbook: %s %s\n", market, dur)
				return nil
			}

			var out microtick.ResponseOrderBook
			cdc.MustUnmarshalJSON(res, &out)
			//fmt.Println(out.String())
			return cliCtx.PrintOutput(out)
		},
	}
}

func GetCmdActiveQuote(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "quote [id]",
		Short: "quote id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			id := args[0]

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/quote/%s", queryRoute, id), nil)
			if err != nil {
				fmt.Printf("No such quote: %s \n", string(id))
				return nil
			}

			var out microtick.ResponseQuoteStatus
			cdc.MustUnmarshalJSON(res, &out)
			//fmt.Println(out.String())
			return cliCtx.PrintOutput(out)
		},
	}
}

func GetCmdActiveTrade(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "trade [id]",
		Short: "trade id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			id := args[0]

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/trade/%s", queryRoute, id), nil)
			if err != nil {
				fmt.Printf("No such trade: %s \n", string(id))
				return nil
			}

			var out microtick.ResponseTradeStatus
			cdc.MustUnmarshalJSON(res, &out)
			//fmt.Println(out.String())
			return cliCtx.PrintOutput(out)
		},
	}
}
