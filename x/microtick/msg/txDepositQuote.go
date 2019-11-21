package msg

import (
    "fmt"
    "time"
    
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
    Id mt.MicrotickId `json:"id"`
    Originator string `json:"originator"`
    Consensus mt.MicrotickSpot `json:"consensus"`
    Time time.Time `json:"time"`
    Backing mt.MicrotickCoin `json:"backing"`
    QuoteBacking mt.MicrotickCoin `json:"quoteBacking"`
    Balance mt.MicrotickCoin `json:"balance"`
    Commission mt.MicrotickCoin `json:"commission"`
}

func (msg TxDepositQuote) Route() string { return "microtick" }

func (msg TxDepositQuote) Type() string { return "deposit_quote" }

func (msg TxDepositQuote) ValidateBasic() sdk.Error {
    if msg.Requester.Empty() {
        return sdk.ErrInvalidAddress(msg.Requester.String())
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

func HandleTxDepositQuote(ctx sdk.Context, keeper keeper.Keeper, msg TxDepositQuote) sdk.Result {
    params := keeper.GetParams(ctx)
    
    quote, err := keeper.GetActiveQuote(ctx, msg.Id)
    if err != nil {
        return sdk.ErrInternal(fmt.Sprintf("No such quote: %d", msg.Id)).Result()
    }
    
    if quote.Provider.String() != msg.Requester.String() {
        return sdk.ErrInternal("Account can't modify quote").Result()
    }
    
    // No freeze for deposits
    //if quote.Frozen(ctx.BlockHeader().Time) {
        //return sdk.ErrInternal(fmt.Sprintf("Quote is frozen until: %s", quote.CanModify)).Result()
    //}
    
    commission := mt.NewMicrotickCoinFromDec(msg.Deposit.Amount.Mul(params.CommissionQuotePercent))
    
    total := msg.Deposit.Add(commission)
    
    // Subtract coins from requester
    keeper.WithdrawMicrotickCoin(ctx, msg.Requester, total)
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
    
    dataMarket.FactorIn(quote)
    keeper.SetDataMarket(ctx, dataMarket)
    keeper.SetActiveQuote(ctx, quote)
    
     // DataAccountStatus
    
    accountStatus := keeper.GetAccountStatus(ctx, msg.Requester)
    accountStatus.QuoteBacking = accountStatus.QuoteBacking.Add(msg.Deposit)
    keeper.SetAccountStatus(ctx, msg.Requester, accountStatus)
    
    balance := accountStatus.Change
    coins := keeper.CoinKeeper.GetCoins(ctx, msg.Requester)
    for i := 0; i < len(coins); i++ {
        if coins[i].Denom == mt.TokenType {
            balance = balance.Add(mt.NewMicrotickCoinFromInt(coins[i].Amount.Int64()))
        }
    }
    
    ctx.EventManager().EmitEvent(
        sdk.NewEvent(
            sdk.EventTypeMessage,
            sdk.NewAttribute(fmt.Sprintf("quote.%d", quote.Id), "event.deposit"),
            sdk.NewAttribute(fmt.Sprintf("acct.%s", msg.Requester.String()), "quote.deposit"),
            sdk.NewAttribute("mtm.MarketTick", quote.Market),
        ),
    )
    
    // Data
    data := DepositQuoteData {
      Id: quote.Id,
      Originator: "depositQuote",
      Consensus: dataMarket.Consensus,
      Time: now,
      Backing: msg.Deposit,
      QuoteBacking: quote.Backing,
      Balance: balance,
      Commission: commission,
    }
    bz, _ := codec.MarshalJSONIndent(ModuleCdc, data)
    
    return sdk.Result {
        Data: bz,
        Events: ctx.EventManager().Events(),
    }
}
