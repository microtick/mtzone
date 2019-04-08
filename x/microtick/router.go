package microtick

import (
    "fmt"
    
    sdk "github.com/cosmos/cosmos-sdk/types"
    abci "github.com/tendermint/tendermint/abci/types"
    "github.com/cosmos/cosmos-sdk/codec"
)

func NewQuerier(keeper Keeper) sdk.Querier {
    return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err sdk.Error) {
        switch path[0] {
        case "account":
            return queryAccountStatus(ctx, path[1:], req, keeper)
        case "market":
            return queryMarketStatus(ctx, path[1:], req, keeper)
        case "orderbook":
            return queryOrderBook(ctx, path[1:], req, keeper)
        case "quote":
            return queryQuoteStatus(ctx, path[1:], req, keeper)
        case "trade":
            return queryTradeStatus(ctx, path[1:], req, keeper)
        default:
            return nil, sdk.ErrUnknownRequest("unknown microtick query endpoint")
        }
    }
}

func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case TxCreateMarket:
		    return handleTxCreateMarket(ctx, keeper, msg)
		case TxCreateQuote:
			return handleTxCreateQuote(ctx, keeper, msg)
		case TxCancelQuote:
			return handleTxCancelQuote(ctx, keeper, msg)
		case TxUpdateQuote:
			return handleTxUpdateQuote(ctx, keeper, msg)
		case TxDepositQuote:
			return handleTxDepositQuote(ctx, keeper, msg)
		case TxMarketTrade:
			return handleTxMarketTrade(ctx, keeper, msg)
		case TxLimitTrade:
			return handleTxLimitTrade(ctx, keeper, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized microtick tx type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Codec

func RegisterCodec(cdc *codec.Codec) {
    cdc.RegisterConcrete(TxCreateMarket{}, "microtick/CreateMarket", nil)
    cdc.RegisterConcrete(TxCreateQuote{}, "microtick/CreateQuote", nil)
    cdc.RegisterConcrete(TxCancelQuote{}, "microtick/CancelQuote", nil)
    cdc.RegisterConcrete(TxUpdateQuote{}, "microtick/UpdateQuote", nil)
    cdc.RegisterConcrete(TxDepositQuote{}, "microtick/DepositQuote", nil)
    cdc.RegisterConcrete(TxMarketTrade{}, "microtick/MarketTrade", nil)
    cdc.RegisterConcrete(TxLimitTrade{}, "microtick/LimitTrade", nil)
}
