package microtick

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
)

// DataAccountStatus

type DataAccountStatus struct {
    Account MicrotickAccount `json:"account"`
    NumQuotes uint32 `json:"numQuotes"`
    NumTrades uint32 `json:"numTrades"`
    QuoteBacking sdk.Coins `json:"quoteBacking"`
    TradeBacking sdk.Coins `json:"tradeBacking"`
}

func NewDataAccountStatus(account MicrotickAccount) DataAccountStatus {
    return DataAccountStatus {
        Account: account,
        NumQuotes: 0,
        NumTrades: 0,
        QuoteBacking: sdk.Coins{sdk.NewInt64Coin(TokenType, 0)},
        TradeBacking: sdk.Coins{sdk.NewInt64Coin(TokenType, 0)},
    }
}

