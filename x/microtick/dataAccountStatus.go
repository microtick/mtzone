package microtick

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
)

// DataAccountStatus

type DataAccountStatus struct {
    Account MicrotickAccount `json:"account"`
    ActiveQuotes OrderedList `json:"activeQuotes"`
    ActiveTrades OrderedList `json:"activeTrades"`
    NumQuotes uint32 `json:"numQuotes"`
    NumTrades uint32 `json:"numTrades"`
    QuoteBacking sdk.Coins `json:"quoteBacking"`
    TradeBacking sdk.Coins `json:"tradeBacking"`
}

func quoteCompare(x, y ListItem) int {
    x1 := x.(DataActiveQuote)
    y1 := y.(DataActiveQuote)
    return int(x1.Id) - int(y1.Id)
}

func tradeCompare(x, y ListItem) int {
    x1 := x.(DataActiveQuote)
    y1 := y.(DataActiveQuote)
    return int(x1.Id) - int(y1.Id)
}

func NewDataAccountStatus(account MicrotickAccount) DataAccountStatus {
    return DataAccountStatus {
        Account: account,
        ActiveQuotes: NewOrderedList(quoteCompare),
        ActiveTrades: NewOrderedList(tradeCompare),
        NumQuotes: 0,
        NumTrades: 0,
        QuoteBacking: sdk.Coins{sdk.NewInt64Coin(TokenType, 0)},
        TradeBacking: sdk.Coins{sdk.NewInt64Coin(TokenType, 0)},
    }
}

