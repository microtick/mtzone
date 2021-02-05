package msg

import (
    "github.com/cosmos/cosmos-sdk/codec"
   	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
    cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
    govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

var (
    //ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
    amino = codec.NewLegacyAmino()
    ModuleCdc = codec.NewAminoCodec(amino)
    
	_, _, _, _ sdk.Msg = &TxCancelQuote{}, &TxCreateQuote{}, &TxUpdateQuote{}, &TxDepositQuote{}
	_, _, _, _ sdk.Msg = &TxWithdrawQuote{}, &TxMarketTrade{}, &TxPickTrade{}, &TxSettleTrade{}
)

func init() {
    RegisterLegacyAminoCodec(amino)
    cryptocodec.RegisterCrypto(amino)
    amino.Seal()
}

// Register concrete types on codec codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
    cdc.RegisterConcrete(TxCreateQuote{}, "microtick/Create", nil)
    cdc.RegisterConcrete(TxCancelQuote{}, "microtick/Cancel", nil)
    cdc.RegisterConcrete(TxUpdateQuote{}, "microtick/Update", nil)
    cdc.RegisterConcrete(TxDepositQuote{}, "microtick/Deposit", nil)
    cdc.RegisterConcrete(TxWithdrawQuote{}, "microtick/Withdraw", nil)
    cdc.RegisterConcrete(TxMarketTrade{}, "microtick/Trade", nil)
    cdc.RegisterConcrete(TxPickTrade{}, "mic rotick/Pick", nil)
    cdc.RegisterConcrete(TxSettleTrade{}, "microtick/Settle", nil)
    cdc.RegisterConcrete(DenomChangeProposal{}, "microtick/DenomChangeProposal", nil)
    cdc.RegisterConcrete(AddMarketsProposal{}, "microtick/AddMarketsProposal", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
  registry.RegisterImplementations((*sdk.Msg)(nil),
    &TxCancelQuote{},
    &TxCreateQuote{},
    &TxUpdateQuote{},
    &TxDepositQuote{},
    &TxWithdrawQuote{},
    &TxMarketTrade{},
    &TxPickTrade{},
    &TxSettleTrade{},
  )
  registry.RegisterImplementations(
    (*govtypes.Content)(nil),
    &DenomChangeProposal{},
    &AddMarketsProposal{},
  )
}
