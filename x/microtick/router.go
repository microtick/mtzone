package microtick

import (
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    sdk "github.com/cosmos/cosmos-sdk/types"
    abci "github.com/tendermint/tendermint/abci/types"
    
    mt "gitlab.com/microtick/mtzone/x/microtick/types"
    "gitlab.com/microtick/mtzone/x/microtick/keeper"
    "gitlab.com/microtick/mtzone/x/microtick/msg"
)

func NewQuerier(keeper keeper.Keeper) sdk.Querier {
    return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
        keeper.GetParams(ctx)
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
            return nil, sdkerrors.Wrap(mt.ErrInvalidRequest, "query endpoint")
        }
    }
}

func NewHandler(keeper keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, txmsg sdk.Msg) (*sdk.Result, error) {
	    ctx = ctx.WithEventManager(sdk.NewEventManager())
        params := keeper.GetParams(ctx)
		switch tmp := txmsg.(type) {
		case msg.TxCreateQuote:
			return msg.HandleTxCreateQuote(ctx, keeper, params, tmp)
		case msg.TxCancelQuote:
			return msg.HandleTxCancelQuote(ctx, keeper, tmp)
		case msg.TxUpdateQuote:
			return msg.HandleTxUpdateQuote(ctx, keeper, params, tmp)
		case msg.TxDepositQuote:
			return msg.HandleTxDepositQuote(ctx, keeper, params, tmp)
		case msg.TxWithdrawQuote:
			return msg.HandleTxWithdrawQuote(ctx, keeper, params, tmp)
		case msg.TxMarketTrade:
			return msg.HandleTxMarketTrade(ctx, keeper, params, tmp)
		case msg.TxLimitTrade:
			return msg.HandleTxLimitTrade(ctx, keeper, params, tmp)
		case msg.TxPickTrade:
		    return msg.HandleTxPickTrade(ctx, keeper, params, tmp)
		case msg.TxSettleTrade:
			return msg.HandleTxSettleTrade(ctx, keeper, params, tmp)
		default:
            return nil, sdkerrors.Wrapf(mt.ErrInvalidRequest, "tx type: %v", txmsg.Type())
		}
	}
}

