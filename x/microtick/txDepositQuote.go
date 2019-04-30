package microtick

import (
    "fmt"
    "encoding/json"
    
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
    b, err := json.Marshal(msg)
    if err != nil {
        panic(err)
    }
    return sdk.MustSortJSON(b)
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
    //if quote.Frozen() {
        //return sdk.ErrInternal(fmt.Sprintf("Quote is frozen until: %s", quote.CanModify)).Result()
    //}
    
    // Subtract coins from requester
    keeper.WithdrawMicrotickCoin(ctx, msg.Requester, msg.Deposit)
    
    dataMarket, _ := keeper.GetDataMarket(ctx, quote.Market)
    dataMarket.factorOut(quote)
    
    quote.Backing = NewMicrotickCoinFromDec(quote.Backing.Amount.Add(msg.Deposit.Amount))
    quote.ComputeQuantity()
    
    // But we do freeze the new backing from any other updates
    quote.Freeze(params)
    
    dataMarket.factorIn(quote)
    keeper.SetDataMarket(ctx, dataMarket)
    keeper.SetActiveQuote(ctx, quote)
    
     // DataAccountStatus
    
    accountStatus := keeper.GetAccountStatus(ctx, msg.Requester)
    accountStatus.QuoteBacking = accountStatus.QuoteBacking.Plus(msg.Deposit)
    keeper.SetAccountStatus(ctx, msg.Requester, accountStatus)
    
    tags := sdk.NewTags(
        fmt.Sprintf("acct.%s", msg.Requester.String()), "quote.deposit",
        fmt.Sprintf("quote.%d", quote.Id), "deposit",
        "mtm.MarketTick", quote.Market,
    )
    
    // Data
    data := DepositQuoteData {
      Id: quote.Id,
      Originator: "depositQuote",
      Consensus: dataMarket.Consensus,
      Backing: msg.Deposit,
      QuoteBacking: quote.Backing,
      Balance: NewMicrotickCoinFromInt(0),
      Commission: NewMicrotickCoinFromInt(0),
    }
    bz, _ := codec.MarshalJSONIndent(keeper.cdc, data)
    
    return sdk.Result {
        Data: bz,
        Tags: tags,
    }
}
