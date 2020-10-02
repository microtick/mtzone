package keeper

import (
    mt "gitlab.com/microtick/mtzone/x/microtick/types"
)

// DataAccountStatus

func NewDataAccountStatus(account mt.MicrotickAccount) DataAccountStatus {
    return DataAccountStatus {
        Account: account,
        ActiveQuotes: NewOrderedList(),
        ActiveTrades: NewOrderedList(),
        PlacedQuotes: 0,
        PlacedTrades: 0,
        QuoteBacking: mt.NewMicrotickCoinFromExtCoinInt(0),
        TradeBacking: mt.NewMicrotickCoinFromExtCoinInt(0),
        SettleBacking: mt.NewMicrotickCoinFromExtCoinInt(0),
    }
}

