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

func (msg TxWithdrawQuote) Route() string { return "microtick" }

func (msg TxWithdrawQuote) Type() string { return "quote_withdraw" }

func (msg TxWithdrawQuote) ValidateBasic() error {
    if msg.Requester.Empty() {
        return sdkerrors.Wrap(mt.ErrInvalidAddress, msg.Requester.String())
    }
    return nil
}

func (msg TxWithdrawQuote) GetSignBytes() []byte {
    return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg TxWithdrawQuote) GetSigners() []sdk.AccAddress {
    return []sdk.AccAddress{ msg.Requester }
}

// Handler

func HandleTxWithdrawQuote(ctx sdk.Context, keeper keeper.Keeper, params mt.MicrotickParams, 
    msg TxWithdrawQuote) (*sdk.Result, error) {
        
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
    
    // Withdraw amount must be strictly less than quote backing (to withdraw the full amount, use CancelQUote)
    if msg.Withdraw.IsGTE(quote.Backing) {
        return nil, mt.ErrQuoteBacking
    }
    
    commission := mt.NewMicrotickCoinFromDec(msg.Withdraw.Amount.Mul(params.CommissionQuotePercent))
    
    total := msg.Withdraw.Sub(commission)
    
    // Add coins from requester
    err = keeper.DepositMicrotickCoin(ctx, msg.Requester, total)
    if err != nil {
        return nil, mt.ErrInsufficientFunds
    }
    // Add commission to pool
    //fmt.Printf("Withdraw Commission: %s\n", commission.String())
    reward, err := keeper.PoolCommission(ctx, msg.Requester, commission, true)
    if err != nil {
        return nil, err
    }
    
    dataMarket, err := keeper.GetDataMarket(ctx, quote.Market)
    if err != nil {
        return nil, mt.ErrInvalidMarket
    }
    
    dataMarket.FactorOut(quote)
    
    quote.Backing = mt.NewMicrotickCoinFromDec(quote.Backing.Amount.Sub(msg.Withdraw.Amount))
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
    accountStatus.QuoteBacking = accountStatus.QuoteBacking.Sub(msg.Withdraw)
    keeper.SetAccountStatus(ctx, msg.Requester, accountStatus)
    
    // Data
    data := WithdrawQuoteData {
      Account: msg.Requester,
      Id: quote.Id,
      Market: dataMarket.Market,
      Consensus: dataMarket.Consensus,
      Time: now.Unix(),
      Backing: msg.Withdraw,
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
        sdk.NewAttribute(fmt.Sprintf("quote.%d", quote.Id), "event.withdraw"),
        sdk.NewAttribute(fmt.Sprintf("acct.%s", msg.Requester.String()), "quote.withdraw"),
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
