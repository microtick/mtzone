package msg

import (
    "fmt"
    "time"
    
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    
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
    Account string `json:"account"`
    Id mt.MicrotickId `json:"id"`
    Market mt.MicrotickMarket `json:"market"`
    Duration mt.MicrotickDurationName `json:"duration"`
    Spot mt.MicrotickSpot `json:"spot"`
    Premium mt.MicrotickPremium `json:"premium"`
    Consensus mt.MicrotickSpot `json:"consensus"`
    Time time.Time `json:"time"`
    Commission mt.MicrotickCoin `json:"commission"`
}

func (msg TxUpdateQuote) Route() string { return "microtick" }

func (msg TxUpdateQuote) Type() string { return "quote_update" }

func (msg TxUpdateQuote) ValidateBasic() error {
    if msg.Requester.Empty() {
        return sdkerrors.Wrap(mt.ErrInvalidAddress, msg.Requester.String())
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

func HandleTxUpdateQuote(ctx sdk.Context, keeper keeper.Keeper, params mt.Params, 
    msg TxUpdateQuote) (*sdk.Result, error) {
        
    quote, err := keeper.GetActiveQuote(ctx, msg.Id)
    if err != nil {
        return nil, sdkerrors.Wrapf(mt.ErrInvalidQuote, "%d", msg.Id)
    }
    
    if quote.Provider.String() != msg.Requester.String() {
        return nil, mt.ErrNotOwner
    }
    
    if quote.Frozen(ctx.BlockHeader().Time) {
        return nil, sdkerrors.Wrap(mt.ErrQuoteFrozen, quote.CanModify.String())
    }
    
    commission := mt.NewMicrotickCoinFromDec(quote.Backing.Amount.Mul(params.CommissionUpdatePercent))
    
    dataMarket, err2 := keeper.GetDataMarket(ctx, quote.Market)
    if err2 != nil {
        return nil, mt.ErrInvalidMarket
    }
    
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
    if !dataMarket.FactorIn(quote, true) {
        return nil, mt.ErrQuoteParams
    }
    
    keeper.SetDataMarket(ctx, dataMarket)
    keeper.SetActiveQuote(ctx, quote)
    
    // Subtract coins from requester
    err = keeper.WithdrawMicrotickCoin(ctx, msg.Requester, commission)
    if err != nil {
        return nil, mt.ErrInsufficientFunds
    }
    
    // Add commission to pool
    //fmt.Printf("Update Commission: %s\n", commission.String())
    err = keeper.PoolCommission(ctx, msg.Requester, commission)
    if err != nil {
        return nil, err
    }
    
    // Data
    data := UpdateQuoteData {
      Account: msg.Requester.String(),
      Id: quote.Id,
      Market: quote.Market,
      Duration: quote.DurationName,
      Spot: quote.Spot,
      Premium: quote.Premium,
      Consensus: dataMarket.Consensus,
      Time: now,
      Commission: commission,
    }
    bz, _ := codec.MarshalJSONIndent(ModuleCdc, data)
    
    var events []sdk.Event
    events = append(events, sdk.NewEvent(
        sdk.EventTypeMessage,
        sdk.NewAttribute(sdk.AttributeKeyModule, mt.ModuleKey),
    ), sdk.NewEvent(
        sdk.EventTypeMessage,
        sdk.NewAttribute(fmt.Sprintf("quote.%d", quote.Id), "event.update"),
        sdk.NewAttribute(fmt.Sprintf("acct.%s", msg.Requester.String()), "quote.update"),
        sdk.NewAttribute("mtm.MarketTick", quote.Market),
    ))
    
    ctx.EventManager().EmitEvents(events)
    
    return &sdk.Result {
        Data: bz,
        Events: ctx.EventManager().ABCIEvents(),
    }, nil
}
