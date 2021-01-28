package cli

import (
	"fmt"
	"strings"
	
  "github.com/cosmos/cosmos-sdk/client"
  "github.com/cosmos/cosmos-sdk/client/flags"
  "github.com/cosmos/cosmos-sdk/client/tx"
  
  "github.com/spf13/cobra"
  
	sdk "github.com/cosmos/cosmos-sdk/types"
  "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
 	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
 	"github.com/cosmos/cosmos-sdk/version"
	
	mt "gitlab.com/microtick/mtzone/x/microtick/types"
	"gitlab.com/microtick/mtzone/x/microtick/msg"
)

const (
	FlagExtPerInt = "ext-per-int"
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
			clientCtx, err := client.GetClientTxContext(cmd)
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
		Use:   "create [market] [duration] [backing] [spot] [premium]",
		Short: "Create a new quote",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			
			bid, _ := cmd.Flags().GetString("bid")
			
			message := msg.TxCreateQuote {
				Market: args[0],
				Duration: args[1],
				Provider: clientCtx.GetFromAddress(),
				Backing: args[2],
				Spot: args[3],
				Ask: args[4],
				Bid: bid,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &message)
		},
	}
	
	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String("bid", "0premium", "bid premium")
	
	return cmd
}

func cmdQuoteDeposit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit [id] [amount]",
		Short: "Deposit more backing to a quote",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			
			message := msg.TxDepositQuote {
				Id: mt.NewMicrotickIdFromString(args[0]),
				Requester: clientCtx.GetFromAddress(),
				Deposit: args[1],
			}
			
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &message)
		},
	}
	
	flags.AddTxFlagsToCmd(cmd)
	
	return cmd
}

func cmdQuoteUpdate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [id] [spot] [premium]",
		Short: "Update a quote",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			
			bid, _ := cmd.Flags().GetString("bid")
			
			message := msg.TxUpdateQuote {
				Id: mt.NewMicrotickIdFromString(args[0]),
				Requester: clientCtx.GetFromAddress(),
				NewSpot: args[1],
				NewAsk: args[2],
				NewBid: bid,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &message)
		},
	}
	
	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String("bid", "0premium", "bid premium")
	
	return cmd
}

func cmdQuoteWithdraw() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw [id] [amount]",
		Short: "Withdraw backing from a quote",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			
			message := msg.TxWithdrawQuote {
				Id: mt.NewMicrotickIdFromString(args[0]),
				Requester: clientCtx.GetFromAddress(),
				Withdraw: args[1],
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
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			
			message := msg.TxMarketTrade {
				Market: args[0],
				Duration: args[1],
				Taker: clientCtx.GetFromAddress(),
				OrderType: args[2] + "-" + args[3],
				Quantity: args[4],
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
			clientCtx, err := client.GetClientTxContext(cmd)
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
			clientCtx, err := client.GetClientTxContext(cmd)
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

// NewCmdSubmitDenomChangeProposal implements a command handler for submitting a software upgrade proposal transaction.
func NewCmdSubmitDenomChangeProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "microtick-denom-change [denom] [flags]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a denom change proposal for the microtick module",
		Long: "Submit a denom along with an initial deposit.\n",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			extDenom := args[0]
			content, err := parseDenomChangeArgsToContent(cmd, extDenom)
			if err != nil {
				return err
			}

			from := clientCtx.GetFromAddress()

			depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
			if err != nil {
				return err
			}
			
			deposit, err := sdk.ParseCoinsNormalized(depositStr)
			if err != nil {
				return err
			}

			msg, err := gov.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}

			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(cli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(cli.FlagDescription, "", "description of proposal")
	cmd.Flags().String(cli.FlagDeposit, "", "deposit of proposal")
	cmd.Flags().Int64(FlagExtPerInt, 1000000, "how many external units per internal backing (default 1000000)")

	return cmd
}

func parseDenomChangeArgsToContent(cmd *cobra.Command, extDenom string) (gov.Content, error) {
	title, err := cmd.Flags().GetString(cli.FlagTitle)
	if err != nil {
		return nil, err
	}

	description, err := cmd.Flags().GetString(cli.FlagDescription)
	if err != nil {
		return nil, err
	}
	
	extPerInt, err := cmd.Flags().GetInt64(FlagExtPerInt)
	if err != nil {
		return nil, err
	}
	
	content := msg.NewDenomChangeProposal(title, description, extDenom, extPerInt)
	return content, nil
}

func NewCmdSubmitAddMarketsProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "microtick-add-markets --proposal=[proposal-file] [flags]",
		Args:  cobra.ExactArgs(0),
		Short: "Submit a proposal to add markets in the microtick module",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit the new proposal along with an initial deposit. 
The proposal details must be supplied via a JSON file.FlagExtPerInt

Example:
$ %s tx gov submit-proposal microtick-add-markets --proposal=<path/to/proposal.json> --from=<key_or_address>

Where proposal.json contains:

{
  "title": "Proposal Name Goes Here",
  "description": "Fill in with the proposal description",
  "markets": [
    {
    	"name": "XBTUSD",
    	"description": "XBTUSD: Crypto - Bitcoin / USD"
    },
    {
    	"name": "ETHUSD",
    	"description": "ETHUSD: Crypto - Ethereum / USD"
    }
  ]
}
`,
			version.AppName)),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			from := clientCtx.GetFromAddress()

			depositStr, err := cmd.Flags().GetString(cli.FlagDeposit)
			if err != nil {
				return err
			}
			
			deposit, err := sdk.ParseCoinsNormalized(depositStr)
			if err != nil {
				return err
			}
			
			file, _ := cmd.Flags().GetString(cli.FlagProposal)
			content, err := msg.NewAddMarketsProposal(file)
			if err != nil {
				return err
			}

			msg, err := gov.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}
			
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(cli.FlagTitle, "", "title of proposal")
	cmd.Flags().String(cli.FlagDescription, "", "description of proposal")
	cmd.Flags().String(cli.FlagDeposit, "", "deposit of proposal")
	cmd.Flags().String(cli.FlagProposal, "", "proposal filename")

	return cmd
}
