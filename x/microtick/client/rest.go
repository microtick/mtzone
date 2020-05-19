package client

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/client/context"
	_ "github.com/cosmos/cosmos-sdk/codec"

	"github.com/gorilla/mux"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc("/microtick/account/{acct}", queryAccountStatusHandler(cliCtx)).Methods("GET")
	r.HandleFunc("/microtick/market/{market}", queryMarketStatusHandler(cliCtx)).Methods("GET")
	r.HandleFunc("/microtick/consensus/{market}", queryMarketConsensusHandler(cliCtx)).Methods("GET")
	r.HandleFunc("/microtick/orderbook/{market}/{duration}", queryMarketOrderbookHandler(cliCtx)).Methods("GET")
	r.HandleFunc("/microtick/quote/{id}", queryQuoteHandler(cliCtx)).Methods("GET")
	r.HandleFunc("/microtick/trade/{id}", queryTradeHandler(cliCtx)).Methods("GET")
}

type signedReq struct {
	Tx string `json:"tx"`
}

func queryAccountStatusHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		account := vars["acct"]

		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/microtick/account/%s", account), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

    cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryMarketStatusHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		market := vars["market"]

		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/microtick/market/%s", market), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

    cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryMarketConsensusHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		market := vars["market"]

		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/microtick/consensus/%s", market), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

    cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryMarketOrderbookHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		market := vars["market"]
		duration := vars["duration"]

		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/microtick/orderbook/%s/%s", market, duration), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

    cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryQuoteHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/microtick/quote/%s", id), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

    cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryTradeHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/microtick/trade/%s", id), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

    cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

