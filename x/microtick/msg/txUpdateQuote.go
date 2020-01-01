package msg

import (
    "fmt"
    "time"
    
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
    
    mt "github.com/mjackson001/mtzone/x/microtick/types"
    "github.com/mjackson001/mtzone/x/microtick/keeper"
)

type TxUpdateQuote struct {
    Id mt.MicrotickId
    Requester mt.MicrotickAccount
    NewSpot mt.MicrotickSpot
    NewPremium mt.MicrotickPremium
}

func NewTxUpdateQuote(id mt.MicrotickId, requester sdk.AccAddress, 
    newSpot mt.MicrotickSpot, newPremium mt.MicrotickPremium) TxUpdateQuote {
    return TxUpdateQuote {
        Id: id,
        Requester: requester,
        NewSpot: newSpot,
        NewPremium: newPremium,
    }
}

type UpdateQuoteData struct {
    Id mt.MicrotickId `json:"id"`
    Originator string `json:"originator"`
    Market mt.MicrotickMarket `json:"market"`
    Duration mt.MicrotickDurationName `json:"duration"`
    Spot mt.MicrotickSpot `json:"spot"`
    Premium mt.MicrotickPremium `json:"premium"`
    Consensus mt.MicrotickSpot `json:"consensus"`
    Time time.Time `json:"time"`
    Balance mt.MicrotickCoin `json:"balance"`
    Commission mt.MicrotickCoin `json:"commission"`
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
    return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg TxUpdateQuote) GetSigners() []sdk.AccAddress {
    return []sdk.AccAddress{ msg.Requester }
}

// Handler

func HandleTxUpdateQuote(ctx sdk.Context, keeper keeper.Keeper, msg TxUpdateQuote) sdk.Result {
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
    
    commission := mt.NewMicrotickCoinFromDec(quote.Backing.Amount.Mul(params.CommissionUpdatePercent))
    
    dataMarket, _ := keeper.GetDataMarket(ctx, quote.Market)
    dataMarket.FactorOut(quote)
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
    if !dataMarket.FactorIn(quote) {
        return sdk.ErrInternal("Quote params out of range").Result()
    }
    
    keeper.SetDataMarket(ctx, dataMarket)
    keeper.SetActiveQuote(ctx, quote)
    
    // Subtract coins from requester
    keeper.WithdrawMicrotickCoin(ctx, msg.Requester, commission)
    // Add commission to pool
    //fmt.Printf("Update Commission: %s\n", commission.String())
    keeper.PoolCommission(ctx, msg.Requester, commission)
    
    balance := keeper.GetTotalBalance(ctx, msg.Requester)
   
    // Events
    event := sdk.NewEvent(
        sdk.EventTypeMessage,
        sdk.NewAttribute(fmt.Sprintf("quote.%d", quote.Id), "event.update"),
        sdk.NewAttribute(fmt.Sprintf("acct.%s", msg.Requester.String()), "quote.update"),
        sdk.NewAttribute("mtm.MarketTick", quote.Market),
    )
    
    // Data
    data := UpdateQuoteData {
      Id: quote.Id,
      Originator: "updateQuote",
      Market: quote.Market,
      Duration: mt.MicrotickDurationNameFromDur(quote.Duration),
      Spot: quote.Spot,
      Premium: quote.Premium,
      Consensus: dataMarket.Consensus,
      Time: now,
      Balance: balance,
      Commission: commission,
    }
    bz, _ := codec.MarshalJSONIndent(ModuleCdc, data)
    
    return sdk.Result {
        Data: bz,
        Events: []sdk.Event{ event },
    }
}
