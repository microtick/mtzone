package microtick

import (
    "fmt"
    "time"
    
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
)

type TxUpdateQuote struct {
    Id MicrotickId
    Requester MicrotickAccount
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
    Market MicrotickMarket `json:"market"`
    Duration MicrotickDurationName `json:"duration"`
    Spot MicrotickSpot `json:"spot"`
    Premium MicrotickPremium `json:"premium"`
    Consensus MicrotickSpot `json:"consensus"`
    Time time.Time `json:"time"`
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
    
    if quote.Frozen(ctx.BlockHeader().Time) {
        return sdk.ErrInternal(fmt.Sprintf("Quote is frozen until: %s", quote.CanModify)).Result()
    }
    
    commission := NewMicrotickCoinFromDec(quote.Backing.Amount.Mul(params.CommissionUpdatePercent))
    
    // Subtract coins from requester
    keeper.WithdrawMicrotickCoin(ctx, msg.Requester, commission)
    // Add commission to pool
    //fmt.Printf("Update Commission: %s\n", commission.String())
    keeper.PoolCommission(ctx, commission)
    
    dataMarket, _ := keeper.GetDataMarket(ctx, quote.Market)
    dataMarket.factorOut(quote)
    dataMarket.DeleteQuote(quote)
    
    now := ctx.BlockHeader().Time
    
    if msg.NewSpot.Amount.IsPositive() {
        quote.Spot = msg.NewSpot
        quote.Freeze(now, params)
    }
    
    if msg.NewPremium.Amount.IsPositive() {
        quote.Premium = msg.NewPremium
        quote.ComputeQuantity()
        quote.Freeze(now, params)
    }
    
    dataMarket.AddQuote(quote)
    dataMarket.factorIn(quote)
    keeper.SetDataMarket(ctx, dataMarket)
    keeper.SetActiveQuote(ctx, quote)
    
    balance := keeper.GetTotalBalance(ctx, msg.Requester)
   
    // Tags
    tags := sdk.NewTags(
        fmt.Sprintf("quote.%d", quote.Id), "event.update",
        fmt.Sprintf("acct.%s", msg.Requester.String()), "quote.update",
        "mtm.MarketTick", quote.Market,
    )
    
    // Data
    data := UpdateQuoteData {
      Id: quote.Id,
      Originator: "updateQuote",
      Market: quote.Market,
      Duration: MicrotickDurationNameFromDur(quote.Duration),
      Spot: quote.Spot,
      Premium: quote.Premium,
      Consensus: dataMarket.Consensus,
      Time: now,
      Balance: balance,
      Commission: commission,
    }
    bz, _ := codec.MarshalJSONIndent(keeper.cdc, data)
    
    return sdk.Result {
        Data: bz,
        Tags: tags,
    }
}
