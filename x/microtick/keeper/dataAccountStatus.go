package keeper

import (
    mt "gitlab.com/microtick/mtzone/x/microtick/types"
)

// DataAccountStatus

type DataAccountStatus struct {
    Account mt.MicrotickAccount `json:"account"`
    ActiveQuotes OrderedList `json:"activeQuotes"`
    ActiveTrades OrderedList `json:"activeTrades"`
    NumQuotes uint32 `json:"numQuotes"`
    NumTrades uint32 `json:"numTrades"`
    QuoteBacking mt.MicrotickCoin `json:"quoteBacking"`
    TradeBacking mt.MicrotickCoin `json:"tradeBacking"`
    SettleBacking mt.MicrotickCoin `json:"settleBacking"`
}

func NewDataAccountStatus(account mt.MicrotickAccount) DataAccountStatus {
    return DataAccountStatus {
        Account: account,
        ActiveQuotes: NewOrderedList(),
        ActiveTrades: NewOrderedList(),
        NumQuotes: 0,
        NumTrades: 0,
        QuoteBacking: mt.NewMicrotickCoinFromExtCoinInt(0),
        TradeBacking: mt.NewMicrotickCoinFromExtCoinInt(0),
        SettleBacking: mt.NewMicrotickCoinFromExtCoinInt(0),
    }
}

