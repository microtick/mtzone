package msg

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/codec"
    "github.com/cosmos/cosmos-sdk/x/auth"
    abci "github.com/tendermint/tendermint/abci/types"
    
    mt "github.com/mjackson001/mtzone/x/microtick/types"
    "github.com/mjackson001/mtzone/x/microtick/keeper"
)

func GenerateTx(ctx sdk.Context, txType string, path []string, 
    req abci.RequestQuery, keeper keeper.MicrotickKeeper) (res []byte, err sdk.Error) {
        
    defer func() {
        if r := recover(); r != nil {
            switch x := r.(type) {
            case string:
                err = sdk.ErrInternal(x)
            default:
                err = sdk.ErrInternal("Unknown error")
            }
        }
    }()
        
    var txmsg sdk.Msg
    
    acct := path[0]
    accAddr, _ := sdk.AccAddressFromBech32(acct)
    account := keeper.AccountKeeper.GetAccount(ctx, accAddr)
    if account == nil {
        return nil, sdk.ErrInternal("No such address")
    }
        
    switch txType {
    case "createmarket":
        market := path[1]
        txmsg = NewTxCreateMarket(accAddr, market)
    case "createquote":
        market := path[1]
        duration := mt.MicrotickDurationFromName(path[2])
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
    case "updatequote":
        id := mt.NewMicrotickIdFromString(path[1])
        spot := mt.NewMicrotickSpotFromString(path[2])
        premium := mt.NewMicrotickPremiumFromString(path[3])
        txmsg = NewTxUpdateQuote(id, accAddr, spot, premium)
    case "markettrade":
        market := path[1]
        duration := mt.MicrotickDurationFromName(path[2])
        tradetype := mt.MicrotickTradeTypeFromName(path[3])
        quantity := mt.NewMicrotickQuantityFromString(path[4])
        txmsg = NewTxMarketTrade(market, duration, accAddr, tradetype, quantity)
    case "limittrade":
        market := path[1]
        duration := mt.MicrotickDurationFromName(path[2])
        tradetype := mt.MicrotickTradeTypeFromName(path[3])
        limit := mt.NewMicrotickPremiumFromString(path[4])
        maxcost := mt.NewMicrotickCoinFromString(path[5])
        txmsg = NewTxLimitTrade(market, duration, accAddr, tradetype, limit, maxcost)
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
    
    bz, _ := codec.MarshalJSONIndent(ModuleCdc, response)
    
    return bz, nil
}
