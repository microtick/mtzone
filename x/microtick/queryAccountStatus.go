package microtick 

import (
    "fmt"
    "strings"
    
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
    abci "github.com/tendermint/tendermint/abci/types"
)

type ResponseAccountStatus struct {
    Account string `json:"account"`
    NumQuotes uint32 `json:"numQuotes"`
    NumTrades uint32 `json:"numTrades"`
    QuoteBacking sdk.Coins `json:"quoteBacking"`
    TradeBacking sdk.Coins `json:"tradeBacking"`
}

func (ras ResponseAccountStatus) String() string {
    return strings.TrimSpace(fmt.Sprintf(`Account: %s
NumQuotes: %s
NumTrades: %s
QuoteBacking: %s
TradeBacking: %s`, ras.Account, 
    fmt.Sprintf("%d", ras.NumQuotes), 
    fmt.Sprintf("%d", ras.NumTrades), 
    ras.QuoteBacking, ras.TradeBacking))
}

func queryAccountStatus(ctx sdk.Context, path []string, 
    req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
    acct := path[0]
    data := keeper.GetAccountStatus(ctx, acct)
    response := ResponseAccountStatus {
        Account: acct,
        NumQuotes: data.NumQuotes,
        NumTrades: data.NumTrades,
        QuoteBacking: data.QuoteBacking,
        TradeBacking: data.TradeBacking,
    }
    
    bz, err2 := codec.MarshalJSONIndent(keeper.cdc, response)
    if err2 != nil {
        panic("could not marshal result to JSON")
    }
    
    return bz, nil
}
