package microtick

import (
    "fmt"
    "strings"
    
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
    abci "github.com/tendermint/tendermint/abci/types"
)

type ResponseAccountActive struct {
    Quotes []uint `json:"quotes"`
    Trades []uint `json:"trades"`
}

func (raa ResponseAccountActive) String() string {
    return strings.TrimSpace(fmt.Sprintf(`Quotes: %v
Trades: %v`, raa.Quotes, raa.Trades))
}

func queryAccountActive(ctx sdk.Context, path []string, 
    req abci.RequestQuery, keeper Keeper)(res []byte, err sdk.Error) {
    acct := path[0]
    data := keeper.GetAccountStatus(ctx, acct)
    quotes := make([]uint, len(data.ActiveQuotes.Data))
    trades := make([]uint, len(data.ActiveTrades.Data))
    for i := 0; i < len(data.ActiveQuotes.Data); i++ {
        quotes[i] = data.ActiveQuotes.Data[i].Id
    }
    for i := 0; i < len(data.ActiveTrades.Data); i++ {
        trades[i] = data.ActiveTrades.Data[i].Id
    }
    response := ResponseAccountActive {
        Quotes: quotes,
        Trades: trades,
    }
    
    bz, err2 := codec.MarshalJSONIndent(keeper.cdc, response)
    if err2 != nil {
        panic("could not marshal result to JSON")
    }
    
    return bz, nil
}