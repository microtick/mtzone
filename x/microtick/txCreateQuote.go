package microtick

import (
    "fmt"
    "encoding/json"
    
    sdk "github.com/cosmos/cosmos-sdk/types"
)

type TxCreateQuote struct {
    Market MicrotickMarket
    Duration MicrotickDuration
    Provider sdk.AccAddress
    Backing sdk.Coins
    Spot MicrotickSpot
    Premium MicrotickPremium
}

func NewTxCreateQuote(market MicrotickMarket, dur MicrotickDuration, provider sdk.AccAddress, 
    backing sdk.Coins, spot MicrotickSpot, premium MicrotickPremium) TxCreateQuote {
    return TxCreateQuote {
        Market: market,
        Duration: dur,
        Provider: provider,
        Backing: backing,
        Spot: spot,
        Premium: premium,
    }
}

func (msg TxCreateQuote) Route() string { return "microtick" }

func (msg TxCreateQuote) Type() string { return "create_quote" }

func (msg TxCreateQuote) ValidateBasic() sdk.Error {
    if len(msg.Market) == 0 {
        return sdk.ErrInternal("Unknown market")
    }
    if msg.Provider.Empty() {
        return sdk.ErrInvalidAddress(msg.Provider.String())
    }
    if !msg.Backing.IsAllPositive() {
        return sdk.ErrInsufficientCoins("Backing must be positive")
    }
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
    return []sdk.AccAddress{msg.Provider}
}

// Handler

func handleTxCreateQuote(ctx sdk.Context, keeper Keeper, 
    msg TxCreateQuote) sdk.Result {
    // Subtract coins from quote provider
  	_, _, err := keeper.coinKeeper.SubtractCoins(ctx, msg.Provider, msg.Backing) 
	if err != nil {
		return sdk.ErrInsufficientCoins("Buyer does not have enough coins").Result()
	}
	
    id := keeper.GetNextActiveQuoteId(ctx)
    provider := msg.Provider.String()
     
    dataActiveQuote := NewDataActiveQuote(id, msg.Market, msg.Duration, provider,
        msg.Backing, msg.Spot, msg.Premium)
    keeper.SetActiveQuote(ctx, dataActiveQuote)
    
    accountStatus := keeper.GetAccountStatus(ctx, provider)
    fmt.Printf("before: %+v\n", accountStatus.ActiveQuotes)
    accountStatus.ActiveQuotes.Insert(NewListItem(uint(id), 
        int(id)))
    fmt.Printf("after: %+v\n", accountStatus.ActiveQuotes)
    accountStatus.NumQuotes++
    keeper.SetAccountStatus(ctx, provider, accountStatus)
	return sdk.Result{}
}
