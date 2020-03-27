package msg

import (
    "fmt"
    "time"
    "errors"
    
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
    Account string `json:"account"`
    Id mt.MicrotickId `json:"id"`
    Market mt.MicrotickMarket `json:"market"`
    Duration mt.MicrotickDurationName `json:"duration"`
    Consensus mt.MicrotickSpot `json:"consensus"`
    Time time.Time `json:"time"`
    Refund mt.MicrotickCoin `json:"refund"`
}

func (msg TxCancelQuote) Route() string { return "microtick" }

func (msg TxCancelQuote) Type() string { return "quote_cancel" }

func (msg TxCancelQuote) ValidateBasic() error {
    if msg.Requester.Empty() {
        return errors.New(fmt.Sprintf("Invalid address: %s", msg.Requester.String()))
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

func HandleTxCancelQuote(ctx sdk.Context, keeper keeper.Keeper, msg TxCancelQuote) (*sdk.Result, error) {
    quote, err := keeper.GetActiveQuote(ctx, msg.Id)
    if err != nil {
        return nil, errors.New(fmt.Sprintf("No such quote: %d", msg.Id))
    }
    
    // Time 2x invariant:
    // If a quote has not been updated within 2x the time duration of the quote, the
    // backing is forfeited.
    // Purpose: keeps market makers out of short-term orderbooks if they do not intend
    // to keep the quotes timely.
    if quote.Provider.String() != msg.Requester.String() {
        if !quote.Stale(ctx.BlockHeader().Time) {
            return nil, errors.New("Quote is not stale")
        }
    }
    
    if quote.Frozen(ctx.BlockHeader().Time) {
        return nil, errors.New(fmt.Sprintf("Quote is frozen until: %s", quote.CanModify))
    }
    
    // Everything ok, let's refund the backing and delete the quote
    err = keeper.DepositMicrotickCoin(ctx, msg.Requester, quote.Backing)
    if err != nil {
        return nil, errors.New("Fund mismatch")
    }
    
    dataMarket, _ := keeper.GetDataMarket(ctx, quote.Market)
    dataMarket.FactorOut(quote)
    dataMarket.DeleteQuote(quote)
    keeper.SetDataMarket(ctx, dataMarket)
    
    keeper.DeleteActiveQuote(ctx, quote.Id)
    
    accountStatus := keeper.GetAccountStatus(ctx, quote.Provider)
    accountStatus.QuoteBacking = accountStatus.QuoteBacking.Sub(quote.Backing)
    accountStatus.ActiveQuotes.Delete(quote.Id)
    keeper.SetAccountStatus(ctx, quote.Provider, accountStatus)
    
    // Data
    data := CancelQuoteData {
      Account: msg.Requester.String(),
      Id: quote.Id,
      Market: quote.Market,
      Duration: mt.MicrotickDurationNameFromDur(quote.Duration),
      Consensus: dataMarket.Consensus,
      Time: ctx.BlockHeader().Time,
      Refund: quote.Backing,
    }
    bz, _ := codec.MarshalJSONIndent(ModuleCdc, data)
    
    var events []sdk.Event
    events = append(events, sdk.NewEvent(
        sdk.EventTypeMessage,
        sdk.NewAttribute(sdk.AttributeKeyModule, mt.ModuleKey),
    ), sdk.NewEvent(
        sdk.EventTypeMessage,
        sdk.NewAttribute(fmt.Sprintf("quote.%d", quote.Id), "event.cancel"),
        sdk.NewAttribute(fmt.Sprintf("acct.%s", msg.Requester.String()), "quote.cancel"),
        sdk.NewAttribute("mtm.MarketTick", quote.Market),
    ))
    
    return &sdk.Result {
        Data: bz,
        Events: events,
    }, nil
}
