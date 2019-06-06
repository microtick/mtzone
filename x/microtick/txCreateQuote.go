package microtick

import (
    "fmt"
    "time"
    
    "github.com/cosmos/cosmos-sdk/codec"
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

type CreateQuoteData struct {
    Id MicrotickId `json:"id"`
    Originator string `json:"originator"`
    Market MicrotickMarket `json:"market"`
    Duration MicrotickDurationName `json:"duration"`
    Spot MicrotickSpot `json:"spot"`
    Premium MicrotickPremium `json:"premium"`
    Consensus MicrotickSpot `json:"consensus"`
    Time time.Time `json:"time"`
    Backing MicrotickCoin `json:"backing"`
    Balance MicrotickCoin `json:"balance"`
    Commission MicrotickCoin `json:"commission"`
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
    return sdk.MustSortJSON(msgCdc.MustMarshalJSON(msg))
}

func (msg TxCreateQuote) GetSigners() []sdk.AccAddress {
    return []sdk.AccAddress{msg.Provider}
}

// Handler

func handleTxCreateQuote(ctx sdk.Context, keeper Keeper, 
    msg TxCreateQuote) sdk.Result {
        
    params := keeper.GetParams(ctx)
        
    if !keeper.HasDataMarket(ctx, msg.Market) {
        return sdk.ErrInternal("No such market: " + msg.Market).Result()
    }
    
    if !ValidMicrotickDuration(msg.Duration) {
        return sdk.ErrInternal(fmt.Sprintf("Invalid duration: %d", msg.Duration)).Result()
    }
    
    commission := NewMicrotickCoinFromDec(msg.Backing.Amount.Mul(params.CommissionQuotePercent))
    total := msg.Backing.Add(commission)
        
    // Subtract coins from quote provider
    keeper.WithdrawMicrotickCoin(ctx, msg.Provider, total)
    fmt.Printf("Create Commission: %s\n", commission.String())
    keeper.PoolCommission(ctx, commission)
	
	// DataActiveQuote
	
    id := keeper.GetNextActiveQuoteId(ctx)
     
    now := ctx.BlockHeader().Time
    dataActiveQuote := NewDataActiveQuote(now, id, msg.Market, msg.Duration, msg.Provider,
        msg.Backing, msg.Spot, msg.Premium)
    dataActiveQuote.ComputeQuantity()
    dataActiveQuote.Freeze(now, params)
    keeper.SetActiveQuote(ctx, dataActiveQuote)
    
    // DataAccountStatus
    
    accountStatus := keeper.GetAccountStatus(ctx, msg.Provider)
    accountStatus.ActiveQuotes.Insert(NewListItem(id, sdk.NewDec(int64(id))))
    accountStatus.NumQuotes++
    accountStatus.QuoteBacking = accountStatus.QuoteBacking.Add(msg.Backing)
    balance := accountStatus.Change
    coins := keeper.coinKeeper.GetCoins(ctx, msg.Provider)
    for i := 0; i < len(coins); i++ {
        if coins[i].Denom == TokenType {
            balance = balance.Add(NewMicrotickCoinFromInt(coins[i].Amount.Int64()))
        }
    }
    keeper.SetAccountStatus(ctx, msg.Provider, accountStatus)
    
    // DataMarket
    
    dataMarket, err2 := keeper.GetDataMarket(ctx, msg.Market)
    if err2 != nil {
        panic("Invalid market")
    }
    dataMarket.AddQuote(dataActiveQuote)
    dataMarket.factorIn(dataActiveQuote)
    keeper.SetDataMarket(ctx, dataMarket)
    
    // Tags
    tags := sdk.NewTags(
        "mtm.NewQuote", fmt.Sprintf("%d", id),
        fmt.Sprintf("quote.%d", id), "event.create",
        fmt.Sprintf("acct.%s", msg.Provider.String()), "quote.create",
        "mtm.MarketTick", msg.Market,
    )
    
    // Data
    data := CreateQuoteData {
      Id: id,
      Originator: "createQuote",
      Market: msg.Market,
      Duration: MicrotickDurationNameFromDur(msg.Duration),
      Spot: msg.Spot,
      Premium: msg.Premium,
      Consensus: dataMarket.Consensus,
      Time: now,
      Backing: msg.Backing,
      Balance: balance,
      Commission: commission,
    }
    bz, _ := codec.MarshalJSONIndent(keeper.cdc, data)
    
	return sdk.Result {
	    Data: bz,
	    Tags: tags,
	}
}
