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

type TxCreateQuote struct {
    Market mt.MicrotickMarket
    Duration mt.MicrotickDurationName
    Provider mt.MicrotickAccount
    Backing mt.MicrotickCoin
    Spot mt.MicrotickSpot
    Premium mt.MicrotickPremium
}

func NewTxCreateQuote(market mt.MicrotickMarket, dur mt.MicrotickDurationName, provider mt.MicrotickAccount, 
    backing mt.MicrotickCoin, spot mt.MicrotickSpot, premium mt.MicrotickPremium) TxCreateQuote {
    return TxCreateQuote {
        Market: market,
        Duration: dur,
        Provider: provider,
        Backing: backing,
        Spot: spot,
        Premium: premium,
    }
}

type CreateQuoteData struct {
    Account string `json:"account"`
    Id mt.MicrotickId `json:"id"`
    Market mt.MicrotickMarket `json:"market"`
    Duration mt.MicrotickDurationName `json:"duration"`
    Spot mt.MicrotickSpot `json:"spot"`
    Premium mt.MicrotickPremium `json:"premium"`
    Consensus mt.MicrotickSpot `json:"consensus"`
    Time time.Time `json:"time"`
    Backing mt.MicrotickCoin `json:"backing"`
    Commission mt.MicrotickCoin `json:"commission"`
}

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
    return nil
}

func (msg TxCreateQuote) GetSignBytes() []byte {
    return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg TxCreateQuote) GetSigners() []sdk.AccAddress {
    return []sdk.AccAddress{msg.Provider}
}

// Handler

func HandleTxCreateQuote(ctx sdk.Context, mtKeeper keeper.Keeper, params mt.Params,
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
        msg.Backing, msg.Spot, msg.Premium)
    dataActiveQuote.ComputeQuantity()
    dataActiveQuote.Freeze(now, params)
    mtKeeper.SetActiveQuote(ctx, dataActiveQuote)
    
    // DataAccountStatus
    
    accountStatus := mtKeeper.GetAccountStatus(ctx, msg.Provider)
    accountStatus.ActiveQuotes.Insert(keeper.NewListItem(id, sdk.NewDec(int64(id))))
    accountStatus.NumQuotes++
    accountStatus.QuoteBacking = accountStatus.QuoteBacking.Add(msg.Backing)
    
    // DataMarket
    
    dataMarket, err2 := mtKeeper.GetDataMarket(ctx, msg.Market)
    if err2 != nil {
        return nil, sdkerrors.Wrap(mt.ErrInvalidMarket, msg.Market)
    }
    dataMarket.AddQuote(dataActiveQuote)
    if !dataMarket.FactorIn(dataActiveQuote, true) {
        return nil, mt.ErrQuoteParams
    }
    
    mtKeeper.SetAccountStatus(ctx, msg.Provider, accountStatus)
    mtKeeper.SetDataMarket(ctx, dataMarket)
    
    // Subtract coins from quote provider
    //fmt.Printf("Total: %s\n", total.String())
    
    err2 = mtKeeper.WithdrawMicrotickCoin(ctx, msg.Provider, total)
    if err2!= nil {
        return nil, mt.ErrInsufficientFunds
    }
    
    //fmt.Printf("Create Commission: %s\n", commission.String())
    mtKeeper.PoolCommission(ctx, msg.Provider, commission)
    
    // Data
    data := CreateQuoteData {
      Account: msg.Provider.String(),
      Id: id,
      Market: msg.Market,
      Duration: msg.Duration,
      Spot: msg.Spot,
      Premium: msg.Premium,
      Consensus: dataMarket.Consensus,
      Time: now,
      Backing: msg.Backing,
      Commission: commission,
    }
    bz, _ := codec.MarshalJSONIndent(ModuleCdc, data)
    
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
    ))
    
    ctx.EventManager().EmitEvents(events)
    
	return &sdk.Result {
	    Data: bz,
	    Events: ctx.EventManager().ABCIEvents(),
	}, nil
}
