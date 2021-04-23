package msg

import (
    "time"
    
    "github.com/gogo/protobuf/proto"
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    
    mt "github.com/microtick/mtzone/x/microtick/types"
    "github.com/microtick/mtzone/x/microtick/keeper"
)

func (msg TxWithdrawQuote) Route() string { return "microtick" }

func (msg TxWithdrawQuote) Type() string { return "withdraw" }

func (msg TxWithdrawQuote) ValidateBasic() error {
    if msg.Requester.Empty() {
        return sdkerrors.Wrap(mt.ErrInvalidAddress, msg.Requester.String())
    }
    return nil
}

func (msg TxWithdrawQuote) GetSignBytes() []byte {
    return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg TxWithdrawQuote) GetSigners() []sdk.AccAddress {
    return []sdk.AccAddress{ msg.Requester }
}

// Handler

func HandleTxWithdrawQuote(ctx sdk.Context, mtKeeper keeper.Keeper, params mt.MicrotickParams, 
    msg TxWithdrawQuote) (*sdk.Result, error) {
        
    withdraw := mt.NewMicrotickCoinFromString(msg.Withdraw)
        
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
    
    // Withdraw amount must be strictly less than quote backing (to withdraw the full amount, use CancelQUote)
    if withdraw.IsGTE(quote.Backing) {
        return nil, mt.ErrQuoteBacking
    }
    
    commission := mtKeeper.PoolCommission(ctx, withdraw.Amount.Mul(params.CommissionCancelPerunit))
    total := withdraw.Sub(commission)
    
    // Add coins to requester
    err = mtKeeper.DepositMicrotickCoin(ctx, msg.Requester, total)
    if err != nil {
        // Not sure why this error
        return nil, mt.ErrInsufficientFunds
    }
    
    dataMarket, err := mtKeeper.GetDataMarket(ctx, quote.Market)
    if err != nil {
        return nil, mt.ErrInvalidMarket
    }
    
    dataMarket.FactorOut(quote)
    
    quote.Backing = mt.NewMicrotickCoinFromDec(quote.Backing.Amount.Sub(withdraw.Amount))
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
    accountStatus.QuoteBacking = accountStatus.QuoteBacking.Sub(withdraw)
    mtKeeper.SetAccountStatus(ctx, msg.Requester, accountStatus)
    
    // Data
    data := WithdrawQuoteData {
      Time: now.Unix(),
      Market: dataMarket.Market,
      Duration: quote.DurationName,
      Consensus: dataMarket.Consensus,
      Backing: quote.Backing,
      Commission: commission,
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