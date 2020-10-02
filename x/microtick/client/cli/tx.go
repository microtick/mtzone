package cli

import (
  "github.com/cosmos/cosmos-sdk/client"
  "github.com/cosmos/cosmos-sdk/client/flags"
  "github.com/cosmos/cosmos-sdk/client/tx"
  
  "github.com/spf13/cobra"
	
	mt "gitlab.com/microtick/mtzone/x/microtick/types"
	"gitlab.com/microtick/mtzone/x/microtick/msg"
)

func GetTxCmd(key string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "microtick",
		Short: "Microtick transactions subcommands",
		RunE: client.ValidateCmd,
	}

	cmd.AddCommand(
		cmdQuoteCancel(),
		cmdQuoteCreate(),
		cmdQuoteDeposit(),
		cmdQuoteUpdate(),
		cmdQuoteWithdraw(),
		cmdTradeMarket(),
		cmdTradePick(),
		cmdTradeSettle(),
	)

	return cmd
}

func cmdQuoteCancel() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel [id]",
		Short: "Cancel a quote",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			
			message := msg.TxCancelQuote {
			}
			
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &message)
		},
	}
	
	flags.AddTxFlagsToCmd(cmd)
	
	return cmd
}

func cmdQuoteCreate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [market] [duration] [backing] [spot] [ask_premium] [bid_premium]",
		Short: "Create a new quote",
		Args:  cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			
			message := msg.TxCreateQuote {
				Market: args[0],
				Duration: args[1],
				Provider: clientCtx.GetFromAddress(),
				Backing: mt.NewMicrotickCoinFromString(args[2]),
				Spot: mt.NewMicrotickSpotFromString(args[3]),
				Ask: mt.NewMicrotickPremiumFromString(args[4]),
				Bid: mt.NewMicrotickPremiumFromString(args[5]),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &message)
		},
	}
	
	flags.AddTxFlagsToCmd(cmd)
	
	return cmd
}

func cmdQuoteDeposit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit [id] [amount]",
		Short: "Deposit more backing to a quote",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			
			message := msg.TxDepositQuote {
				Id: mt.NewMicrotickIdFromString(args[0]),
				Requester: clientCtx.GetFromAddress(),
				Deposit: mt.NewMicrotickCoinFromString(args[1]),
			}
			
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &message)
		},
	}
	
	flags.AddTxFlagsToCmd(cmd)
	
	return cmd
}

func cmdQuoteUpdate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [id] [newspot] [new_ask_premium] [new_bid_premium]",
		Short: "Update a quote",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			
			message := msg.TxUpdateQuote {
				Id: mt.NewMicrotickIdFromString(args[0]),
				Requester: clientCtx.GetFromAddress(),
				NewSpot: mt.NewMicrotickSpotFromString(args[1]),
				NewAsk: mt.NewMicrotickPremiumFromString(args[2]),
				NewBid: mt.NewMicrotickPremiumFromString(args[3]),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &message)
		},
	}
	
	flags.AddTxFlagsToCmd(cmd)
	
	return cmd
}

func cmdQuoteWithdraw() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw [id] [amount]",
		Short: "Withdraw backing from a quote",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			
			message := msg.TxWithdrawQuote {
				Id: mt.NewMicrotickIdFromString(args[0]),
				Requester: clientCtx.GetFromAddress(),
				Withdraw: mt.NewMicrotickCoinFromString(args[1]),
			}
			
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &message)
		},
	}
	
	flags.AddTxFlagsToCmd(cmd)
	
	return cmd
}

func cmdTradeMarket() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trade [market] [duration] [buy/sell] [call/put/syn] [quantity]",
		Short: "Create a new market trade",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			
			message := msg.TxMarketTrade {
				Market: args[0],
				Duration: args[1],
				Taker: clientCtx.GetFromAddress(),
				OrderType: args[2] + "-" + args[3],
				Quantity: mt.NewMicrotickQuantityFromString(args[4]),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &message)
		},
	}
	
	flags.AddTxFlagsToCmd(cmd)
	
	return cmd
}

func cmdTradePick() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pick [id] [buy/sell] [call/put]",
		Short: "Create a new trade against specific quote id",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			
			message := msg.TxPickTrade {
				Id: mt.NewMicrotickIdFromString(args[0]),
				Taker: clientCtx.GetFromAddress(),
				OrderType: args[1] + "-" + args[2],
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &message)
		},
	}
	
	flags.AddTxFlagsToCmd(cmd)
	
	return cmd
}

func cmdTradeSettle() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "settle [id]",
		Short: "Settle trade",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			clientCtx, err := client.ReadTxCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			
			message := msg.TxSettleTrade {
				Id: mt.NewMicrotickIdFromString(args[0]),
				Requester: clientCtx.GetFromAddress(),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &message)
		},
	}
	
	flags.AddTxFlagsToCmd(cmd)
	
	return cmd
}
