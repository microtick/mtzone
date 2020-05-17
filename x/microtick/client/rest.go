package client

import (
	"fmt"
	"net/http"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/client/context"
	_ "github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth"

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
	
	// These tx functions just generate the signing bytes with correct chain ID, account and sequence numbers
	r.HandleFunc("/microtick/createquote/{acct}/{market}/{duration}/{backing}/{spot}/{premium}", txCreateQuoteHandler(cliCtx)).Methods("GET")
	r.HandleFunc("/microtick/cancelquote/{requester}/{id}", txCancelQuoteHandler(cliCtx)).Methods("GET")
	r.HandleFunc("/microtick/depositquote/{requester}/{id}/{amount}", txDepositQuoteHandler(cliCtx)).Methods("GET")
	r.HandleFunc("/microtick/updatequote/{requester}/{id}/{spot}/{premium}", txUpdateQuoteHandler(cliCtx)).Methods("GET")
	r.HandleFunc("/microtick/markettrade/{buyer}/{market}/{duration}/{tradetype}/{quantity}", txMarketTradeHandler(cliCtx)).Methods("GET")
	r.HandleFunc("/microtick/limittrade/{buyer}/{market}/{duration}/{tradetype}/{limit}/{maxcost}", txLimitTradeHandler(cliCtx)).Methods("GET")
	r.HandleFunc("/microtick/settletrade/{requester}/{id}", txSettleTradeHandler(cliCtx)).Methods("GET")
	
	// Broadcast signed tx
	r.HandleFunc("/microtick/broadcast", broadcastSignedTx(cliCtx)).Methods("POST")
}

type signedReq struct {
	Tx string `json:"tx"`
}

func broadcastSignedTx(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req signedReq
		
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return	
		}
		
		//fmt.Printf("%s\n", req.Tx)
		
		var msg auth.StdTx
		cliCtx.Codec.MustUnmarshalJSON([]byte(req.Tx), &msg)
		
		encoder := sdk.GetConfig().GetTxEncoder()
		
		bytes, _ := encoder(msg)
		
		res, err := cliCtx.BroadcastTxAsync(bytes)
		if err != nil {
			fmt.Printf("%s\n", err.Error())
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func txCreateQuoteHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		account := vars["acct"]
		market := vars["market"]
		duration := vars["duration"]
		backing := vars["backing"]
		spot := vars["spot"]
		premium := vars["premium"]
		
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/microtick/generate/createquote/%s/%s/%s/%s/%s/%s", 
			account, market, duration, backing, spot, premium), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func txCancelQuoteHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		requester := vars["requester"]
		
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/microtick/generate/cancelquote/%s/%s",
			requester, id), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func txDepositQuoteHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		requester := vars["requester"]
		amount := vars["amount"]
		
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/microtick/generate/depositquote/%s/%s/%s", 
			requester, id, amount), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func txWithdrawQuoteHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		requester := vars["requester"]
		amount := vars["amount"]
		
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/microtick/generate/withdrawquote/%s/%s/%s", 
			requester, id, amount), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func txUpdateQuoteHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		requester := vars["requester"]
		spot := vars["spot"]
		premium := vars["premium"]
		
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/microtick/generate/updatequote/%s/%s/%s/%s", 
			requester, id, spot, premium), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func txMarketTradeHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		market := vars["market"]
		duration := vars["duration"]
		buyer := vars["buyer"]
		tradetype := vars["tradetype"]
		quantity := vars["quantity"]
		
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/microtick/generate/markettrade/%s/%s/%s/%s/%s", 
			buyer, market, duration, tradetype, quantity), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func txLimitTradeHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		market := vars["market"]
		duration := vars["duration"]
		buyer := vars["buyer"]
		tradetype := vars["tradetype"]
		limit := vars["limit"]
		maxcost := vars["maxcost"]
		
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/microtick/generate/limittrade/%s/%s/%s/%s/%s/%s", 
			buyer, market, duration, tradetype, limit, maxcost), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func txSettleTradeHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		requester := vars["requester"]
		
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/micrtick/generate/settletrade/%s/%s", 
			requester, id), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryAccountStatusHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		account := vars["acct"]

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/microtick/account/%s", account), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryMarketStatusHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		market := vars["market"]

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/microtick/market/%s", market), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryMarketConsensusHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		market := vars["market"]

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/microtick/consensus/%s", market), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryMarketOrderbookHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		market := vars["market"]
		duration := vars["duration"]

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/microtick/orderbook/%s/%s", market, duration), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryQuoteHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/microtick/quote/%s", id), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func queryTradeHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/microtick/trade/%s", id), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

