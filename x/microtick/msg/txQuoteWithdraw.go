package msg

import (
    "fmt"
    "time"
    
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
    sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
    
    mt "gitlab.com/microtick/mtzone/x/microtick/types"
    "gitlab.com/microtick/mtzone/x/microtick/keeper"
)

type TxWithdrawQuote struct {
    Id mt.MicrotickId
    Requester mt.MicrotickAccount
    Withdraw mt.MicrotickCoin
}

func NewTxWithdrawQuote(id mt.MicrotickId, requester sdk.AccAddress, 
    withdraw mt.MicrotickCoin) TxWithdrawQuote {
    return TxWithdrawQuote {
        Id: id,
        Requester: requester,
        Withdraw: withdraw,
    }
}

type WithdrawQuoteData struct {
    Account string `json:"account"`
    Id mt.MicrotickId `json:"id"`
    Market mt.MicrotickMarket `json:"market"`
    Consensus mt.MicrotickSpot `json:"consensus"`
    Time time.Time `json:"time"`
    Backing mt.MicrotickCoin `json:"backing"`
    QuoteBacking mt.MicrotickCoin `json:"quoteBacking"`
    Commission mt.MicrotickCoin `json:"commission"`
}

func (msg TxWithdrawQuote) Route() string { return "microtick" }

func (msg TxWithdrawQuote) Type() string { return "quote_withdraw" }

func (msg TxWithdrawQuote) ValidateBasic() error {
    if msg.Requester.Empty() {
        return sdkerrors.Wrap(mt.ErrInvalidAddress, msg.Requester.String())
    }
    return nil
}

func (msg TxWithdrawQuote) GetSignBytes() []byte {
    return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg TxWithdrawQuote) GetSigners() []sdk.AccAddress {
    return []sdk.AccAddress{ msg.Requester }
}

// Handler

func HandleTxWithdrawQuote(ctx sdk.Context, keeper keeper.Keeper, params mt.Params, 
    msg TxWithdrawQuote) (*sdk.Result, error) {
        
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
    reward, err := keeper.PoolCommission(ctx, msg.Requester, commission)
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
      Account: msg.Requester.String(),
      Id: quote.Id,
      Market: dataMarket.Market,
      Consensus: dataMarket.Consensus,
      Time: now,
      Backing: msg.Withdraw,
      QuoteBacking: quote.Backing,
      Commission: commission,
    }
    bz, _ := codec.MarshalJSONIndent(ModuleCdc, data)
    
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
