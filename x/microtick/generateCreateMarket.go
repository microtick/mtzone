package microtick

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/x/auth"
    abci "github.com/tendermint/tendermint/abci/types"
)


func generateTxCreateMarket(ctx sdk.Context, path []string, 
    req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
        
    acct := path[0]
    market := path[1]
    
    accAddr, _ := sdk.AccAddressFromBech32(acct)
    msg := NewTxCreateMarket(accAddr, market)
    
    account := keeper.accountKeeper.GetAccount(ctx, accAddr)
    if account == nil {
        return nil, sdk.ErrInternal("No such address")
    }
    
    bz := auth.StdSignBytes(ctx.ChainID(), account.GetAccountNumber(), account.GetSequence(), 
        auth.StdFee{}, []sdk.Msg{msg}, "")
    
    return bz, nil
}