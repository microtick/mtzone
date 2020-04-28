package msg

import (
    "github.com/cosmos/cosmos-sdk/codec"
)

// Register concrete types on codec codec
func RegisterCodec(cdc *codec.Codec) {
    cdc.RegisterConcrete(TxCreateMarket{}, "microtick/CreateMarket", nil)
    cdc.RegisterConcrete(TxCreateQuote{}, "microtick/CreateQuote", nil)
    cdc.RegisterConcrete(TxCancelQuote{}, "microtick/CancelQuote", nil)
    cdc.RegisterConcrete(TxUpdateQuote{}, "microtick/UpdateQuote", nil)
    cdc.RegisterConcrete(TxDepositQuote{}, "microtick/DepositQuote", nil)
    cdc.RegisterConcrete(TxWithdrawQuote{}, "microtick/WithdrawQuote", nil)
    cdc.RegisterConcrete(TxMarketTrade{}, "microtick/MarketTrade", nil)
    cdc.RegisterConcrete(TxLimitTrade{}, "microtick/LimitTrade", nil)
    cdc.RegisterConcrete(TxPickTrade{}, "microtick/PickTrade", nil)
    cdc.RegisterConcrete(TxSettleTrade{}, "microtick/SettleTrade", nil)
}

// generic sealed codec to be used throughout this module
var (
    amino = codec.New()
    ModuleCdc = codec.NewHybridCodec(amino)
)

func init() {
    RegisterCodec(amino)
    codec.RegisterCrypto(amino)
    amino.Seal()
}