package microtick

import (
    "fmt"
    "strings"
    
    sdk "github.com/cosmos/cosmos-sdk/types"
)

const TokenType = "mt$"

// AccountInfo

type AccountInfo struct {
    Owner string `json:"owner"`
    NumQuotes int `json:"numQuotes"`
    NumTrades int `json:"numTrades"`
    QuoteBacking sdk.Coins `json:"quoteBacking"`
    TradeBacking sdk.Coins `json:"tradeBacking"`
}

func NewAccountInfo(owner string) AccountInfo {
    return AccountInfo{
        Owner: owner,
        NumQuotes: 0,
        NumTrades: 0,
        QuoteBacking: sdk.Coins{sdk.NewInt64Coin(TokenType, 0)},
        TradeBacking: sdk.Coins{sdk.NewInt64Coin(TokenType, 0)},
    }
}

func (ai AccountInfo) String() string {
    return strings.TrimSpace(fmt.Sprintf(`Owner: %s
NumQuotes: %s
NumTrades: %s
QuoteBacking: %s
TradeBacking: %s`, ai.Owner, ai.NumQuotes, ai.NumTrades, ai.QuoteBacking, ai.TradeBacking))
}
