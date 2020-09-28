package microtick

import (
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    sdk "github.com/cosmos/cosmos-sdk/types"
    
    mt "github.com/mjackson001/mtzone/x/microtick/types"
    "github.com/mjackson001/mtzone/x/microtick/keeper"
    "github.com/mjackson001/mtzone/x/microtick/msg"
)

func NewHandler(keeper keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, txmsg sdk.Msg) (*sdk.Result, error) {
	    ctx = ctx.WithEventManager(sdk.NewEventManager())
        params := keeper.GetParams(ctx)
		switch tmp := txmsg.(type) {
		case *msg.TxCreateQuote:
			return msg.HandleTxCreateQuote(ctx, keeper, params, *tmp)
		case *msg.TxCancelQuote:
			return msg.HandleTxCancelQuote(ctx, keeper, params, *tmp)
		case *msg.TxUpdateQuote:
			return msg.HandleTxUpdateQuote(ctx, keeper, params, *tmp)
		case *msg.TxDepositQuote:
			return msg.HandleTxDepositQuote(ctx, keeper, params, *tmp)
		case *msg.TxWithdrawQuote:
			return msg.HandleTxWithdrawQuote(ctx, keeper, params, *tmp)
		case *msg.TxMarketTrade:
			return msg.HandleTxMarketTrade(ctx, keeper, params, *tmp)
		case *msg.TxPickTrade:
		    return msg.HandleTxPickTrade(ctx, keeper, params, *tmp)
		case *msg.TxSettleTrade:
			return msg.HandleTxSettleTrade(ctx, keeper, params, *tmp)
		default:
            return nil, sdkerrors.Wrapf(mt.ErrInvalidRequest, "tx type: %v", txmsg.Type())
		}
	}
}

