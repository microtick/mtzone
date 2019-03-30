package microtick

import (
    "fmt"
    "encoding/json"
    
    sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/cosmos/cosmos-sdk/codec"
)

// Tx

type TxCreateQuote struct {
    Account sdk.AccAddress
    Backing sdk.Coins
}

func NewTxCreateQuote(account sdk.AccAddress, backing sdk.Coins) TxCreateQuote {
    return TxCreateQuote {
        Account: account,
        Backing: backing,
    }
}

func (msg TxCreateQuote) Route() string { return "microtick" }

func (msg TxCreateQuote) Type() string { return "create_quote" }

func (msg TxCreateQuote) ValidateBasic() sdk.Error {
    if msg.Account.Empty() {
        return sdk.ErrInvalidAddress(msg.Account.String())
    }
    //if !msg.Backing.IsAllPositive() {
        //return sdk.ErrInsufficientCoins("Backing must be positive")
    //}
    return nil
}

func (msg TxCreateQuote) GetSignBytes() []byte {
    b, err := json.Marshal(msg)
    if err != nil {
        panic(err)
    }
    return sdk.MustSortJSON(b)
}

func (msg TxCreateQuote) GetSigners() []sdk.AccAddress {
    return []sdk.AccAddress{msg.Account}
}

// Codec

func RegisterCodec(cdc *codec.Codec) {
    cdc.RegisterConcrete(TxCreateQuote{}, "microtick/CreateQuote", nil)
}

// Handler

func handleTxCreateQuote(ctx sdk.Context, keeper Keeper, msg TxCreateQuote) sdk.Result {
    id := keeper.GetNextActiveQuoteId(ctx)
    str := fmt.Sprint(id)
    fmt.Println("next id=" + str)
    acct := msg.Account.String()
    accountStatus := keeper.GetAccountStatus(ctx, acct)
    accountStatus.NumQuotes++
    keeper.SetAccountStatus(ctx, acct, accountStatus)
	return sdk.Result{}
}
