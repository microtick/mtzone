package microtick

import (
    "fmt"
    "encoding/json"
    
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

func (msg TxCancelQuote) Route() string { return "microtick" }

func (msg TxCancelQuote) Type() string { return "cancel_quote" }

func (msg TxCancelQuote) ValidateBasic() sdk.Error {
    if msg.Requester.Empty() {
        return sdk.ErrInvalidAddress(msg.Requester.String())
    }
    return nil
}

func (msg TxCancelQuote) GetSignBytes() []byte {
    b, err := json.Marshal(msg)
    if err != nil {
        panic(err)
    }
    return sdk.MustSortJSON(b)
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
        return sdk.ErrInternal("Cannot modify quote").Result()
    }
    
    // Everything ok, let's refund the backing and delete the quote
    dataMarket, err2 := keeper.GetDataMarket(ctx, quote.Market)
    if err2 != nil {
        panic("Mismatched market / quote")
    }
    
    keeper.DepositDecCoin(ctx, msg.Requester, quote.Backing)
    
    dataMarket.factorOut(quote)
    dataMarket.DeleteQuote(quote)
    
    keeper.DeleteActiveQuote(ctx, quote.Id)
    
    accountStatus := keeper.GetAccountStatus(ctx, msg.Requester)
    accountStatus.QuoteBacking = accountStatus.QuoteBacking.Minus(quote.Backing)
    accountStatus.ActiveQuotes.Delete(quote.Id)
    keeper.SetAccountStatus(ctx, msg.Requester, accountStatus)
    
    return sdk.Result {}
}
