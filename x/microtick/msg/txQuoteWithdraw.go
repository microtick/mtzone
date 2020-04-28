package msg

import (
    "fmt"
    "time"
    "errors"
    
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
    
    mt "github.com/mjackson001/mtzone/x/microtick/types"
    "github.com/mjackson001/mtzone/x/microtick/keeper"
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
        return errors.New(fmt.Sprintf("Invalid address: %s", msg.Requester.String()))
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

func HandleTxWithdrawQuote(ctx sdk.Context, keeper keeper.Keeper, msg TxWithdrawQuote) (*sdk.Result, error) {
    params := keeper.GetParams(ctx)
    
    quote, err := keeper.GetActiveQuote(ctx, msg.Id)
    if err != nil {
        return nil, errors.New(fmt.Sprintf("No such quote: %d", msg.Id))
    }
    
    if quote.Provider.String() != msg.Requester.String() {
        return nil, errors.New("Account can't modify quote")
    }
    
    if quote.Frozen(ctx.BlockHeader().Time) {
        return nil, errors.New(fmt.Sprintf("Quote is frozen until: %s", quote.CanModify))
    }
    
    // Withdraw amount must be strictly less than quote backing (to withdraw the full amount, use CancelQUote)
    if msg.Withdraw.IsGTE(quote.Backing) {
        return nil, errors.New("Not enough backing in quote")
    }
    
    commission := mt.NewMicrotickCoinFromDec(msg.Withdraw.Amount.Mul(params.CommissionQuotePercent))
    
    total := msg.Withdraw.Sub(commission)
    
    // Add coins from requester
    err = keeper.DepositMicrotickCoin(ctx, msg.Requester, total)
    if err != nil {
        return nil, errors.New("Fund mismatch")
    }
    // Add commission to pool
    fmt.Printf("Withdraw Commission: %s\n", commission.String())
    keeper.PoolCommission(ctx, msg.Requester, commission)
    
    dataMarket, _ := keeper.GetDataMarket(ctx, quote.Market)
    dataMarket.FactorOut(quote)
    
    quote.Backing = mt.NewMicrotickCoinFromDec(quote.Backing.Amount.Sub(msg.Withdraw.Amount))
    quote.ComputeQuantity()
    
    // But we do freeze the new backing from any other updates
    now := ctx.BlockHeader().Time
    quote.Freeze(now, params)
    
    if !dataMarket.FactorIn(quote, true) {
        return nil, errors.New("Quote params out of range")
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
    ))
    
    ctx.EventManager().EmitEvents(events)
    
    return &sdk.Result {
        Data: bz,
        Events: ctx.EventManager().ABCIEvents(),
    }, nil
}
