package msg

import (
    "time"
    
    "github.com/gogo/protobuf/proto"
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    
    mt "gitlab.com/microtick/mtzone/x/microtick/types"
    "gitlab.com/microtick/mtzone/x/microtick/keeper"
)

func (msg TxUpdateQuote) Route() string { return "microtick" }

func (msg TxUpdateQuote) Type() string { return "update" }

func (msg TxUpdateQuote) ValidateBasic() error {
    if msg.Requester.Empty() {
        return sdkerrors.Wrap(mt.ErrInvalidAddress, msg.Requester.String())
    }
    newBid := mt.NewMicrotickPremiumFromString(msg.NewBid)
    newAsk := mt.NewMicrotickPremiumFromString(msg.NewAsk)
    if newBid.Amount.GT(newAsk.Amount) {
        return sdkerrors.Wrap(mt.ErrInvalidQuote, "bid > ask")
    }
    return nil
}

func (msg TxUpdateQuote) GetSignBytes() []byte {
    return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg TxUpdateQuote) GetSigners() []sdk.AccAddress {
    return []sdk.AccAddress{ msg.Requester }
}

// Handler

func HandleTxUpdateQuote(ctx sdk.Context, mtKeeper keeper.Keeper, params mt.MicrotickParams, 
    msg TxUpdateQuote) (*sdk.Result, error) {
        
    newSpot := mt.NewMicrotickSpotFromString(msg.NewSpot)
    newBid := mt.NewMicrotickPremiumFromString(msg.NewBid)
    newAsk := mt.NewMicrotickPremiumFromString(msg.NewAsk)
    
    quote, err := mtKeeper.GetActiveQuote(ctx, msg.Id)
    if err != nil {
        return nil, sdkerrors.Wrapf(mt.ErrInvalidQuote, "%d", msg.Id)
    }
    
    if quote.Provider.String() != msg.Requester.String() {
        return nil, mt.ErrNotOwner
    }
    
    if quote.Frozen(ctx.BlockHeader().Time) {
        return nil, sdkerrors.Wrap(mt.ErrQuoteFrozen, time.Unix(quote.CanModify, 0).String())
    }
    
    dataMarket, err := mtKeeper.GetDataMarket(ctx, quote.Market)
    if err != nil {
        return nil, mt.ErrInvalidMarket
    }
    
    orderBook := dataMarket.GetOrderBook(quote.DurationName)
    adjustment := sdk.OneDec()
    if len(orderBook.CallAsks.Data) > 0 {
        bestCallAsk, _ := mtKeeper.GetActiveQuote(ctx, orderBook.CallAsks.Data[0].Id)
        bestPutAsk, _ := mtKeeper.GetActiveQuote(ctx, orderBook.PutAsks.Data[0].Id)
        average := bestCallAsk.CallAsk(dataMarket.Consensus).Amount.Add(bestPutAsk.PutAsk(dataMarket.Consensus).Amount).QuoInt64(2)
        if quote.Ask.Amount.GT(average) {
            adjustment = average.Quo(quote.Ask.Amount)
        }
    }    
    
    commission := mtKeeper.PoolCommission(ctx, quote.Backing.Amount.Mul(params.CommissionUpdatePerunit).Quo(adjustment))
    
    dataMarket.FactorOut(quote)
    dataMarket.DeleteQuote(quote)
    
    now := ctx.BlockHeader().Time
    
    if newSpot.Amount.IsPositive() {
        quote.Spot = newSpot
        quote.Freeze(now, params)
    }
    
    if newAsk.Amount.IsPositive() {
        quote.Ask = newAsk
        quote.Freeze(now, params)
    }
    
    if newBid.Amount.GTE(sdk.ZeroDec()) {
        quote.Bid = newBid
    }
    
    // Recompute quantity
    quote.ComputeQuantity()
    
    dataMarket.AddQuote(quote)
    if !dataMarket.FactorIn(quote, true) {
        return nil, mt.ErrQuoteParams
    }
    
    mtKeeper.SetDataMarket(ctx, dataMarket)
    mtKeeper.SetActiveQuote(ctx, quote)
    
    // Subtract coins from requester
    err = mtKeeper.WithdrawMicrotickCoin(ctx, msg.Requester, commission)
    if err != nil {
        return nil, mt.ErrInsufficientFunds
    }
    
    // Add commission to pool
    //fmt.Printf("Update Commission: %s\n", commission.String())
    reward, err := mtKeeper.AwardRebate(ctx, msg.Requester, quote.Backing.Amount.Mul(params.MintRewardUpdatePerunit).Mul(adjustment))
    if err != nil {
        return nil, err
    }
    
    // Data
    data := UpdateQuoteData {
      Time: now.Unix(),
      Market: quote.Market,
      Duration: quote.DurationName,
      Consensus: dataMarket.Consensus,
      Commission: commission,
      Reward: *reward,
      Adjustment: adjustment.String(),
    }
    bz, err := proto.Marshal(&data)
    
    var events []sdk.Event
    events = append(events, sdk.NewEvent(
        sdk.EventTypeMessage,
        sdk.NewAttribute(sdk.AttributeKeyModule, mt.ModuleKey),
    ))
    
    ctx.EventManager().EmitEvents(events)
    
    return &sdk.Result {
        Data: bz,
        Events: ctx.EventManager().ABCIEvents(),
    }, nil
}
