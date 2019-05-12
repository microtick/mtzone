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

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
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
			dur := microtick.MicrotickDurationFromName(args[1])
			coins := microtick.NewMicrotickCoinFromString(args[2])
			spot := microtick.NewMicrotickSpotFromString(args[3])
			premium := microtick.NewMicrotickPremiumFromString(args[4])

			msg := microtick.NewTxCreateQuote(market, dur, cliCtx.GetFromAddress(), coins,
				spot, premium)
			err := msg.ValidateBasic()
			if err != nil {
				return err
			}

			cliCtx.PrintResponse = true

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}
}

func GetCmdCancelQuote(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "cancel-quote [id]",
		Short: "cancel a quote",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)

			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			if err := cliCtx.EnsureAccountExists(); err != nil {
				return err
			}
			
			id := microtick.NewMicrotickIdFromString(args[0])

			msg := microtick.NewTxCancelQuote(id, cliCtx.GetFromAddress())
			err := msg.ValidateBasic()
			if err != nil {
				return err
			}

			cliCtx.PrintResponse = true

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}
}

func GetCmdUpdateQuote(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "update-quote [id] [newspot] [newpremium]",
		Short: "update a quote",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)

			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			if err := cliCtx.EnsureAccountExists(); err != nil {
				return err
			}
			
			id := microtick.NewMicrotickIdFromString(args[0])
			newspot := microtick.NewMicrotickSpotFromString(args[1])
			newpremium := microtick.NewMicrotickPremiumFromString(args[2])

			msg := microtick.NewTxUpdateQuote(id, cliCtx.GetFromAddress(), newspot, newpremium)
			err := msg.ValidateBasic()
			if err != nil {
				return err
			}

			cliCtx.PrintResponse = true

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}
}

func GetCmdDepositQuote(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "deposit-quote [id] [amount]",
		Short: "deposit more backing to a quote",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)

			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			if err := cliCtx.EnsureAccountExists(); err != nil {
				return err
			}
			
			id := microtick.NewMicrotickIdFromString(args[0])
			deposit := microtick.NewMicrotickCoinFromString(args[1])

			msg := microtick.NewTxDepositQuote(id, cliCtx.GetFromAddress(), deposit)
			err := msg.ValidateBasic()
			if err != nil {
				return err
			}

			cliCtx.PrintResponse = true

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}
}

func GetCmdMarketTrade(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "trade-market [market] [duration] [call/put] [quantity]",
		Short: "create a new trade",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)

			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			if err := cliCtx.EnsureAccountExists(); err != nil {
				return err
			}
			
			market := args[0]
			
			dur := microtick.MicrotickDurationFromName(args[1])
			ttype := microtick.MicrotickTradeTypeFromName(args[2])
			quantity := microtick.NewMicrotickQuantityFromString(args[3])
			
			msg := microtick.NewTxMarketTrade(market, dur, cliCtx.GetFromAddress(), ttype,
				quantity)
			err := msg.ValidateBasic()
			if err != nil {
				return err
			}

			cliCtx.PrintResponse = true

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}
}

func GetCmdLimitTrade(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "trade-limit [market] [duration] [call/put] [limit] [maxcost]",
		Short: "create a new limit trade",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)

			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			if err := cliCtx.EnsureAccountExists(); err != nil {
				return err
			}
			
			market := args[0]
			
			dur := microtick.MicrotickDurationFromName(args[1])
			ttype := microtick.MicrotickTradeTypeFromName(args[2])
			limit := microtick.NewMicrotickPremiumFromString(args[3])
			maxcost := microtick.NewMicrotickCoinFromString(args[4])
			
			msg := microtick.NewTxLimitTrade(market, dur, cliCtx.GetFromAddress(), ttype,
				limit, maxcost)
			err := msg.ValidateBasic()
			if err != nil {
				return err
			}

			cliCtx.PrintResponse = true

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}
}

func GetCmdSettleTrade(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "settle [id]",
		Short: "settle trade",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc).WithAccountDecoder(cdc)

			txBldr := authtxb.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))

			if err := cliCtx.EnsureAccountExists(); err != nil {
				return err
			}
			
			id := microtick.NewMicrotickIdFromString(args[0])
			
			msg := microtick.NewTxSettleTrade(id, cliCtx.GetFromAddress())
			err := msg.ValidateBasic()
			if err != nil {
				return err
			}

			cliCtx.PrintResponse = true

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg}, false)
		},
	}
}
