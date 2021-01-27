package msg

import (
	"strconv"
	
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"
	
	"gitlab.com/microtick/mtzone/x/microtick/keeper"
	mt "gitlab.com/microtick/mtzone/x/microtick/types"
)

// Querier if used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper
type Querier struct {
	keeper.Keeper
}

const (
	QueryAccount = "account"
	QueryConsensus = "consensus"
	QueryMarket = "market"
	QueryOrderBook = "orderbook"
	QuerySynthetic = "synthetic"
	QueryQuote = "quote"
	QueryTrade = "trade"
	QueryParams = "params"
)

type QueryBooksParams struct {
	Offset int `json:"offset"`
	Limit int `json:"limit"`
}

func NewQueryBooksParams(offset, limit int) QueryBooksParams {
	return QueryBooksParams {
		Offset: offset,
		Limit: limit,
	}
}

func NewQuerier(keeper keeper.Keeper, legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
			case "account":
			  return queryAccount(ctx, path[1:], req, keeper, legacyQuerierCdc)
			case "consensus":
			  return queryConsensus(ctx, path[1:], req, keeper, legacyQuerierCdc)
			case QueryMarket:
			  return queryMarket(ctx, path[1:], req, keeper, legacyQuerierCdc)
			case "orderbook":
			  return queryOrderbook(ctx, path[1:], req, keeper, legacyQuerierCdc)
			case "synthetic":
			  return querySynthetic(ctx, path[1:], req, keeper, legacyQuerierCdc)
			case "quote":
			  return queryQuote(ctx, path[1:], req, keeper, legacyQuerierCdc)
			case "trade":
			  return queryTrade(ctx, path[1:], req, keeper, legacyQuerierCdc)
			case "params":
			  return queryParams(ctx, path[1:], req, keeper, legacyQuerierCdc)
			default:
			  return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown query path: %s", path[0])
		}
	}
}

func queryAccount(ctx sdk.Context, path []string, req abci.RequestQuery, keeper keeper.Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	acct, err := sdk.AccAddressFromBech32(path[0])
	if err != nil {
		return nil, err
	}
	
	pbReq := QueryAccountRequest {
		Account: acct,
	}
	
	pbRes, err := baseQueryAccount(ctx, keeper, &pbReq)
	if err != nil {
		return nil, err
	}
	
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, pbRes)
	if err != nil {
		return nil, err
	}
	
	return bz, nil
}

func queryConsensus(ctx sdk.Context, path []string, req abci.RequestQuery, keeper keeper.Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	pbReq := QueryConsensusRequest {
		Market: path[0],
	}
	
	pbRes, err := baseQueryConsensus(ctx, keeper, &pbReq)
	if err != nil {
		return nil, err
	}
	
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, pbRes)
	if err != nil {
		return nil, err
	}
	
	return bz, nil
}

func queryMarket(ctx sdk.Context, path []string, req abci.RequestQuery, keeper keeper.Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	pbReq := QueryMarketRequest {
		Market: path[0],
	}
	
	pbRes, err := baseQueryMarket(ctx, keeper, &pbReq)
	if err != nil {
		return nil, err
	}
	
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, pbRes)
	if err != nil {
		return nil, err
	}
	
	return bz, nil
}

func queryOrderbook(ctx sdk.Context, path []string, req abci.RequestQuery, keeper keeper.Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params QueryBooksParams
	
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		// defaults
		params.Offset = 0
		params.Limit = 10
	}
	
	pbReq := QueryOrderBookRequest {
		Market: path[0],
		Duration: path[1],
		Offset: uint32(params.Offset),
		Limit: uint32(params.Limit),
	}
	
	pbRes, err := baseQueryOrderBook(ctx, keeper, &pbReq)
	if err != nil {
		return nil, err
	}
	
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, pbRes)
	if err != nil {
		return nil, err
	}
	
	return bz, nil
}

func querySynthetic(ctx sdk.Context, path []string, req abci.RequestQuery, keeper keeper.Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	var params QueryBooksParams
	
	err := legacyQuerierCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		// defaults
		params.Offset = 0
		params.Limit = 10
	}
	
	pbReq := QuerySyntheticRequest {
		Market: path[0],
		Duration: path[1],
		Offset: uint32(params.Offset),
		Limit: uint32(params.Limit),
	}
	
	pbRes, err := baseQuerySynthetic(ctx, keeper, &pbReq)
	if err != nil {
		return nil, err
	}
	
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, pbRes)
	if err != nil {
		return nil, err
	}
	
	return bz, nil
}

func queryQuote(ctx sdk.Context, path []string, req abci.RequestQuery, keeper keeper.Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	id, err := strconv.ParseUint(path[0], 10, 32)
	if err != nil {
		return nil, err
	}
	
	pbReq := QueryQuoteRequest {
		Id: mt.MicrotickId(id),
	}
	
	pbRes, err := baseQueryQuote(ctx, keeper, &pbReq)
	if err != nil {
		return nil, err
	}
	
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, pbRes)
	if err != nil {
		return nil, err
	}
	
	return bz, nil
}

func queryTrade(ctx sdk.Context, path []string, req abci.RequestQuery, keeper keeper.Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	id, err := strconv.ParseUint(path[0], 10, 32)
	if err != nil {
		return nil, err
	}
	
	pbReq := QueryTradeRequest {
		Id: mt.MicrotickId(id),
	}
	
	pbRes, err := baseQueryTrade(ctx, keeper, &pbReq)
	if err != nil {
		return nil, err
	}
	
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, pbRes)
	if err != nil {
		return nil, err
	}
	
	return bz, nil
}

func queryParams(ctx sdk.Context, path []string, req abci.RequestQuery, keeper keeper.Keeper, legacyQuerierCdc *codec.LegacyAmino) ([]byte, error) {
	pbReq := QueryParamsRequest {}
	
	pbRes, err := baseQueryParams(ctx, keeper, &pbReq)
	if err != nil {
		return nil, err
	}
	
	bz, err := codec.MarshalJSONIndent(legacyQuerierCdc, pbRes)
	if err != nil {
		return nil, err
	}
	
	return bz, nil
}