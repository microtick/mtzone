package microtick

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/codec"
    "github.com/cosmos/cosmos-sdk/x/auth"
    abci "github.com/tendermint/tendermint/abci/types"
)

func generateTx(ctx sdk.Context, txType string, path []string, 
    req abci.RequestQuery, keeper Keeper) (res []byte, err sdk.Error) {
        
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
        
    var msg sdk.Msg
    
    acct := path[0]
    accAddr, _ := sdk.AccAddressFromBech32(acct)
    account := keeper.accountKeeper.GetAccount(ctx, accAddr)
    if account == nil {
        return nil, sdk.ErrInternal("No such address")
    }
        
    switch txType {
    case "createmarket":
        market := path[1]
        msg = NewTxCreateMarket(accAddr, market)
    case "createquote":
        market := path[1]
        duration := MicrotickDurationFromName(path[2])
        backing := NewMicrotickCoinFromString(path[3])
        spot := NewMicrotickSpotFromString(path[4])
        premium := NewMicrotickPremiumFromString(path[5])
        msg = NewTxCreateQuote(market, duration, accAddr, backing, spot, premium)
    case "cancelquote":
        id := NewMicrotickIdFromString(path[1])
        msg = NewTxCancelQuote(id, accAddr)
    case "depositquote":
        id := NewMicrotickIdFromString(path[1])
        amount := NewMicrotickCoinFromString(path[2])
        msg = NewTxDepositQuote(id, accAddr, amount)
    case "updatequote":
        id := NewMicrotickIdFromString(path[1])
        spot := NewMicrotickSpotFromString(path[2])
        premium := NewMicrotickPremiumFromString(path[3])
        msg = NewTxUpdateQuote(id, accAddr, spot, premium)
    case "markettrade":
        market := path[1]
        duration := MicrotickDurationFromName(path[2])
        tradetype := MicrotickTradeTypeFromName(path[3])
        quantity := NewMicrotickQuantityFromString(path[4])
        msg = NewTxMarketTrade(market, duration, accAddr, tradetype, quantity)
    case "limittrade":
        market := path[1]
        duration := MicrotickDurationFromName(path[2])
        tradetype := MicrotickTradeTypeFromName(path[3])
        limit := NewMicrotickPremiumFromString(path[4])
        maxcost := NewMicrotickCoinFromString(path[5])
        msg = NewTxLimitTrade(market, duration, accAddr, tradetype, limit, maxcost)
    case "settletrade":
        id := NewMicrotickIdFromString(path[1])
        msg = NewTxSettleTrade(id, accAddr)
    }
        
    response := GenTx {
        Tx: auth.NewStdTx([]sdk.Msg{msg}, auth.NewStdFee(1000000, nil), nil, ""),
        AccountNumber: account.GetAccountNumber(),
        ChainID: ctx.ChainID(),
        Sequence: account.GetSequence(),
    }
    
    bz, _ := codec.MarshalJSONIndent(keeper.cdc, response)
    
    return bz, nil
}
