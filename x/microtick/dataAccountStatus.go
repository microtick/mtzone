package microtick

// DataAccountStatus

type DataAccountStatus struct {
    Account MicrotickAccount `json:"account"`
    ActiveQuotes OrderedList `json:"activeQuotes"`
    ActiveTrades OrderedList `json:"activeTrades"`
    NumQuotes uint32 `json:"numQuotes"`
    NumTrades uint32 `json:"numTrades"`
    Change MicrotickCoin `json:"change"`
    QuoteBacking MicrotickCoin `json:"quoteBacking"`
    TradeBacking MicrotickCoin `json:"tradeBacking"`
}

func NewDataAccountStatus(account MicrotickAccount) DataAccountStatus {
    return DataAccountStatus {
        Account: account,
        ActiveQuotes: NewOrderedList(),
        ActiveTrades: NewOrderedList(),
        NumQuotes: 0,
        NumTrades: 0,
        Change: NewMicrotickCoinFromInt(0),
        QuoteBacking: NewMicrotickCoinFromInt(0),
        TradeBacking: NewMicrotickCoinFromInt(0),
    }
}

