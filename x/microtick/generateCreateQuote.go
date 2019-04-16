package microtick

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/codec"
    "github.com/cosmos/cosmos-sdk/x/auth"
    abci "github.com/tendermint/tendermint/abci/types"
)

func generateTxCreateQuote(ctx sdk.Context, path []string,
    req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
        
    acct := path[0]
    market := path[1]
    duration := NewMicrotickDurationFromString(path[2])
    backing := NewMicrotickCoinFromString(path[3])
    spot := NewMicrotickSpotFromString(path[4])
    premium := NewMicrotickPremiumFromString(path[5])
    
    accAddr, _ := sdk.AccAddressFromBech32(acct)
    
    account := keeper.accountKeeper.GetAccount(ctx, accAddr)
    if account == nil {
        return nil, sdk.ErrInternal("No such address")
    }
    
    msg := NewTxCreateQuote(market, duration, accAddr, backing, spot, premium)
    response := GenTx {
        Tx: auth.NewStdTx([]sdk.Msg{msg}, auth.NewStdFee(200000, nil), nil, ""),
        AccountNumber: account.GetAccountNumber(),
        ChainID: ctx.ChainID(),
        Sequence: account.GetSequence(),
    }
    
    bz, _ := codec.MarshalJSONIndent(keeper.cdc, response)
    
    return bz, nil
}
