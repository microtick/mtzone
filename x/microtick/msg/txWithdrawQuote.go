package msg

import (
    "fmt"
    "time"
    
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
    Id mt.MicrotickId `json:"id"`
    Originator string `json:"originator"`
    Consensus mt.MicrotickSpot `json:"consensus"`
    Time time.Time `json:"time"`
    Backing mt.MicrotickCoin `json:"backing"`
    QuoteBacking mt.MicrotickCoin `json:"quoteBacking"`
    Balance mt.MicrotickCoin `json:"balance"`
    Commission mt.MicrotickCoin `json:"commission"`
}

func (msg TxWithdrawQuote) Route() string { return "microtick" }

func (msg TxWithdrawQuote) Type() string { return "withdraw_quote" }

func (msg TxWithdrawQuote) ValidateBasic() sdk.Error {
    if msg.Requester.Empty() {
        return sdk.ErrInvalidAddress(msg.Requester.String())
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

func HandleTxWithdrawQuote(ctx sdk.Context, keeper keeper.Keeper, msg TxWithdrawQuote) sdk.Result {
    params := keeper.GetParams(ctx)
    
    quote, err := keeper.GetActiveQuote(ctx, msg.Id)
    if err != nil {
        return sdk.ErrInternal(fmt.Sprintf("No such quote: %d", msg.Id)).Result()
    }
    
    if quote.Provider.String() != msg.Requester.String() {
        return sdk.ErrInternal("Account can't modify quote").Result()
    }
    
    // No freeze for withdraws
    if quote.Frozen(ctx.BlockHeader().Time) {
        return sdk.ErrInternal(fmt.Sprintf("Quote is frozen until: %s", quote.CanModify)).Result()
    }
    
    // Withdraw amount must be strictly less than quote backing (to withdraw the full amount, use CancelQUote)
    if msg.Withdraw.IsGTE(quote.Backing) {
        return sdk.ErrInternal("Not enough backing in quote").Result()
    }
    
    commission := mt.NewMicrotickCoinFromDec(msg.Withdraw.Amount.Mul(params.CommissionQuotePercent))
    
    total := msg.Withdraw.Sub(commission)
    
    // Add coins from requester
    keeper.DepositMicrotickCoin(ctx, msg.Requester, total)
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
    
    dataMarket.FactorIn(quote)
    keeper.SetDataMarket(ctx, dataMarket)
    keeper.SetActiveQuote(ctx, quote)
    
     // DataAccountStatus
    
    accountStatus := keeper.GetAccountStatus(ctx, msg.Requester)
    accountStatus.QuoteBacking = accountStatus.QuoteBacking.Sub(msg.Withdraw)
    keeper.SetAccountStatus(ctx, msg.Requester, accountStatus)
    
    coins := keeper.CoinKeeper.GetCoins(ctx, msg.Requester)
    balance := mt.NewMicrotickCoinFromExtCoinInt(coins.AmountOf(mt.ExtTokenType).Int64())
    
    event := sdk.NewEvent(
        sdk.EventTypeMessage,
        sdk.NewAttribute(fmt.Sprintf("quote.%d", quote.Id), "event.withdraw"),
        sdk.NewAttribute(fmt.Sprintf("acct.%s", msg.Requester.String()), "quote.withdraw"),
        sdk.NewAttribute("mtm.MarketTick", quote.Market),
    )
    
    // Data
    data := WithdrawQuoteData {
      Id: quote.Id,
      Originator: "withdrawQuote",
      Consensus: dataMarket.Consensus,
      Time: now,
      Backing: msg.Withdraw,
      QuoteBacking: quote.Backing,
      Balance: balance,
      Commission: commission,
    }
    bz, _ := codec.MarshalJSONIndent(ModuleCdc, data)
    
    return sdk.Result {
        Data: bz,
        Events: []sdk.Event{ event },
    }
}
