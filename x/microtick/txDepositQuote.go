package microtick

import (
    "fmt"
    "time"
    
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
)

type TxDepositQuote struct {
    Id MicrotickId
    Requester sdk.AccAddress
    Deposit MicrotickCoin
}

func NewTxDepositQuote(id MicrotickId, requester sdk.AccAddress, 
    deposit MicrotickCoin) TxDepositQuote {
    return TxDepositQuote {
        Id: id,
        Requester: requester,
        Deposit: deposit,
    }
}

type DepositQuoteData struct {
    Id MicrotickId `json:"id"`
    Originator string `json:"originator"`
    Consensus MicrotickSpot `json:"consensus"`
    Time time.Time `json:"time"`
    Backing MicrotickCoin `json:"backing"`
    QuoteBacking MicrotickCoin `json:"quoteBacking"`
    Balance MicrotickCoin `json:"balance"`
    Commission MicrotickCoin `json:"commission"`
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
    return sdk.MustSortJSON(msgCdc.MustMarshalJSON(msg))
}

func (msg TxDepositQuote) GetSigners() []sdk.AccAddress {
    return []sdk.AccAddress{ msg.Requester }
}

// Handler

func handleTxDepositQuote(ctx sdk.Context, keeper Keeper, msg TxDepositQuote) sdk.Result {
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
    
    commission := NewMicrotickCoinFromDec(msg.Deposit.Amount.Mul(params.CommissionQuotePercent))
    
    total := msg.Deposit.Add(commission)
    
    // Subtract coins from requester
    keeper.WithdrawMicrotickCoin(ctx, msg.Requester, total)
    // Add commission to pool
    fmt.Printf("Deposit Commission: %s\n", commission.String())
    keeper.PoolCommission(ctx, commission)
    
    dataMarket, _ := keeper.GetDataMarket(ctx, quote.Market)
    dataMarket.factorOut(quote)
    
    quote.Backing = NewMicrotickCoinFromDec(quote.Backing.Amount.Add(msg.Deposit.Amount))
    quote.ComputeQuantity()
    
    // But we do freeze the new backing from any other updates
    now := ctx.BlockHeader().Time
    quote.Freeze(now, params)
    
    dataMarket.factorIn(quote)
    keeper.SetDataMarket(ctx, dataMarket)
    keeper.SetActiveQuote(ctx, quote)
    
     // DataAccountStatus
    
    accountStatus := keeper.GetAccountStatus(ctx, msg.Requester)
    accountStatus.QuoteBacking = accountStatus.QuoteBacking.Add(msg.Deposit)
    keeper.SetAccountStatus(ctx, msg.Requester, accountStatus)
    
    balance := accountStatus.Change
    coins := keeper.coinKeeper.GetCoins(ctx, msg.Requester)
    for i := 0; i < len(coins); i++ {
        if coins[i].Denom == TokenType {
            balance = balance.Add(NewMicrotickCoinFromInt(coins[i].Amount.Int64()))
        }
    }
    
    tags := sdk.NewTags(
        fmt.Sprintf("quote.%d", quote.Id), "event.deposit",
        fmt.Sprintf("acct.%s", msg.Requester.String()), "quote.deposit",
        "mtm.MarketTick", quote.Market,
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
    bz, _ := codec.MarshalJSONIndent(keeper.cdc, data)
    
    return sdk.Result {
        Data: bz,
        Tags: tags,
    }
}
