package msg

import (
   	"github.com/gogo/protobuf/proto"
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    
    mt "gitlab.com/microtick/mtzone/x/microtick/types"
    "gitlab.com/microtick/mtzone/x/microtick/keeper"
)

func (msg TxCreateQuote) Route() string { return "microtick" }

func (msg TxCreateQuote) Type() string { return "create" }

func (msg TxCreateQuote) ValidateBasic() error {
    if msg.Market == "" {
        return sdkerrors.Wrap(mt.ErrInvalidMarket, msg.Market)
    }
    if msg.Provider.Empty() {
        return sdkerrors.Wrap(mt.ErrInvalidAddress, msg.Provider.String())
    }
    bid := mt.NewMicrotickPremiumFromString(msg.Bid)
    ask := mt.NewMicrotickPremiumFromString(msg.Ask)
    if bid.Amount.GT(ask.Amount) {
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
        
    backing := mt.NewMicrotickCoinFromString(msg.Backing)
    bid := mt.NewMicrotickPremiumFromString(msg.Bid)
    ask := mt.NewMicrotickPremiumFromString(msg.Ask)
    spot := mt.NewMicrotickSpotFromString(msg.Spot)
        
    if !mtKeeper.ValidDurationName(ctx, msg.Duration) {
        return nil, sdkerrors.Wrapf(mt.ErrInvalidDuration, "%s", msg.Duration)
    }
    
	// DataActiveQuote
	
    id := mtKeeper.GetNextActiveQuoteId(ctx)
     
    now := ctx.BlockHeader().Time
    dataActiveQuote := keeper.NewDataActiveQuote(now, id, msg.Market, 
        mtKeeper.DurationFromName(ctx, msg.Duration), msg.Duration, msg.Provider,
        backing, spot, ask, bid)
    dataActiveQuote.ComputeQuantity()
    dataActiveQuote.Freeze(now, params)
    mtKeeper.SetActiveQuote(ctx, dataActiveQuote)
    
    // DataAccountStatus
    
    accountStatus := mtKeeper.GetAccountStatus(ctx, msg.Provider)
    accountStatus.ActiveQuotes.Insert(keeper.NewListItem(id, sdk.NewDec(int64(id))))
    accountStatus.PlacedQuotes++
    accountStatus.QuoteBacking = accountStatus.QuoteBacking.Add(backing)
    
    // DataMarket
    mtKeeper.AssertDataMarketHasDuration(ctx, msg.Market, msg.Duration)
    dataMarket, err := mtKeeper.GetDataMarket(ctx, msg.Market)
    if err != nil {
        return nil, sdkerrors.Wrap(mt.ErrInvalidMarket, msg.Market)
    }
    
    orderBook := dataMarket.GetOrderBook(msg.Duration)
    
    adjustment := sdk.OneDec()
    if len(orderBook.CallAsks.Data) > 0 {
        bestCallAsk, _ := mtKeeper.GetActiveQuote(ctx, orderBook.CallAsks.Data[0].Id)
        bestPutAsk, _ := mtKeeper.GetActiveQuote(ctx, orderBook.PutAsks.Data[0].Id)
        average := bestCallAsk.CallAsk(dataMarket.Consensus).Amount.Add(bestPutAsk.PutAsk(dataMarket.Consensus).Amount).QuoInt64(2)
        if dataActiveQuote.Ask.Amount.GT(average) {
            adjustment = average.Quo(dataActiveQuote.Ask.Amount)
        }
    }
    
    commission := mtKeeper.PoolCommission(ctx, backing.Amount.Mul(params.CommissionCreatePerunit).Quo(adjustment))
    total := backing.Add(commission)
        
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
    reward, err := mtKeeper.AwardRebate(ctx, msg.Provider, backing.Amount.Mul(params.MintRewardCreatePerunit).Mul(adjustment))
    if err != nil {
        return nil, err
    }
    
    // Data
    data := CreateQuoteData {
      Time: now.Unix(),
      Id: id,
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
