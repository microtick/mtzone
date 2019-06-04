package microtick

import (
    "fmt"
    "time"
    
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
)

type TxCancelQuote struct {
    Id MicrotickId
    Requester sdk.AccAddress
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
    
    if quote.Provider.String() != msg.Requester.String() {
        return sdk.ErrInternal("Account can't modify quote").Result()
    }
    
    if quote.Frozen() {
        return sdk.ErrInternal(fmt.Sprintf("Quote is frozen until: %s", quote.CanModify)).Result()
    }
    
    // Everything ok, let's refund the backing and delete the quote
    keeper.DepositMicrotickCoin(ctx, msg.Requester, quote.Backing)
    
    dataMarket, _ := keeper.GetDataMarket(ctx, quote.Market)
    dataMarket.factorOut(quote)
    dataMarket.DeleteQuote(quote)
    keeper.SetDataMarket(ctx, dataMarket)
    
    keeper.DeleteActiveQuote(ctx, quote.Id)
    
    accountStatus := keeper.GetAccountStatus(ctx, msg.Requester)
    accountStatus.QuoteBacking = accountStatus.QuoteBacking.Sub(quote.Backing)
    accountStatus.ActiveQuotes.Delete(quote.Id)
    keeper.SetAccountStatus(ctx, msg.Requester, accountStatus)
    
    tags := sdk.NewTags(
        fmt.Sprintf("quote.%d", quote.Id), "cancel",
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
      Balance: NewMicrotickCoinFromInt(0),
    }
    bz, _ := codec.MarshalJSONIndent(keeper.cdc, data)
    
    return sdk.Result {
        Data: bz,
        Tags: tags,
    }
}
