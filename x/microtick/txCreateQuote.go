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
    Backing MicrotickCoin
    Spot MicrotickSpot
    Premium MicrotickPremium
}

func NewTxCreateQuote(market MicrotickMarket, dur MicrotickDuration, provider MicrotickAccount, 
    backing MicrotickCoin, spot MicrotickSpot, premium MicrotickPremium) TxCreateQuote {
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
    if !msg.Backing.IsPositive() {
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
        
    if !keeper.HasDataMarket(ctx, msg.Market) {
        return sdk.ErrInternal("No such market: " + msg.Market).Result()
    }
    
    if !ValidMicrotickDuration(msg.Duration) {
        return sdk.ErrInternal(fmt.Sprintf("Invalid duration: %d", msg.Duration)).Result()
    }
        
    // Subtract coins from quote provider
    keeper.WithdrawMicrotickCoin(ctx, msg.Provider, msg.Backing)
	
	// DataActiveQuote
	
    id := keeper.GetNextActiveQuoteId(ctx)
     
    dataActiveQuote := NewDataActiveQuote(id, msg.Market, msg.Duration, msg.Provider,
        msg.Backing, msg.Spot, msg.Premium)
    dataActiveQuote.ComputeQuantity()
    keeper.SetActiveQuote(ctx, dataActiveQuote)
    
    // DataAccountStatus
    
    accountStatus := keeper.GetAccountStatus(ctx, msg.Provider)
    accountStatus.ActiveQuotes.Insert(NewListItem(id, sdk.NewDec(int64(id))))
    accountStatus.NumQuotes++
    accountStatus.QuoteBacking = accountStatus.QuoteBacking.Plus(msg.Backing)
    keeper.SetAccountStatus(ctx, msg.Provider, accountStatus)
    
    // DataMarket
    
    dataMarket, err2 := keeper.GetDataMarket(ctx, msg.Market)
    if err2 != nil {
        panic("Invalid market")
    }
    dataMarket.AddQuote(dataActiveQuote)
    dataMarket.factorIn(dataActiveQuote)
    keeper.SetDataMarket(ctx, dataMarket)
    
    // Add tags
    
    tags := sdk.NewTags(
        "mtm.NewQuote", fmt.Sprintf("%d", id),
        fmt.Sprintf("quote.%d", id), "create",
        fmt.Sprintf("acct.%s", msg.Provider.String()), "quote.create",
        "mtm.MarketTick", msg.Market,
    )
    
	return sdk.Result {
	    Tags: tags,
	}
}
