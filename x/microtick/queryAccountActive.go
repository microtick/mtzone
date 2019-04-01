package microtick

import (
    "fmt"
    "strings"
    
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
    abci "github.com/tendermint/tendermint/abci/types"
)

type ResponseAccountActive struct {
    Quotes OrderedList `json:"quotes"`
    Trades OrderedList `json:"trades"`
}

func (raa ResponseAccountActive) String() string {
    return strings.TrimSpace(fmt.Sprintf(`Quotes: %v
Trades: %v`, raa.Quotes.Data, raa.Trades.Data))
}

func queryAccountActive(ctx sdk.Context, path []string, 
    req abci.RequestQuery, keeper Keeper)(res []byte, err sdk.Error) {
    acct := path[0]
    data := keeper.GetAccountStatus(ctx, acct)
    response := ResponseAccountActive {
        Quotes: data.ActiveQuotes,
        Trades: data.ActiveTrades,
    }
    
    bz, err2 := codec.MarshalJSONIndent(keeper.cdc, response)
    if err2 != nil {
        panic("could not marshal result to JSON")
    }
    
    return bz, nil
}