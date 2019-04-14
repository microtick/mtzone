package microtick

import (
    "encoding/json"
    
    sdk "github.com/cosmos/cosmos-sdk/types"
)

type TxCreateMarket struct {
    Account MicrotickAccount
    Market MicrotickMarket
}

func NewTxCreateMarket(account MicrotickAccount, market MicrotickMarket) TxCreateMarket {
    return TxCreateMarket {
        Account: account,
        Market: market,
    }
}

func (msg TxCreateMarket) Route() string { return "microtick" }

func (msg TxCreateMarket) Type() string { return "create_market" }

func (msg TxCreateMarket) ValidateBasic() sdk.Error {
    if msg.Account.Empty() {
        return sdk.ErrInvalidAddress(msg.Account.String())
    }
    if len(msg.Market) == 0 {
        return sdk.ErrInternal("Invalid market: " + msg.Market)
    }
    return nil
}

func (msg TxCreateMarket) GetSignBytes() []byte {
    b, err := json.Marshal(msg)
    if err != nil {
        panic(err)
    }
    return sdk.MustSortJSON(b)
}

func (msg TxCreateMarket) GetSigners() []sdk.AccAddress {
    return []sdk.AccAddress{msg.Account}
}

// Handler

func handleTxCreateMarket(ctx sdk.Context, keeper Keeper, msg TxCreateMarket) sdk.Result {
    if !keeper.HasDataMarket(ctx, msg.Market) {
        keeper.SetDataMarket(ctx, NewDataMarket(msg.Market))
    }
    return sdk.Result{}
}