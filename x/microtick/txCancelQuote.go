package microtick

import (
    "fmt"
    "time"
    
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
)

type TxCancelQuote struct {
    Id MicrotickId
    Requester MicrotickAccount
}

func NewTxCancelQuote(id MicrotickId, requester sdk.AccAddress) TxCancelQuote {
    return TxCancelQuote {
        Id: id,
        Requester: requester,
    }
}

type CancelQuoteData struct {
    Id MicrotickId `json:"id"`
    Originator string `json:"originator"`
    Consensus MicrotickSpot `json:"consensus"`
    Time time.Time `json:"time"`
    Refund MicrotickCoin `json:"refund"`
    Balance MicrotickCoin `json:"balance"`
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
    return sdk.MustSortJSON(msgCdc.MustMarshalJSON(msg))
}

func (msg TxCancelQuote) GetSigners() []sdk.AccAddress {
    return []sdk.AccAddress{ msg.Requester }
}

// Handler

func handleTxCancelQuote(ctx sdk.Context, keeper Keeper, msg TxCancelQuote) sdk.Result {
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
    dataMarket.factorOut(quote)
    dataMarket.DeleteQuote(quote)
    keeper.SetDataMarket(ctx, dataMarket)
    
    keeper.DeleteActiveQuote(ctx, quote.Id)
    
    accountStatus := keeper.GetAccountStatus(ctx, quote.Provider)
    accountStatus.QuoteBacking = accountStatus.QuoteBacking.Sub(quote.Backing)
    accountStatus.ActiveQuotes.Delete(quote.Id)
    keeper.SetAccountStatus(ctx, quote.Provider, accountStatus)
    
    balance := keeper.GetTotalBalance(ctx, quote.Provider)
    
    tags := sdk.NewTags(
        fmt.Sprintf("quote.%d", quote.Id), "event.cancel",
        fmt.Sprintf("acct.%s", msg.Requester.String()), "quote.cancel",
        "mtm.MarketTick", quote.Market,
    )
    
    // Data
    data := CancelQuoteData {
      Id: quote.Id,
      Originator: "cancelQuote",
      Consensus: dataMarket.Consensus,
      Time: ctx.BlockHeader().Time,
      Refund: quote.Backing,
      Balance: balance,
    }
    bz, _ := codec.MarshalJSONIndent(keeper.cdc, data)
    
    return sdk.Result {
        Data: bz,
        Tags: tags,
    }
}
