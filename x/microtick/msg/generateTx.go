package msg

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    "github.com/cosmos/cosmos-sdk/codec"
    "github.com/cosmos/cosmos-sdk/x/auth"
    abci "github.com/tendermint/tendermint/abci/types"
    
    mt "gitlab.com/microtick/mtzone/x/microtick/types"
    "gitlab.com/microtick/mtzone/x/microtick/keeper"
)

func GenerateTx(ctx sdk.Context, txType string, path []string, 
    req abci.RequestQuery, keeper keeper.Keeper) (res []byte, err error) {
        
    defer func() {
        if r := recover(); r != nil {
            err = mt.ErrInvalidRequest
        }
    }()
        
    var txmsg sdk.Msg
    
    acct := path[0]
    accAddr, _ := sdk.AccAddressFromBech32(acct)
    account := keeper.AccountKeeper.GetAccount(ctx, accAddr)
    if account == nil {
        return nil, sdkerrors.Wrap(mt.ErrInvalidAddress, acct)
    }
        
    switch txType {
    case "createquote":
        market := path[1]
        duration := path[2]
        backing := mt.NewMicrotickCoinFromString(path[3])
        spot := mt.NewMicrotickSpotFromString(path[4])
        premium := mt.NewMicrotickPremiumFromString(path[5])
        txmsg = NewTxCreateQuote(market, duration, accAddr, backing, spot, premium)
    case "cancelquote":
        id := mt.NewMicrotickIdFromString(path[1])
        txmsg = NewTxCancelQuote(id, accAddr)
    case "depositquote":
        id := mt.NewMicrotickIdFromString(path[1])
        amount := mt.NewMicrotickCoinFromString(path[2])
        txmsg = NewTxDepositQuote(id, accAddr, amount)
    case "withdrawquote":
        id := mt.NewMicrotickIdFromString(path[1])
        amount := mt.NewMicrotickCoinFromString(path[2])
        txmsg = NewTxWithdrawQuote(id, accAddr, amount)
    case "updatequote":
        id := mt.NewMicrotickIdFromString(path[1])
        spot := mt.NewMicrotickSpotFromString(path[2])
        premium := mt.NewMicrotickPremiumFromString(path[3])
        txmsg = NewTxUpdateQuote(id, accAddr, spot, premium)
    case "markettrade":
        market := path[1]
        duration := path[2]
        tradetype := mt.MicrotickTradeTypeFromName(path[3])
        quantity := mt.NewMicrotickQuantityFromString(path[4])
        txmsg = NewTxMarketTrade(market, duration, accAddr, tradetype, quantity)
    case "limittrade":
        market := path[1]
        duration := path[2]
        tradetype := mt.MicrotickTradeTypeFromName(path[3])
        limit := mt.NewMicrotickPremiumFromString(path[4])
        maxcost := mt.NewMicrotickCoinFromString(path[5])
        txmsg = NewTxLimitTrade(market, duration, accAddr, tradetype, limit, maxcost)
    case "picktrade":
        id := mt.NewMicrotickIdFromString(path[1])
        tradetype := mt.MicrotickTradeTypeFromName(path[2])
        txmsg = NewTxPickTrade(accAddr, id, tradetype)
    case "settletrade":
        id := mt.NewMicrotickIdFromString(path[1])
        txmsg = NewTxSettleTrade(id, accAddr)
    }
    
    response := mt.GenTx {
        Tx: auth.NewStdTx([]sdk.Msg{txmsg}, auth.NewStdFee(2000000, nil), nil, ""),
        AccountNumber: account.GetAccountNumber(),
        ChainID: ctx.ChainID(),
        Sequence: account.GetSequence(),
    }
    
    bz, _ := codec.MarshalJSONIndent(keeper.Cdc, response)
    
    return bz, nil
}
