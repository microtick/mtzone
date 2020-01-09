package msg

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
    
    mt "github.com/mjackson001/mtzone/x/microtick/types"
    "github.com/mjackson001/mtzone/x/microtick/keeper"
)

type TxCreateMarket struct {
    Account mt.MicrotickAccount
    Market mt.MicrotickMarket
}

func NewTxCreateMarket(account mt.MicrotickAccount, market mt.MicrotickMarket) TxCreateMarket {
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
    return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

func (msg TxCreateMarket) GetSigners() []sdk.AccAddress {
    return []sdk.AccAddress{msg.Account}
}

// Handler

func HandleTxCreateMarket(ctx sdk.Context, mtKeeper keeper.Keeper, msg TxCreateMarket) sdk.Result {
    if !mtKeeper.HasDataMarket(ctx, msg.Market) {
        mtKeeper.SetDataMarket(ctx, keeper.NewDataMarket(msg.Market))
    }
    
    event := sdk.NewEvent(
        sdk.EventTypeMessage,
        sdk.NewAttribute(sdk.AttributeKeyModule, mt.ModuleKey),
    )
    events := []sdk.Event{ event }
    events = append(events, sdk.NewEvent(
        sdk.EventTypeMessage,
        sdk.NewAttribute("market", msg.Market),
    ))
    
    return sdk.Result{
        Events: events,
    }
}