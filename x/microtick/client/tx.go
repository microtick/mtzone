package client

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	
	mt "github.com/mjackson001/mtzone/x/microtick/types"
	"github.com/mjackson001/mtzone/x/microtick/msg"
)

func GetCmdCreateMarket(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "create-market [market]",
		Short: "Create a new market",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			var market = args[0]

			txmsg := msg.NewTxCreateMarket(cliCtx.GetFromAddress(), market)
			err := txmsg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{txmsg})
		},
	}
}

func GetCmdCreateQuote(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "create-quote [market] [duration] [backing] [spot] [premium]",
		Short: "Create a new quote",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			market := args[0]
			dur := mt.MicrotickDurationFromName(args[1])
			coins := mt.NewMicrotickCoinFromString(args[2])
			spot := mt.NewMicrotickSpotFromString(args[3])
			premium := mt.NewMicrotickPremiumFromString(args[4])

			msg := msg.NewTxCreateQuote(market, dur, cliCtx.GetFromAddress(), coins,
				spot, premium)
			err := msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdCancelQuote(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "cancel-quote [id]",
		Short: "Cancel a quote",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			id := mt.NewMicrotickIdFromString(args[0])

			txmsg := msg.NewTxCancelQuote(id, cliCtx.GetFromAddress())
			err := txmsg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{txmsg})
		},
	}
}

func GetCmdUpdateQuote(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "update-quote [id] [newspot] [newpremium]",
		Short: "Update a quote",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			id := mt.NewMicrotickIdFromString(args[0])
			newspot := mt.NewMicrotickSpotFromString(args[1])
			newpremium := mt.NewMicrotickPremiumFromString(args[2])

			txmsg := msg.NewTxUpdateQuote(id, cliCtx.GetFromAddress(), newspot, newpremium)
			err := txmsg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{txmsg})
		},
	}
}

func GetCmdDepositQuote(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "deposit-quote [id] [amount]",
		Short: "Deposit more backing to a quote",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			id := mt.NewMicrotickIdFromString(args[0])
			deposit := mt.NewMicrotickCoinFromString(args[1])

			txmsg := msg.NewTxDepositQuote(id, cliCtx.GetFromAddress(), deposit)
			err := txmsg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{txmsg})
		},
	}
}

func GetCmdWithdrawQuote(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "withdraw-quote [id] [amount]",
		Short: "Withdraw backing from a quote",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			id := mt.NewMicrotickIdFromString(args[0])
			withdraw := mt.NewMicrotickCoinFromString(args[1])

			txmsg := msg.NewTxWithdrawQuote(id, cliCtx.GetFromAddress(), withdraw)
			err := txmsg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{txmsg})
		},
	}
}

func GetCmdMarketTrade(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "trade-market [market] [duration] [call/put] [quantity]",
		Short: "Create a new trade",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			market := args[0]
			
			dur := mt.MicrotickDurationFromName(args[1])
			ttype := mt.MicrotickTradeTypeFromName(args[2])
			quantity := mt.NewMicrotickQuantityFromString(args[3])
			
			txmsg := msg.NewTxMarketTrade(market, dur, cliCtx.GetFromAddress(), ttype,
				quantity)
			err := txmsg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{txmsg})
		},
	}
}

func GetCmdLimitTrade(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "trade-limit [market] [duration] [call/put] [limit] [maxcost]",
		Short: "Create a new limit trade",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			market := args[0]
			
			dur := mt.MicrotickDurationFromName(args[1])
			ttype := mt.MicrotickTradeTypeFromName(args[2])
			limit := mt.NewMicrotickPremiumFromString(args[3])
			maxcost := mt.NewMicrotickCoinFromString(args[4])
			
			txmsg := msg.NewTxLimitTrade(market, dur, cliCtx.GetFromAddress(), ttype,
				limit, maxcost)
			err := txmsg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{txmsg})
		},
	}
}

func GetCmdSettleTrade(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "settle [id]",
		Short: "Settle trade",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			txBldr := auth.NewTxBuilderFromCLI().WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			id := mt.NewMicrotickIdFromString(args[0])
			
			txmsg := msg.NewTxSettleTrade(id, cliCtx.GetFromAddress())
			err := txmsg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{txmsg})
		},
	}
}
