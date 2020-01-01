package msg

import (
    "fmt"
    "time"
    
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
    
    mt "github.com/mjackson001/mtzone/x/microtick/types"
    "github.com/mjackson001/mtzone/x/microtick/keeper"
)

type TxCancelQuote struct {
    Id mt.MicrotickId
    Requester mt.MicrotickAccount
}

func NewTxCancelQuote(id mt.MicrotickId, requester sdk.AccAddress) TxCancelQuote {
    return TxCancelQuote {
        Id: id,
        Requester: requester,
    }
}

type CancelQuoteData struct {
    Id mt.MicrotickId `json:"id"`
    Originator string `json:"originator"`
    Market mt.MicrotickMarket `json:"market"`
    Duration mt.MicrotickDurationName `json:"duration"`
    Consensus mt.MicrotickSpot `json:"consensus"`
    Time time.Time `json:"time"`
    Refund mt.MicrotickCoin `json:"refund"`
    Balance mt.MicrotickCoin `json:"balance"`
}

func (msg TxCancelQuote) Route() string { return "microtick" }

func (msg TxCancelQuote) Type() string { return "cancel_quote" }

func (msg TxCancelQuote) ValidateBasic() sdk.Error {
    if msg.Requester.Empty() {
        return sdk.ErrInvalidAddress(msg.Requester.String())
    }
    return nil
}

func (msg TxCancelQuote) GetSignBytes() []byte {
    return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg TxCancelQuote) GetSigners() []sdk.AccAddress {
    return []sdk.AccAddress{ msg.Requester }
}

// Handler

func HandleTxCancelQuote(ctx sdk.Context, keeper keeper.Keeper, msg TxCancelQuote) sdk.Result {
    quote, err := keeper.GetActiveQuote(ctx, msg.Id)
    if err != nil {
        return sdk.ErrInternal(fmt.Sprintf("No such quote: %d", msg.Id)).Result()
    }
    
    // Time 2x invariant:
    // If a quote has not been updated within 2x the time duration of the quote, the
    // backing is forfeited.
    // Purpose: keeps market makers out of short-term orderbooks if they do not intend
    // to keep the quotes timely.
    if quote.Provider.String() != msg.Requester.String() {
        if !quote.Stale(ctx.BlockHeader().Time) {
            return sdk.ErrInternal("Quote is not stale").Result()
        }
    }
    
    if quote.Frozen(ctx.BlockHeader().Time) {
        return sdk.ErrInternal(fmt.Sprintf("Quote is frozen until: %s", quote.CanModify)).Result()
    }
    
    // Everything ok, let's refund the backing and delete the quote
    keeper.DepositMicrotickCoin(ctx, msg.Requester, quote.Backing)
    
    dataMarket, _ := keeper.GetDataMarket(ctx, quote.Market)
    dataMarket.FactorOut(quote)
    dataMarket.DeleteQuote(quote)
    keeper.SetDataMarket(ctx, dataMarket)
    
    keeper.DeleteActiveQuote(ctx, quote.Id)
    
    accountStatus := keeper.GetAccountStatus(ctx, quote.Provider)
    accountStatus.QuoteBacking = accountStatus.QuoteBacking.Sub(quote.Backing)
    accountStatus.ActiveQuotes.Delete(quote.Id)
    keeper.SetAccountStatus(ctx, quote.Provider, accountStatus)
    
    balance := keeper.GetTotalBalance(ctx, quote.Provider)
    
    event := sdk.NewEvent(
        sdk.EventTypeMessage,
        sdk.NewAttribute(fmt.Sprintf("quote.%d", quote.Id), "event.cancel"),
        sdk.NewAttribute(fmt.Sprintf("acct.%s", msg.Requester.String()), "quote.cancel"),
        sdk.NewAttribute("mtm.MarketTick", quote.Market),
    )
    
    // Data
    data := CancelQuoteData {
      Id: quote.Id,
      Originator: "cancelQuote",
      Market: quote.Market,
      Duration: mt.MicrotickDurationNameFromDur(quote.Duration),
      Consensus: dataMarket.Consensus,
      Time: ctx.BlockHeader().Time,
      Refund: quote.Backing,
      Balance: balance,
    }
    bz, _ := codec.MarshalJSONIndent(ModuleCdc, data)
    
    return sdk.Result {
        Data: bz,
        Events: []sdk.Event{ event },
    }
}
