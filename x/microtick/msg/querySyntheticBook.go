package msg

import (
    "fmt"
    "strings"
    
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    abci "github.com/tendermint/tendermint/abci/types"
    
    mt "github.com/mjackson001/mtzone/x/microtick/types"
    "github.com/mjackson001/mtzone/x/microtick/keeper"
)

type ResponseSyntheticBook struct {
    Consensus mt.MicrotickSpot `json:"consensus"`
    Weight mt.MicrotickQuantity `json:"weight"`
    Asks []ResponseSyntheticQuote `json:"asks"`
    Bids []ResponseSyntheticQuote `json:"bids"`
}

type ResponseSyntheticQuote struct {
    Spot sdk.Dec `json:"spot"`
    Quantity sdk.Dec `json:"quantity"`
}

func (rma ResponseSyntheticBook) String() string {
    var i int
    var a, b string
    for i = 0; i < len(rma.Asks); i++ {
        a += formatSyntheticQuote(rma.Asks[i]) + "\n"
    }
    for i = 0; i < len(rma.Bids); i++ {
        b += formatSyntheticQuote(rma.Bids[i]) + "\n"
    }
    return strings.TrimSpace(fmt.Sprintf(`Consensus: %s
Weight: %s
Asks: 
%sBids: 
%s`, rma.Consensus, rma.Weight, a, b))
}

func formatSyntheticQuote(robq ResponseSyntheticQuote) string {
    return fmt.Sprintf(`  spot: %s quantity: %s`, robq.Spot.String(), robq.Quantity.String())
}

func QuerySyntheticBook(ctx sdk.Context, path []string, 
    req abci.RequestQuery, keeper keeper.Keeper)(res []byte, err error) {
        
    market := path[0]
    durName := path[1]
    
    dataMarket, err := keeper.GetDataMarket(ctx, market)
    if err != nil {
        return nil, sdkerrors.Wrap(mt.ErrInvalidMarket, market)
    }
    
    syntheticBook := keeper.GetSyntheticBook(ctx, dataMarket, durName)
    
    asks := make([]ResponseSyntheticQuote, len(syntheticBook.Asks))
    bids := make([]ResponseSyntheticQuote, len(syntheticBook.Bids))
    for i := 0; i < len(syntheticBook.Asks); i++ {
        asks[i] = ResponseSyntheticQuote {
            Spot: syntheticBook.Asks[i].Spot.Amount,
            Quantity: syntheticBook.Asks[i].Quantity.Amount,
        }
    }
    for i := 0; i < len(syntheticBook.Bids); i++ {
        bids[i] = ResponseSyntheticQuote {
            Spot: syntheticBook.Bids[i].Spot.Amount,
            Quantity: syntheticBook.Bids[i].Quantity.Amount,
        }
    }
    response := ResponseSyntheticBook {
        Consensus:  dataMarket.Consensus,
        Weight: syntheticBook.Weight,
        Asks: asks,
        Bids: bids,
    }
    
    bz, err := codec.MarshalJSONIndent(keeper.Cdc, response)
    if err != nil {
        panic("Could not marshal result to JSON")
    }
    
    return bz, nil
}
