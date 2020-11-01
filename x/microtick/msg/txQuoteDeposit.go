package msg

import (
    "fmt"
    "time"
    
    "github.com/gogo/protobuf/proto"
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    
    mt "gitlab.com/microtick/mtzone/x/microtick/types"
    "gitlab.com/microtick/mtzone/x/microtick/keeper"
)

func (msg TxDepositQuote) Route() string { return "microtick" }

func (msg TxDepositQuote) Type() string { return "quote_deposit" }

func (msg TxDepositQuote) ValidateBasic() error {
    if msg.Requester.Empty() {
        return sdkerrors.Wrap(mt.ErrInvalidAddress, msg.Requester.String())
    }
    return nil
}

func (msg TxDepositQuote) GetSignBytes() []byte {
    return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg TxDepositQuote) GetSigners() []sdk.AccAddress {
    return []sdk.AccAddress{ msg.Requester }
}

// Handler

func HandleTxDepositQuote(ctx sdk.Context, keeper keeper.Keeper, params mt.MicrotickParams, 
    msg TxDepositQuote) (*sdk.Result, error) {
        
    quote, err := keeper.GetActiveQuote(ctx, msg.Id)
    if err != nil {
        return nil, sdkerrors.Wrapf(mt.ErrInvalidQuote, "%d", msg.Id)
    }
    
    if quote.Provider.String() != msg.Requester.String() {
        return nil, mt.ErrNotOwner
    }
    
    if quote.Frozen(ctx.BlockHeader().Time) {
        return nil, sdkerrors.Wrap(mt.ErrQuoteFrozen, time.Unix(quote.CanModify, 0).String())
    }
    
    commission := mt.NewMicrotickCoinFromDec(msg.Deposit.Amount.Mul(params.CommissionQuotePercent))
    
    total := msg.Deposit.Add(commission)
    
    // Subtract coins from requester
    err = keeper.WithdrawMicrotickCoin(ctx, msg.Requester, total)
    if err != nil {
        return nil, mt.ErrInsufficientFunds
    }
    
    dataMarket, err := keeper.GetDataMarket(ctx, quote.Market)
    if err != nil {
        return nil, mt.ErrInvalidMarket
    }
    
    dataMarket.FactorOut(quote)
    
    orderBook := dataMarket.GetOrderBook(quote.DurationName)
    adjustment := sdk.OneDec()
    if len(orderBook.CallAsks.Data) > 0 {
        bestCallAsk, _ := keeper.GetActiveQuote(ctx, orderBook.CallAsks.Data[0].Id)
        bestPutAsk, _ := keeper.GetActiveQuote(ctx, orderBook.PutAsks.Data[0].Id)
        average := bestCallAsk.CallAsk(dataMarket.Consensus).Amount.Add(bestPutAsk.PutAsk(dataMarket.Consensus).Amount).QuoInt64(2)
        if quote.Ask.Amount.GT(average) {
            adjustment = average.Quo(quote.Ask.Amount)
        }
    }
    
    // Add commission to pool
    //fmt.Printf("Deposit Commission: %s\n", commission.String())
    reward, err := keeper.PoolCommission(ctx, msg.Requester, commission, true, adjustment)
    if err != nil {
        return nil, err
    }
    
    quote.Backing = mt.NewMicrotickCoinFromDec(quote.Backing.Amount.Add(msg.Deposit.Amount))
    quote.ComputeQuantity()
    
    // But we do freeze the new backing from any other updates
    now := ctx.BlockHeader().Time
    quote.Freeze(now, params)
    
    if !dataMarket.FactorIn(quote, true) {
        return nil, mt.ErrQuoteParams
    }
    keeper.SetDataMarket(ctx, dataMarket)
    keeper.SetActiveQuote(ctx, quote)
    
     // DataAccountStatus
    
    accountStatus := keeper.GetAccountStatus(ctx, msg.Requester)
    accountStatus.QuoteBacking = accountStatus.QuoteBacking.Add(msg.Deposit)
    keeper.SetAccountStatus(ctx, msg.Requester, accountStatus)
    
    // Data
    data := DepositQuoteData {
      Account: msg.Requester,
      Id: quote.Id,
      Market: dataMarket.Market,
      Consensus: dataMarket.Consensus,
      Time: now.Unix(),
      Backing: msg.Deposit,
      QuoteBacking: quote.Backing,
      Commission: commission,
    }
    bz, err := proto.Marshal(&data)
    
    var events []sdk.Event
    events = append(events, sdk.NewEvent(
        sdk.EventTypeMessage,
        sdk.NewAttribute(sdk.AttributeKeyModule, mt.ModuleKey),
    ), sdk.NewEvent(
        sdk.EventTypeMessage,
        sdk.NewAttribute(fmt.Sprintf("quote.%d", quote.Id), "event.deposit"),
        sdk.NewAttribute(fmt.Sprintf("acct.%s", msg.Requester.String()), "quote.deposit"),
        sdk.NewAttribute("mtm.MarketTick", quote.Market),
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
