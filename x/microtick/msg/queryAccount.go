package msg

import (
    "fmt"
    "strings"
    
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    abci "github.com/tendermint/tendermint/abci/types"
    
    mt "gitlab.com/microtick/mtzone/x/microtick/types"
    "gitlab.com/microtick/mtzone/x/microtick/keeper"
)

type ResponseAccountStatus struct {
    Account string `json:"account"`
    Balance mt.MicrotickCoin `json:"balance"`
    NumQuotes uint32 `json:"numQuotes"`
    NumTrades uint32 `json:"numTrades"`
    ActiveQuotes []mt.MicrotickId `json:"activeQuotes"`
    ActiveTrades []mt.MicrotickId `json:"activeTrades"`
    QuoteBacking mt.MicrotickCoin `json:"quoteBacking"`
    TradeBacking mt.MicrotickCoin `json:"tradeBacking"`
    SettleBacking mt.MicrotickCoin `json:"settleBacking"`
}

func (ras ResponseAccountStatus) String() string {
    return strings.TrimSpace(fmt.Sprintf(`Account: %s
Balance: %s
Num Quotes: %d
Num Trades: %d
Active Quotes: %v
Active Trades: %v
Quote Backing: %s
Trade Backing: %s
Settle Backing: %s`, ras.Account, 
    ras.Balance.String(),
    ras.NumQuotes, 
    ras.NumTrades, 
    ras.ActiveQuotes, ras.ActiveTrades,
    ras.QuoteBacking.String(), ras.TradeBacking.String(),
    ras.SettleBacking.String()))
}

func QueryAccountStatus(ctx sdk.Context, path []string, 
    req abci.RequestQuery, keeper keeper.Keeper) (res []byte, err error) {
    acct := path[0]
    address, err := sdk.AccAddressFromBech32(acct)
    balance := keeper.GetTotalBalance(ctx, address)
    data := keeper.GetAccountStatus(ctx, address)
    if err != nil {
        return nil, sdkerrors.Wrap(mt.ErrInvalidAddress, acct)
    }
    activeQuotes := make([]mt.MicrotickId, len(data.ActiveQuotes.Data))
    activeTrades := make([]mt.MicrotickId, len(data.ActiveTrades.Data))
    for i := 0; i < len(data.ActiveQuotes.Data); i++ {
        activeQuotes[i] = data.ActiveQuotes.Data[i].Id
    }
    for i := 0; i < len(data.ActiveTrades.Data); i++ {
        activeTrades[i] = data.ActiveTrades.Data[i].Id
    }
    response := ResponseAccountStatus {
        Account: acct,
        Balance: balance,
        NumQuotes: data.NumQuotes,
        NumTrades: data.NumTrades,
        ActiveQuotes: activeQuotes,
        ActiveTrades: activeTrades,
        QuoteBacking: data.QuoteBacking,
        TradeBacking: data.TradeBacking,
        SettleBacking: data.SettleBacking,
    }
    
    bz, err := codec.MarshalJSONIndent(keeper.Cdc, response)
    if err != nil {
        panic("Could not marshal result to JSON")
    }
    
    return bz, nil
}
