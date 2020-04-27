package microtick

import (
    "fmt"
    "errors"
    
    sdk "github.com/cosmos/cosmos-sdk/types"
    abci "github.com/tendermint/tendermint/abci/types"
    
    "github.com/mjackson001/mtzone/x/microtick/keeper"
    "github.com/mjackson001/mtzone/x/microtick/msg"
)

func NewQuerier(keeper keeper.Keeper) sdk.Querier {
    return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
        switch path[0] {
        case "account":
            return msg.QueryAccountStatus(ctx, path[1:], req, keeper)
        case "market":
            return msg.QueryMarketStatus(ctx, path[1:], req, keeper)
        case "consensus":
            return msg.QueryMarketConsensus(ctx, path[1:], req, keeper)
        case "orderbook":
            return msg.QueryOrderBook(ctx, path[1:], req, keeper)
        case "quote":
            return msg.QueryQuoteStatus(ctx, path[1:], req, keeper)
        case "trade":
            return msg.QueryTradeStatus(ctx, path[1:], req, keeper)
        case "generate":
            return msg.GenerateTx(ctx, path[1], path[2:], req, keeper)
        default:
            return nil, errors.New("unknown microtick query endpoint")
        }
    }
}

func NewHandler(keeper keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, txmsg sdk.Msg) (*sdk.Result, error) {
		switch tmp := txmsg.(type) {
		case msg.TxCreateMarket:
		    return msg.HandleTxCreateMarket(ctx, keeper, tmp)
		case msg.TxCreateQuote:
			return msg.HandleTxCreateQuote(ctx, keeper, tmp)
		case msg.TxCancelQuote:
			return msg.HandleTxCancelQuote(ctx, keeper, tmp)
		case msg.TxUpdateQuote:
			return msg.HandleTxUpdateQuote(ctx, keeper, tmp)
		case msg.TxDepositQuote:
			return msg.HandleTxDepositQuote(ctx, keeper, tmp)
		case msg.TxWithdrawQuote:
			return msg.HandleTxWithdrawQuote(ctx, keeper, tmp)
		case msg.TxMarketTrade:
			return msg.HandleTxMarketTrade(ctx, keeper, tmp)
		case msg.TxLimitTrade:
			return msg.HandleTxLimitTrade(ctx, keeper, tmp)
		case msg.TxPickTrade:
		    return msg.HandleTxPickTrade(ctx, keeper, tmp)
		case msg.TxSettleTrade:
			return msg.HandleTxSettleTrade(ctx, keeper, tmp)
		default:
			errMsg := fmt.Sprintf("Unrecognized microtick tx type: %v", txmsg.Type())
			return nil, errors.New(errMsg)
		}
	}
}

