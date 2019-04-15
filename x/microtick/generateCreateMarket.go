package microtick

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/codec"
    "github.com/cosmos/cosmos-sdk/x/auth"
    abci "github.com/tendermint/tendermint/abci/types"
)

type GenTx struct {
    Tx auth.StdTx `json:"tx"`
    AccountNumber uint64 `json:"accountNumber"`
    ChainID string `json:"chainId"`
    Sequence uint64 `json:"sequence"`
}

func generateTxCreateMarket(ctx sdk.Context, path []string, 
    req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
        
    acct := path[0]
    market := path[1]
    
    accAddr, _ := sdk.AccAddressFromBech32(acct)
    
    account := keeper.accountKeeper.GetAccount(ctx, accAddr)
    if account == nil {
        return nil, sdk.ErrInternal("No such address")
    }
    
    msg := NewTxCreateMarket(accAddr, market)
    response := GenTx {
        Tx: auth.NewStdTx([]sdk.Msg{msg}, auth.NewStdFee(200000, nil), nil, ""),
        AccountNumber: account.GetAccountNumber(),
        ChainID: ctx.ChainID(),
        Sequence: account.GetSequence(),
    }
    
    bz, _ := codec.MarshalJSONIndent(keeper.cdc, response)
    
    //bz := auth.StdSignBytes(ctx.ChainID(), account.GetAccountNumber(), account.GetSequence(), 
        //auth.NewStdFee(200000, sdk.Coins{}), []sdk.Msg{msg}, "")
    
    return bz, nil
}
