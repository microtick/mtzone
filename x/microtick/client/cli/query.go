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

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/status/%s", queryRoute, acct), nil)
			if err != nil {
				fmt.Printf("could not resolve account - %s \n", string(acct))
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
		Use:   "market [acct]",
		Short: "market acct",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			market := args[0]

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/market/%s", queryRoute, market), nil)
			if err != nil {
				fmt.Printf("could not resolve market - %s \n", string(market))
				return nil
			}

			var out microtick.ResponseMarketStatus
			cdc.MustUnmarshalJSON(res, &out)
			//fmt.Println(out.String())
			return cliCtx.PrintOutput(out)
		},
	}
}