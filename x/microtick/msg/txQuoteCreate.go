package msg

import (
    "fmt"
    
   	"github.com/gogo/protobuf/proto"
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    
    mt "gitlab.com/microtick/mtzone/x/microtick/types"
    "gitlab.com/microtick/mtzone/x/microtick/keeper"
)

func (msg TxCreateQuote) Route() string { return "microtick" }

func (msg TxCreateQuote) Type() string { return "quote_create" }

func (msg TxCreateQuote) ValidateBasic() error {
    if msg.Market == "" {
        return sdkerrors.Wrap(mt.ErrInvalidMarket, msg.Market)
    }
    if msg.Provider.Empty() {
        return sdkerrors.Wrap(mt.ErrInvalidAddress, msg.Provider.String())
    }
    if !msg.Backing.IsPositive() {
        return mt.ErrQuoteBacking
    }
    if msg.Bid.Amount.GT(msg.Ask.Amount) {
        return sdkerrors.Wrap(mt.ErrInvalidQuote, "bid > ask")
    }
    return nil
}

func (msg TxCreateQuote) GetSignBytes() []byte {
    return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg TxCreateQuote) GetSigners() []sdk.AccAddress {
    return []sdk.AccAddress{msg.Provider}
}

// Handler

func HandleTxCreateQuote(ctx sdk.Context, mtKeeper keeper.Keeper, params mt.MicrotickParams,
    msg TxCreateQuote) (*sdk.Result, error) {
        
    // Do not create since markets are now a governance question
    //if !mtKeeper.HasDataMarket(ctx, msg.Market) {
        //mtKeeper.SetDataMarket(ctx, keeper.NewDataMarket(msg.Market))
    //}
    
    if !mtKeeper.ValidDurationName(ctx, msg.Duration) {
        return nil, sdkerrors.Wrapf(mt.ErrInvalidDuration, "%s", msg.Duration)
    }
    
    commission := mt.NewMicrotickCoinFromDec(msg.Backing.Amount.Mul(params.CommissionQuotePercent))
    total := msg.Backing.Add(commission)
        
	// DataActiveQuote
	
    id := mtKeeper.GetNextActiveQuoteId(ctx)
     
    now := ctx.BlockHeader().Time
    dataActiveQuote := keeper.NewDataActiveQuote(now, id, msg.Market, 
        mtKeeper.DurationFromName(ctx, msg.Duration), msg.Duration, msg.Provider,
        msg.Backing, msg.Spot, msg.Ask, msg.Bid)
    dataActiveQuote.ComputeQuantity()
    dataActiveQuote.Freeze(now, params)
    mtKeeper.SetActiveQuote(ctx, dataActiveQuote)
    
    // DataAccountStatus
    
    accountStatus := mtKeeper.GetAccountStatus(ctx, msg.Provider)
    accountStatus.ActiveQuotes.Insert(keeper.NewListItem(id, sdk.NewDec(int64(id))))
    accountStatus.PlacedQuotes++
    accountStatus.QuoteBacking = accountStatus.QuoteBacking.Add(msg.Backing)
    
    // DataMarket
    
    dataMarket, err := mtKeeper.GetDataMarket(ctx, msg.Market)
    if err != nil {
        return nil, sdkerrors.Wrap(mt.ErrInvalidMarket, msg.Market)
    }
    dataMarket.AddQuote(dataActiveQuote)
    if !dataMarket.FactorIn(dataActiveQuote, true) {
        return nil, mt.ErrQuoteParams
    }
    
    mtKeeper.CommitQuoteId(ctx, id)
    mtKeeper.SetAccountStatus(ctx, msg.Provider, accountStatus)
    mtKeeper.SetDataMarket(ctx, dataMarket)
    
    // Subtract coins from quote provider
    //fmt.Printf("Total: %s\n", total.String())
    
    err = mtKeeper.WithdrawMicrotickCoin(ctx, msg.Provider, total)
    if err != nil {
        return nil, mt.ErrInsufficientFunds
    }
    
    //fmt.Printf("Create Commission: %s\n", commission.String())
    reward, err := mtKeeper.PoolCommission(ctx, msg.Provider, commission, true)
    if err != nil {
        return nil, err
    }
    
    // Data
    data := CreateQuoteData {
      Account: msg.Provider,
      Id: id,
      Market: msg.Market,
      Duration: msg.Duration,
      Spot: msg.Spot,
      Ask: msg.Ask,
      Bid: msg.Bid,
      Consensus: dataMarket.Consensus,
      Time: now.Unix(),
      Backing: msg.Backing,
      Commission: commission,
    }
    bz, err := proto.Marshal(&data)
    
    var events []sdk.Event
    events = append(events, sdk.NewEvent(
        sdk.EventTypeMessage,
        sdk.NewAttribute(sdk.AttributeKeyModule, mt.ModuleKey),
    ), sdk.NewEvent(
        sdk.EventTypeMessage,
        sdk.NewAttribute("mtm.NewQuote", fmt.Sprintf("%d", id)),
        sdk.NewAttribute(fmt.Sprintf("quote.%d", id), "event.create"),
        sdk.NewAttribute(fmt.Sprintf("acct.%s", msg.Provider.String()), "quote.create"),
        sdk.NewAttribute("mtm.MarketTick", msg.Market),
    ), sdk.NewEvent(
        sdk.EventTypeMessage,
        sdk.NewAttribute("commission", commission.String()),
        sdk.NewAttribute("reward", reward.String()),
    ))
    
    ctx.EventManager().EmitEvents(events)
    
	return &sdk.Result {
	    Data: bz,
	    Events: ctx.EventManager().ABCIEvents(),
	}, nil
}
