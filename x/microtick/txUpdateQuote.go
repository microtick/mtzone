package microtick

import (
    "fmt"
    
    "github.com/cosmos/cosmos-sdk/codec"
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

type UpdateQuoteData struct {
    Id MicrotickId `json:"id"`
    Originator string `json:"originator"`
    Spot MicrotickSpot `json:"spot"`
    Premium MicrotickPremium `json:"premium"`
    Consensus MicrotickSpot `json:"consensus"`
    Balance MicrotickCoin `json:"balance"`
    Commission MicrotickCoin `json:"commission"`
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
    return sdk.MustSortJSON(msgCdc.MustMarshalJSON(msg))
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
    dataMarket.DeleteQuote(quote)
    
    if msg.NewSpot.Amount.IsPositive() {
        quote.Spot = msg.NewSpot
        quote.Freeze(params)
    }
    
    if msg.NewPremium.Amount.IsPositive() {
        quote.Premium = msg.NewPremium
        quote.ComputeQuantity()
        quote.Freeze(params)
    }
    
    dataMarket.AddQuote(quote)
    dataMarket.factorIn(quote)
    keeper.SetDataMarket(ctx, dataMarket)
    keeper.SetActiveQuote(ctx, quote)
    
    // Tags
    tags := sdk.NewTags(
        fmt.Sprintf("acct.%s", msg.Requester.String()), "quote.update",
        fmt.Sprintf("quote.%d", quote.Id), "update",
        "mtm.MarketTick", quote.Market,
    )
    
    // Data
    data := CreateQuoteData {
      Id: quote.Id,
      Originator: "updateQuote",
      Spot: quote.Spot,
      Premium: quote.Premium,
      Consensus: dataMarket.Consensus,
      Balance: NewMicrotickCoinFromInt(0),
      Commission: NewMicrotickCoinFromInt(0),
    }
    bz, _ := codec.MarshalJSONIndent(keeper.cdc, data)
    
    return sdk.Result {
        Data: bz,
        Tags: tags,
    }
}
