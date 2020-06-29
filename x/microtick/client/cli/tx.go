package cli

import (
	"bufio"
	
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	
	mt "github.com/mjackson001/mtzone/x/microtick/types"
	"github.com/mjackson001/mtzone/x/microtick/msg"
)

func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	mtTxCmd := &cobra.Command{
		Use:   "microtick",
		Short: "Microtick transactions subcommands",
	}

	mtTxCmd.AddCommand(flags.PostCommands(
		GetCmdQuoteCancel(cdc),
		GetCmdQuoteCreate(cdc),
		GetCmdQuoteDeposit(cdc),
		GetCmdQuoteUpdate(cdc),
		GetCmdQuoteWithdraw(cdc),
		GetCmdTradeMarket(cdc),
		GetCmdTradeLimit(cdc),
		GetCmdTradePick(cdc),
		GetCmdTradeSettle(cdc),
	)...)

	return mtTxCmd
}

func GetCmdQuoteCancel(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "quote-cancel [id]",
		Short: "Cancel a quote",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			id := mt.NewMicrotickIdFromString(args[0])

			txmsg := msg.NewTxCancelQuote(id, cliCtx.GetFromAddress())
			err := txmsg.ValidateBasic()
			if err != nil {
				return err
			}

			return authclient.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{txmsg})
		},
	}
}

func GetCmdQuoteCreate(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "quote-create [market] [duration] [backing] [spot] [premium]",
		Short: "Create a new quote",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			market := args[0]
			dur := args[1]
			coins := mt.NewMicrotickCoinFromString(args[2])
			spot := mt.NewMicrotickSpotFromString(args[3])
			premium := mt.NewMicrotickPremiumFromString(args[4])

			msg := msg.NewTxCreateQuote(market, dur, cliCtx.GetFromAddress(), coins,
				spot, premium)
			err := msg.ValidateBasic()
			if err != nil {
				return err
			}

			return authclient.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

func GetCmdQuoteDeposit(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "quote-deposit [id] [amount]",
		Short: "Deposit more backing to a quote",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			id := mt.NewMicrotickIdFromString(args[0])
			deposit := mt.NewMicrotickCoinFromString(args[1])

			txmsg := msg.NewTxDepositQuote(id, cliCtx.GetFromAddress(), deposit)
			err := txmsg.ValidateBasic()
			if err != nil {
				return err
			}

			return authclient.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{txmsg})
		},
	}
}

func GetCmdQuoteUpdate(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "quote-update [id] [newspot] [newpremium]",
		Short: "Update a quote",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			id := mt.NewMicrotickIdFromString(args[0])
			newspot := mt.NewMicrotickSpotFromString(args[1])
			newpremium := mt.NewMicrotickPremiumFromString(args[2])

			txmsg := msg.NewTxUpdateQuote(id, cliCtx.GetFromAddress(), newspot, newpremium)
			err := txmsg.ValidateBasic()
			if err != nil {
				return err
			}

			return authclient.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{txmsg})
		},
	}
}

func GetCmdQuoteWithdraw(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "quote-withdraw [id] [amount]",
		Short: "Withdraw backing from a quote",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			id := mt.NewMicrotickIdFromString(args[0])
			withdraw := mt.NewMicrotickCoinFromString(args[1])

			txmsg := msg.NewTxWithdrawQuote(id, cliCtx.GetFromAddress(), withdraw)
			err := txmsg.ValidateBasic()
			if err != nil {
				return err
			}

			return authclient.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{txmsg})
		},
	}
}

func GetCmdTradeMarket(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "trade-market [market] [duration] [call/put] [quantity]",
		Short: "Create a new market trade",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			market := args[0]
			dur := args[1]
			ttype := mt.MicrotickTradeTypeFromName(args[2])
			quantity := mt.NewMicrotickQuantityFromString(args[3])
			
			txmsg := msg.NewTxMarketTrade(market, dur, cliCtx.GetFromAddress(), ttype,
				quantity)
			err := txmsg.ValidateBasic()
			if err != nil {
				return err
			}

			return authclient.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{txmsg})
		},
	}
}

func GetCmdTradeLimit(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "trade-limit [market] [duration] [call/put] [limit] [maxcost]",
		Short: "Create a new limit trade",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			market := args[0]
			dur := args[1]
			ttype := mt.MicrotickTradeTypeFromName(args[2])
			limit := mt.NewMicrotickPremiumFromString(args[3])
			maxcost := mt.NewMicrotickCoinFromString(args[4])
			
			txmsg := msg.NewTxLimitTrade(market, dur, cliCtx.GetFromAddress(), ttype,
				limit, maxcost)
			err := txmsg.ValidateBasic()
			if err != nil {
				return err
			}

			return authclient.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{txmsg})
		},
	}
}

func GetCmdTradePick(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "trade-pick [id] [call/put]",
		Short: "Create a new trade against specific quote id",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			id := mt.NewMicrotickIdFromString(args[0])
			ttype := mt.MicrotickTradeTypeFromName(args[1])
			
			txmsg := msg.NewTxPickTrade(cliCtx.GetFromAddress(), id, ttype)
			err := txmsg.ValidateBasic()
			if err != nil {
				return err
			}

			return authclient.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{txmsg})
		},
	}
}

func GetCmdTradeSettle(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "trade-settle [id]",
		Short: "Settle trade",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(authclient.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			id := mt.NewMicrotickIdFromString(args[0])
			
			txmsg := msg.NewTxSettleTrade(id, cliCtx.GetFromAddress())
			err := txmsg.ValidateBasic()
			if err != nil {
				return err
			}

			return authclient.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{txmsg})
		},
	}
}
