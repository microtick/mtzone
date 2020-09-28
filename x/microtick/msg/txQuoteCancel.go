package msg

import (
    "fmt"
    "time"
    
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    
    mt "github.com/mjackson001/mtzone/x/microtick/types"
    "github.com/mjackson001/mtzone/x/microtick/keeper"
)

func NewTxCancelQuote(id mt.MicrotickId, requester sdk.AccAddress) TxCancelQuote {
    return TxCancelQuote {
        Id: id,
        Requester: requester,
    }
}

func (msg TxCancelQuote) Route() string { return "microtick" }

func (msg TxCancelQuote) Type() string { return "quote_cancel" }

func (msg TxCancelQuote) ValidateBasic() error {
    if msg.Requester.Empty() {
        return sdkerrors.Wrap(mt.ErrInvalidAddress, msg.Requester.String())
    }
    return nil
}

func (msg TxCancelQuote) GetSignBytes() []byte {
    return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg TxCancelQuote) GetSigners() []sdk.AccAddress {
    return []sdk.AccAddress{ msg.Requester }
}

// Handler

func HandleTxCancelQuote(ctx sdk.Context, mtKeeper keeper.Keeper, params mt.MicrotickParams,
    msg TxCancelQuote) (*sdk.Result, error) {
        
    quote, err := mtKeeper.GetActiveQuote(ctx, msg.Id)
    if err != nil {
        return nil, sdkerrors.Wrapf(mt.ErrInvalidQuote, "%d", msg.Id)
    }
    
    if quote.Frozen(ctx.BlockHeader().Time) {
        return nil, sdkerrors.Wrap(mt.ErrQuoteFrozen, time.Unix(quote.CanModify, 0).String())
    }
    
    // Commission
    commission := mt.NewMicrotickCoinFromDec(quote.Backing.Amount.Mul(params.CommissionCancelPercent))
    err = mtKeeper.WithdrawMicrotickCoin(ctx, msg.Requester, commission)
    if err != nil {
        return nil, mt.ErrInsufficientFunds
    }
    _, err = mtKeeper.PoolCommission(ctx, msg.Requester, commission, false)
    if err != nil {
        return nil, err
    }
    
    // Time 2x invariant:
    // If a quote has not been updated within 2x the time duration of the quote, the
    // backing is slashed and awarded to the canceler
    // Purpose: keeps market makers out of short-term orderbooks if they do not intend
    // to keep the quotes timely.
    slash := sdk.ZeroDec()
    // Save original quote backing
    backing := quote.Backing
    if quote.Provider.String() != msg.Requester.String() {
        if !quote.Stale(ctx.BlockHeader().Time) {
            return nil, mt.ErrQuoteNotStale
        }
        slash := quote.Backing.Amount.Mul(params.CancelSlashRate)
        backing.Amount = backing.Amount.Sub(slash)
        err = mtKeeper.DepositMicrotickCoin(ctx, msg.Requester, mt.NewMicrotickCoinFromDec(slash))
        if err != nil {
            return nil, sdkerrors.Wrap(mt.ErrQuoteBacking, "slash deposit")
        }
    }
    
    // Everything ok, let's refund the (remainder of the) backing and delete the quote
    err = mtKeeper.DepositMicrotickCoin(ctx, quote.Provider, backing)
    if err != nil {
        return nil, mt.ErrQuoteBacking
    }
    
    dataMarket, err := mtKeeper.GetDataMarket(ctx, quote.Market)
    if err != nil {
        return nil, mt.ErrInvalidMarket
    }
    
    dataMarket.FactorOut(quote)
    dataMarket.DeleteQuote(quote)
    mtKeeper.SetDataMarket(ctx, dataMarket)
    
    mtKeeper.DeleteActiveQuote(ctx, quote.Id)
    
    accountStatus := mtKeeper.GetAccountStatus(ctx, quote.Provider)
    accountStatus.QuoteBacking = accountStatus.QuoteBacking.Sub(quote.Backing)
    accountStatus.ActiveQuotes.Delete(quote.Id)
    mtKeeper.SetAccountStatus(ctx, quote.Provider, accountStatus)
    
    // Data
    data := CancelQuoteData {
      Account: msg.Requester,
      Id: quote.Id,
      Market: quote.Market,
      Duration: quote.DurationName,
      Consensus: dataMarket.Consensus,
      Time: ctx.BlockHeader().Time.Unix(),
      Refund: quote.Backing,
      Slash: mt.NewMicrotickCoinFromDec(slash),
      Commission: commission,
    }
    bz := ModuleCdc.MustMarshalJSON(&data)
    
    var events []sdk.Event
    events = append(events, sdk.NewEvent(
        sdk.EventTypeMessage,
        sdk.NewAttribute(sdk.AttributeKeyModule, mt.ModuleKey),
    ), sdk.NewEvent(
        sdk.EventTypeMessage,
        sdk.NewAttribute(fmt.Sprintf("quote.%d", quote.Id), "event.cancel"),
        sdk.NewAttribute(fmt.Sprintf("acct.%s", msg.Requester.String()), "quote.cancel"),
        sdk.NewAttribute("mtm.MarketTick", quote.Market),
    ), sdk.NewEvent(
        sdk.EventTypeMessage,
        sdk.NewAttribute("commission", commission.String()),
    ))
    
    if quote.Provider.String() != msg.Requester.String() {
        events = append(events, sdk.NewEvent(
            sdk.EventTypeMessage,
            sdk.NewAttribute(fmt.Sprintf("acct.%s", quote.Provider.String()), "quote.cancel"),
        ))
    }
    
    ctx.EventManager().EmitEvents(events)
    
    return &sdk.Result {
        Data: bz,
        Events: ctx.EventManager().ABCIEvents(),
    }, nil
}
