package microtick

import (
    "fmt"
    "encoding/json"
    
    sdk "github.com/cosmos/cosmos-sdk/types"
)

type TxUpdateQuote struct {
    Id MicrotickId
    Requester sdk.AccAddress
    NewSpot MicrotickSpot
    NewPremium MicrotickPremium
}

func NewTxUpdateQuote(id MicrotickId, requester sdk.AccAddress, 
    newSpot MicrotickSpot, newPremium MicrotickPremium) TxUpdateQuote {
    return TxUpdateQuote {
        Id: id,
        Requester: requester,
        NewSpot: newSpot,
        NewPremium: newPremium,
    }
}

func (msg TxUpdateQuote) Route() string { return "microtick" }

func (msg TxUpdateQuote) Type() string { return "update_quote" }

func (msg TxUpdateQuote) ValidateBasic() sdk.Error {
    if msg.Requester.Empty() {
        return sdk.ErrInvalidAddress(msg.Requester.String())
    }
    return nil
}

func (msg TxUpdateQuote) GetSignBytes() []byte {
    b, err := json.Marshal(msg)
    if err != nil {
        panic(err)
    }
    return sdk.MustSortJSON(b)
}

func (msg TxUpdateQuote) GetSigners() []sdk.AccAddress {
    return []sdk.AccAddress{ msg.Requester }
}

// Handler

func handleTxUpdateQuote(ctx sdk.Context, keeper Keeper, msg TxUpdateQuote) sdk.Result {
    params := keeper.GetParams(ctx)
    
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
    
    dataMarket, _ := keeper.GetDataMarket(ctx, quote.Market)
    dataMarket.factorOut(quote)
    
    if msg.NewSpot.Amount.IsPositive() {
        quote.Spot = msg.NewSpot
        quote.Freeze(params)
    }
    
    if msg.NewPremium.Amount.IsPositive() {
        quote.Premium = msg.NewPremium
        quote.ComputeQuantity()
        quote.Freeze(params)
    }
    
    dataMarket.factorIn(quote)
    keeper.SetDataMarket(ctx, dataMarket)
    keeper.SetActiveQuote(ctx, quote)
    
    tags := sdk.NewTags(
        fmt.Sprintf("acct.%s", msg.Requester.String()), "quote.update",
        fmt.Sprintf("quote.%d", quote.Id), "update",
        "mtm.MarketTick", quote.Market,
    )
    
    return sdk.Result {
        Tags: tags,
    }
}
