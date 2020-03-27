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

type TxDepositQuote struct {
    Id mt.MicrotickId
    Requester mt.MicrotickAccount
    Deposit mt.MicrotickCoin
}

func NewTxDepositQuote(id mt.MicrotickId, requester sdk.AccAddress, 
    deposit mt.MicrotickCoin) TxDepositQuote {
    return TxDepositQuote {
        Id: id,
        Requester: requester,
        Deposit: deposit,
    }
}

type DepositQuoteData struct {
    Account string `json:"account"`
    Id mt.MicrotickId `json:"id"`
    Market mt.MicrotickMarket `json:"market"`
    Consensus mt.MicrotickSpot `json:"consensus"`
    Time time.Time `json:"time"`
    Backing mt.MicrotickCoin `json:"backing"`
    QuoteBacking mt.MicrotickCoin `json:"quoteBacking"`
    Commission mt.MicrotickCoin `json:"commission"`
}

func (msg TxDepositQuote) Route() string { return "microtick" }

func (msg TxDepositQuote) Type() string { return "quote_deposit" }

func (msg TxDepositQuote) ValidateBasic() error {
    if msg.Requester.Empty() {
        return errors.New(fmt.Sprintf("Invalid address: %s", msg.Requester.String()))
    }
    return nil
}

func (msg TxDepositQuote) GetSignBytes() []byte {
    return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg TxDepositQuote) GetSigners() []sdk.AccAddress {
    return []sdk.AccAddress{ msg.Requester }
}

// Handler

func HandleTxDepositQuote(ctx sdk.Context, keeper keeper.Keeper, msg TxDepositQuote) (*sdk.Result, error) {
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
    
    commission := mt.NewMicrotickCoinFromDec(msg.Deposit.Amount.Mul(params.CommissionQuotePercent))
    
    total := msg.Deposit.Add(commission)
    
    // Subtract coins from requester
    err = keeper.WithdrawMicrotickCoin(ctx, msg.Requester, total)
    if err != nil {
        return nil, errors.New("Insufficient funds")
    }
    
    // Add commission to pool
    //fmt.Printf("Deposit Commission: %s\n", commission.String())
    keeper.PoolCommission(ctx, msg.Requester, commission)
    
    dataMarket, _ := keeper.GetDataMarket(ctx, quote.Market)
    dataMarket.FactorOut(quote)
    
    quote.Backing = mt.NewMicrotickCoinFromDec(quote.Backing.Amount.Add(msg.Deposit.Amount))
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
    accountStatus.QuoteBacking = accountStatus.QuoteBacking.Add(msg.Deposit)
    keeper.SetAccountStatus(ctx, msg.Requester, accountStatus)
    
    // Data
    data := DepositQuoteData {
      Account: msg.Requester.String(),
      Id: quote.Id,
      Market: dataMarket.Market,
      Consensus: dataMarket.Consensus,
      Time: now,
      Backing: msg.Deposit,
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
        sdk.NewAttribute(fmt.Sprintf("quote.%d", quote.Id), "event.deposit"),
        sdk.NewAttribute(fmt.Sprintf("acct.%s", msg.Requester.String()), "quote.deposit"),
        sdk.NewAttribute("mtm.MarketTick", quote.Market),
    ))
    
    return &sdk.Result {
        Data: bz,
        Events: events,
    }, nil
}
