package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/mjackson001/mtzone/x/microtick"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
)

func GetCmdCreateMarket(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "create-market [market]",
		Short: "create a new market",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)

			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			err := cliCtx.EnsureAccountExists()
			if err != nil {
				return err
			}

			var market = args[0]

			msg := microtick.NewTxCreateMarket(cliCtx.GetFromAddress(), market)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			cliCtx.PrintResponse = true

			return utils.CompleteAndBroadcastTxCLI(txBldr, cliCtx, []sdk.Msg{msg})
		},
	}
}

func GetCmdCreateQuote(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "create-quote [market] [duration] [backing] [spot] [premium]",
		Short: "create a new quote",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)

			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			if err := cliCtx.EnsureAccountExists(); err != nil {
				return err
			}
			
			market := args[0]
			
			dur, err := microtick.NewMicrotickDurationFromString(args[1])
			if err != nil {
				return err
			}

			coins, err2 := microtick.NewMicrotickCoinFromString(args[2])
			if err2 != nil {
				return err2
			}
			
			spot, err := microtick.NewMicrotickSpotFromString(args[3])
			if err != nil {
				return err
			}
			
			premium, err := microtick.NewMicrotickPremiumFromString(args[4])
			if err != nil {
				return err
			}

			msg := microtick.NewTxCreateQuote(market, dur, cliCtx.GetFromAddress(), coins,
				spot, premium)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			cliCtx.PrintResponse = true

			return utils.CompleteAndBroadcastTxCLI(txBldr, cliCtx, []sdk.Msg{msg})
		},
	}
}
