package cli

import (
  "context"
  "strconv"
    
  "github.com/cosmos/cosmos-sdk/client"
  "github.com/cosmos/cosmos-sdk/client/flags"
  "github.com/spf13/cobra"
    
  sdk "github.com/cosmos/cosmos-sdk/types"
  mt "gitlab.com/microtick/mtzone/x/microtick/types"
  "gitlab.com/microtick/mtzone/x/microtick/msg"
)

func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "microtick",
		Short: "Querying commands for the microtick module",
	}

	cmd.AddCommand(
		cmdAccountStatus(),
		cmdMarketStatus(),
		cmdMarketConsensus(),
		cmdOrderBook(),
		cmdSyntheticBook(),
		cmdQuote(),
		cmdTrade(),
		cmdParams(),
	)

	return cmd
}

func cmdAccountStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account [acct]",
		Short: "Query account details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			
			acct, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}
			
			message := &msg.QueryAccountRequest {
				Account: acct,
			}
			
			queryClient := msg.NewGRPCClient(clientCtx)
			res, err := queryClient.Account(context.Background(), message)
			if err != nil {
				return err
			}
			
			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	
	return cmd
}

func cmdMarketConsensus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "consensus [market]",
		Short: "Query market consensus",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}			
			
			message := &msg.QueryConsensusRequest {
				Market: args[0],
			}

			queryClient := msg.NewGRPCClient(clientCtx)
			res, err := queryClient.Consensus(context.Background(), message)
			if err != nil {
				return err
			}
			
			return clientCtx.PrintProto(res)
		},
	}
	
	flags.AddQueryFlagsToCmd(cmd)
	
	return cmd
}

func cmdMarketStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "market [name]",
		Short: "Query market status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			message := &msg.QueryMarketRequest {
				Market: args[0],
			}
			
			queryClient := msg.NewGRPCClient(clientCtx)
			res, err := queryClient.Market(context.Background(), message)
			if err != nil {
				return err
			}
			
			return clientCtx.PrintProto(res)
		},
	}
	
	flags.AddQueryFlagsToCmd(cmd)
	
	return cmd
}

func cmdOrderBook() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "orderbook [market] [dur]",
		Short: "Query market orderbook",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}			
			
			message := &msg.QueryOrderBookRequest {
				Market: args[0],
				Duration: args[1],
			}

			queryClient := msg.NewGRPCClient(clientCtx)
			res, err := queryClient.OrderBook(context.Background(), message)
			if err != nil {
				return err
			}
			
			return clientCtx.PrintProto(res)
		},
	}
	
	flags.AddQueryFlagsToCmd(cmd)
	
	return cmd
}

func cmdSyntheticBook() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "synthetic [market] [dur]",
		Short: "Query synthetic orderbook",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}			
			
			message := &msg.QuerySyntheticRequest {
				Market: args[0],
				Duration: args[1],
			}
			
			queryClient := msg.NewGRPCClient(clientCtx)
			res, err := queryClient.Synthetic(context.Background(), message)
			if err != nil {
				return err
			}
			
			return clientCtx.PrintProto(res)			
		},
	}
	
	flags.AddQueryFlagsToCmd(cmd)
	
	return cmd
}

func cmdQuote() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quote [id]",
		Short: "Query quote",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}			
			
			id, err := strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				return err
			}			
			
			message := &msg.QueryQuoteRequest {
				Id: mt.MicrotickId(id),
			}
			
			queryClient := msg.NewGRPCClient(clientCtx)
			res, err := queryClient.Quote(context.Background(), message)
			if err != nil {
				return err
			}
			
			return clientCtx.PrintProto(res)			
		},
	}
	
	flags.AddQueryFlagsToCmd(cmd)
	
	return cmd
}

func cmdTrade() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trade [id]",
		Short: "Query trade",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}			
			
			id, err := strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				return err
			}			
			
			message := &msg.QueryTradeRequest {
				Id: mt.MicrotickId(id),
			}
			
			queryClient := msg.NewGRPCClient(clientCtx)
			res, err := queryClient.Trade(context.Background(), message)
			if err != nil {
				return err
			}
			
			return clientCtx.PrintProto(res)			
		},
	}
	
	flags.AddQueryFlagsToCmd(cmd)
	
	return cmd
}

func cmdParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query module params",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}	
			
			message := &msg.QueryParamsRequest {}
			
			queryClient := msg.NewGRPCClient(clientCtx)
			res, err := queryClient.Params(context.Background(), message)
			if err != nil {
				return err
			}
			
			return clientCtx.PrintProto(res)	
		},
	}
	
	flags.AddQueryFlagsToCmd(cmd)
	
	return cmd
}
