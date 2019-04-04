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
    Balance MicrotickCoin `json:"balance"`
    Change MicrotickCoin `json:"change"`
    NumQuotes uint32 `json:"numQuotes"`
    NumTrades uint32 `json:"numTrades"`
    ActiveQuotes []uint `json:"activeQuotes"`
    ActiveTrades []uint `json:"activeTrades"`
    QuoteBacking MicrotickCoin `json:"quoteBacking"`
    TradeBacking MicrotickCoin `json:"tradeBacking"`
}

func (ras ResponseAccountStatus) String() string {
    return strings.TrimSpace(fmt.Sprintf(`Account: %s
Balance: %s
Change: %s
NumQuotes: %d
NumTrades: %d
ActiveQuotes: %v
ActiveTrades: %v
QuoteBacking: %s
TradeBacking: %s`, ras.Account, 
    ras.Balance.String(),
    ras.Change,
    ras.NumQuotes, 
    ras.NumTrades, 
    ras.ActiveQuotes, ras.ActiveTrades,
    ras.QuoteBacking.String(), ras.TradeBacking.String()))
}

func queryAccountStatus(ctx sdk.Context, path []string, 
    req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
    acct := path[0]
    data := keeper.GetAccountStatus(ctx, acct)
    address, err2 := sdk.AccAddressFromBech32(acct)
    if err2 != nil {
        panic("could not get address balance")
    }
    coins := keeper.coinKeeper.GetCoins(ctx, address)
    balance := data.Change
    for i := 0; i < len(coins); i++ {
        if coins[i].Denom == TokenType {
            balance = balance.Plus(sdk.NewDecCoin(TokenType, coins[i].Amount))
        }
    }
    activeQuotes := make([]uint, len(data.ActiveQuotes.Data))
    activeTrades := make([]uint, len(data.ActiveTrades.Data))
    for i := 0; i < len(data.ActiveQuotes.Data); i++ {
        activeQuotes[i] = data.ActiveQuotes.Data[i].Id
    }
    for i := 0; i < len(data.ActiveTrades.Data); i++ {
        activeTrades[i] = data.ActiveTrades.Data[i].Id
    }
    response := ResponseAccountStatus {
        Account: acct,
        Balance: balance,
        Change: data.Change,
        NumQuotes: data.NumQuotes,
        NumTrades: data.NumTrades,
        ActiveQuotes: activeQuotes,
        ActiveTrades: activeTrades,
        QuoteBacking: data.QuoteBacking,
        TradeBacking: data.TradeBacking,
    }
    
    bz, err2 := codec.MarshalJSONIndent(keeper.cdc, response)
    if err2 != nil {
        panic("could not marshal result to JSON")
    }
    
    return bz, nil
}
