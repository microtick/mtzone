package msg

import (
    "time"
    
    "github.com/gogo/protobuf/proto"
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    
    mt "github.com/microtick/mtzone/x/microtick/types"
    "github.com/microtick/mtzone/x/microtick/keeper"
)

func (msg TxDepositQuote) Route() string { return "microtick" }

func (msg TxDepositQuote) Type() string { return "deposit" }

func (msg TxDepositQuote) ValidateBasic() error {
    if msg.Requester.Empty() {
        return sdkerrors.Wrap(mt.ErrInvalidAddress, msg.Requester.String())
    }
    return nil
}

func (msg TxDepositQuote) GetSignBytes() []byte {
    return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg TxDepositQuote) GetSigners() []sdk.AccAddress {
    return []sdk.AccAddress{ msg.Requester }
}

// Handler

func HandleTxDepositQuote(ctx sdk.Context, mtKeeper keeper.Keeper, params mt.MicrotickParams, 
    msg TxDepositQuote) (*sdk.Result, error) {
        
    deposit := mt.NewMicrotickCoinFromString(msg.Deposit)
        
    quote, err := mtKeeper.GetActiveQuote(ctx, msg.Id)
    if err != nil {
        return nil, sdkerrors.Wrapf(mt.ErrInvalidQuote, "%d", msg.Id)
    }
    
    if quote.Provider.String() != msg.Requester.String() {
        return nil, mt.ErrNotOwner
    }
    
    if quote.Frozen(ctx.BlockHeader().Time) {
        return nil, sdkerrors.Wrap(mt.ErrQuoteFrozen, time.Unix(quote.CanModify, 0).String())
    }

    dataMarket, err := mtKeeper.GetDataMarket(ctx, quote.Market)
    if err != nil {
        return nil, mt.ErrInvalidMarket
    }
    
    orderBook := dataMarket.GetOrderBook(quote.DurationName)
    adjustment := sdk.OneDec()
    if len(orderBook.CallAsks.Data) > 0 {
        bestCallAsk, _ := mtKeeper.GetActiveQuote(ctx, orderBook.CallAsks.Data[0].Id)
        bestPutAsk, _ := mtKeeper.GetActiveQuote(ctx, orderBook.PutAsks.Data[0].Id)
        average := bestCallAsk.CallAsk(dataMarket.Consensus).Amount.Add(bestPutAsk.PutAsk(dataMarket.Consensus).Amount).QuoInt64(2)
        if quote.Ask.Amount.GT(average) {
            adjustment = average.Quo(quote.Ask.Amount)
        }
    }
    
    commission := mtKeeper.PoolCommission(ctx, deposit.Amount.Mul(params.CommissionCreatePerunit).Quo(adjustment))
    total := deposit.Add(commission)
    
    // Subtract coins from requester
    err = mtKeeper.WithdrawMicrotickCoin(ctx, msg.Requester, total)
    if err != nil {
        return nil, mt.ErrInsufficientFunds
    }
    
    // Add commission to pool
    //fmt.Printf("Deposit Commission: %s\n", commission.String())
    reward, err := mtKeeper.AwardRebate(ctx, msg.Requester, deposit.Amount.Mul(params.MintRewardCreatePerunit).Mul(adjustment))
    if err != nil {
        return nil, err
    }
    
    dataMarket.FactorOut(quote)
    
    quote.Backing = mt.NewMicrotickCoinFromDec(quote.Backing.Amount.Add(deposit.Amount))
    quote.ComputeQuantity()
    
    // But we do freeze the new backing from any other updates
    now := ctx.BlockHeader().Time
    quote.Freeze(now, params)
    
    if !dataMarket.FactorIn(quote, true) {
        return nil, mt.ErrQuoteParams
    }
    mtKeeper.SetDataMarket(ctx, dataMarket)
    mtKeeper.SetActiveQuote(ctx, quote)
    
     // DataAccountStatus
    
    accountStatus := mtKeeper.GetAccountStatus(ctx, msg.Requester)
    accountStatus.QuoteBacking = accountStatus.QuoteBacking.Add(deposit)
    mtKeeper.SetAccountStatus(ctx, msg.Requester, accountStatus)
    
    // Data
    data := DepositQuoteData {
      Time: now.Unix(),
      Market: dataMarket.Market,
      Duration: quote.DurationName,
      Consensus: dataMarket.Consensus,
      Backing: quote.Backing,
      Commission: commission,
      Reward: *reward,
      Adjustment: adjustment.String(),
    }
    bz, err := proto.Marshal(&data)
    
    var events []sdk.Event
    events = append(events, sdk.NewEvent(
        sdk.EventTypeMessage,
        sdk.NewAttribute(sdk.AttributeKeyModule, mt.ModuleKey),
    ))
    
    ctx.EventManager().EmitEvents(events)
    
    return &sdk.Result {
        Data: bz,
        Events: ctx.EventManager().ABCIEvents(),
    }, nil
}
