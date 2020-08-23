package msg

import (
    "github.com/cosmos/cosmos-sdk/codec"
    "github.com/cosmos/cosmos-sdk/codec/types"
)

// Register concrete types on codec codec
func RegisterCodec(cdc *codec.Codec) {
    cdc.RegisterConcrete(TxCreateQuote{}, "microtick/QuoteCreate", nil)
    cdc.RegisterConcrete(TxCancelQuote{}, "microtick/QuoteCancel", nil)
    cdc.RegisterConcrete(TxUpdateQuote{}, "microtick/QuoteUpdate", nil)
    cdc.RegisterConcrete(TxDepositQuote{}, "microtick/QuoteDeposit", nil)
    cdc.RegisterConcrete(TxWithdrawQuote{}, "microtick/QuoteWithdraw", nil)
    cdc.RegisterConcrete(TxMarketTrade{}, "microtick/TradeMarket", nil)
    cdc.RegisterConcrete(TxPickTrade{}, "mic rotick/TradePick", nil)
    cdc.RegisterConcrete(TxSettleTrade{}, "microtick/TradeSettle", nil)
}

// generic sealed codec to be used throughout this module
var (
    amino = codec.New()
    ModuleCdc = codec.NewHybridCodec(amino, types.NewInterfaceRegistry())
)

func init() {
    RegisterCodec(amino)
    codec.RegisterCrypto(amino)
    amino.Seal()
}