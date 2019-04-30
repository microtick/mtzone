package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/gorilla/mux"
)


// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec, storeName string) {
	r.HandleFunc(fmt.Sprintf("/%s/account/{acct}", storeName), queryAccountStatusHandler(cdc, cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/market/{market}", storeName), queryMarketStatusHandler(cdc, cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/consensus/{market}", storeName), queryMarketConsensusHandler(cdc, cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/orderbook/{market}/{duration}", storeName), queryMarketOrderbookHandler(cdc, cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/quote/{id}", storeName), queryQuoteHandler(cdc, cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/trade/{id}", storeName), queryTradeHandler(cdc, cliCtx, storeName)).Methods("GET")
	
	// These tx functions just generate the signing bytes with correct chain ID, account and sequence numbers
	r.HandleFunc(fmt.Sprintf("/%s/createmarket/{acct}/{market}", storeName), txCreateMarketHandler(cdc, cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/createquote/{acct}/{market}/{duration}/{backing}/{spot}/{premium}", storeName), txCreateQuoteHandler(cdc, cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/cancelquote/{requester}/{id}", storeName), txCancelQuoteHandler(cdc, cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/depositquote/{requester}/{id}/{amount}", storeName), txDepositQuoteHandler(cdc, cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/updatequote/{requester}/{id}/{spot}/{premium}", storeName), txUpdateQuoteHandler(cdc, cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/markettrade/{buyer}/{market}/{duration}/{tradetype}/{quantity}", storeName), txMarketTradeHandler(cdc, cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/limittrade/{buyer}/{market}/{duration}/{tradetype}/{limit}", storeName), txLimitTradeHandler(cdc, cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/settletrade/{requester}/{id}", storeName), txSettleTradeHandler(cdc, cliCtx, storeName)).Methods("GET")
	
	// Broadcast signed tx
	r.HandleFunc(fmt.Sprintf("/%s/broadcast", storeName), broadcastSignedTx(cdc, cliCtx)).Methods("POST")
}

type signedReq struct {
	Tx string `json:"tx"`
}

func broadcastSignedTx(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req signedReq
		
		if !rest.ReadRESTReq(w, r, cdc, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return	
		}
		
		fmt.Printf("%s\n", req.Tx)
		
		var msg auth.StdTx
		cdc.MustUnmarshalJSON([]byte(req.Tx), &msg)
		
		encoder := utils.GetTxEncoder(cdc)
		
		bytes, _ := encoder(msg)
		
		res, err := cliCtx.BroadcastTxAsync(bytes)
		if err != nil {
			fmt.Printf("%s\n", err.Error())
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		
		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

func txCreateMarketHandler(cdc *codec.Codec, cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		account := vars["acct"]
		market := vars["market"]
		
		res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/generate/createmarket/%s/%s", storeName, account, market), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

func txCreateQuoteHandler(cdc *codec.Codec, cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		account := vars["acct"]
		market := vars["market"]
		duration := vars["duration"]
		backing := vars["backing"]
		spot := vars["spot"]
		premium := vars["premium"]
		
		res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/generate/createquote/%s/%s/%s/%s/%s/%s", storeName, 
			account, market, duration, backing, spot, premium), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

func txCancelQuoteHandler(cdc *codec.Codec, cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		requester := vars["requester"]
		
		res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/generate/cancelquote/%s/%s", storeName, 
			requester, id), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

func txDepositQuoteHandler(cdc *codec.Codec, cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		requester := vars["requester"]
		amount := vars["amount"]
		
		res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/generate/depositquote/%s/%s/%s", storeName, 
			requester, id, amount), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

func txUpdateQuoteHandler(cdc *codec.Codec, cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		requester := vars["requester"]
		spot := vars["spot"]
		premium := vars["premium"]
		
		res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/generate/updatequote/%s/%s/%s/%s", storeName, 
			requester, id, spot, premium), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

func txMarketTradeHandler(cdc *codec.Codec, cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		market := vars["market"]
		duration := vars["duration"]
		buyer := vars["buyer"]
		tradetype := vars["tradetype"]
		quantity := vars["quantity"]
		
		res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/generate/markettrade/%s/%s/%s/%s/%s", storeName, 
			buyer, market, duration, tradetype, quantity), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

func txLimitTradeHandler(cdc *codec.Codec, cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		market := vars["market"]
		duration := vars["duration"]
		buyer := vars["buyer"]
		tradetype := vars["tradetype"]
		limit := vars["limit"]
		
		res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/generate/limittrade/%s/%s/%s/%s/%s", storeName, 
			buyer, market, duration, tradetype, limit), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

func txSettleTradeHandler(cdc *codec.Codec, cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		requester := vars["requester"]
		
		res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/generate/settletrade/%s/%s", storeName, 
			requester, id), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

func queryAccountStatusHandler(cdc *codec.Codec, cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		account := vars["acct"]

		res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/account/%s", storeName, account), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

func queryMarketStatusHandler(cdc *codec.Codec, cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		market := vars["market"]

		res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/market/%s", storeName, market), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

func queryMarketConsensusHandler(cdc *codec.Codec, cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		market := vars["market"]

		res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/consensus/%s", storeName, market), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

func queryMarketOrderbookHandler(cdc *codec.Codec, cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		market := vars["market"]
		duration := vars["duration"]

		res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/orderbook/%s/%s", storeName, market, duration), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

func queryQuoteHandler(cdc *codec.Codec, cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/quote/%s", storeName, id), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

func queryTradeHandler(cdc *codec.Codec, cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/trade/%s", storeName, id), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}

		rest.PostProcessResponse(w, cdc, res, cliCtx.Indent)
	}
}

