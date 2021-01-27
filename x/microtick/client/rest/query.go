package rest

import (
		"fmt"
		"net/http"
		"strconv"
	
		"github.com/gorilla/mux"
		"github.com/cosmos/cosmos-sdk/types/rest"
		"github.com/cosmos/cosmos-sdk/client"
		
		"gitlab.com/microtick/mtzone/x/microtick/msg"
)

func RegisterRoutes(cliCtx client.Context, r *mux.Router) {
		r.HandleFunc("/microtick/account/{acct}", queryAccountStatusHandler(cliCtx)).Methods("GET")
		r.HandleFunc("/microtick/market/{market}", queryMarketStatusHandler(cliCtx)).Methods("GET")
		r.HandleFunc("/microtick/consensus/{market}", queryMarketConsensusHandler(cliCtx)).Methods("GET")
		r.HandleFunc("/microtick/orderbook/{market}/{duration}", queryMarketOrderbookHandler(cliCtx)).Methods("GET")
		r.HandleFunc("/microtick/synthetic/{market}/{duration}", queryMarketSyntheticHandler(cliCtx)).Methods("GET")
		r.HandleFunc("/microtick/quote/{id}", queryQuoteHandler(cliCtx)).Methods("GET")
		r.HandleFunc("/microtick/trade/{id}", queryTradeHandler(cliCtx)).Methods("GET")
		r.HandleFunc("/microtick/params", queryParams(cliCtx)).Methods("GET")
}

func queryAccountStatusHandler(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		account := vars["acct"]
		
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		res, height, err := cliCtx.Query(fmt.Sprintf("custom/microtick/account/%s", account))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

    cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryMarketStatusHandler(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		market := vars["market"]

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		res, height, err := cliCtx.Query(fmt.Sprintf("custom/microtick/market/%s", market))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

    cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryMarketConsensusHandler(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		market := vars["market"]

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		res, height, err := cliCtx.Query(fmt.Sprintf("custom/microtick/consensus/%s", market))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

    cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryMarketOrderbookHandler(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		market := vars["market"]
		duration := vars["duration"]
		
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}
		
		offset, _ := strconv.Atoi(r.FormValue("offset"))
		limit, _ := strconv.Atoi(r.FormValue("limit"))
		params := msg.NewQueryBooksParams(offset, limit)
		
		bz, err := cliCtx.LegacyAmino.MarshalJSON(params)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/microtick/orderbook/%s/%s", market, duration), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

    cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryMarketSyntheticHandler(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		market := vars["market"]
		duration := vars["duration"]

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}
		
		offset, _ := strconv.Atoi(r.FormValue("offset"))
		limit, _ := strconv.Atoi(r.FormValue("limit"))
		params := msg.NewQueryBooksParams(offset, limit)
		
		bz, err := cliCtx.LegacyAmino.MarshalJSON(params)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/microtick/synthetic/%s/%s", market, duration), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

    cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryQuoteHandler(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		res, height, err := cliCtx.Query(fmt.Sprintf("custom/microtick/quote/%s", id))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

    cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryTradeHandler(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		res, height, err := cliCtx.Query(fmt.Sprintf("custom/microtick/trade/%s", id))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

    cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryParams(cliCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		res, height, err := cliCtx.Query(fmt.Sprintf("custom/microtick/params"))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

    cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}
