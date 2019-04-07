package microtick

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
)

type QuoteFillInfo struct {
    Quote DataActiveQuote
    BoughtQuantity sdk.Dec
    PaidPremium MicrotickCoin
}

type FetchQuoteFunc func(MicrotickId) DataActiveQuote

type Matcher struct {
    Trade DataActiveTrade
    TotalQuantity sdk.Dec
    TotalPremium sdk.Dec
    FillInfo []QuoteFillInfo
    FetchQuote FetchQuoteFunc
}

func NewMatcher(trade DataActiveTrade, fetchQuoteFunc FetchQuoteFunc) Matcher {
    return Matcher {
        Trade: trade,
        TotalQuantity: sdk.ZeroDec(),
        TotalPremium: sdk.ZeroDec(),
        FetchQuote: fetchQuoteFunc,
    }
}

func (matcher *Matcher) AssignCounterparties(ctx sdk.Context, keeper Keeper, market *DataMarket) {
    for i := 0; i < len(matcher.FillInfo); i++ {
        thisFill := matcher.FillInfo[i]
        thisQuote := thisFill.Quote
        
        // Pay premium
        keeper.DepositDecCoin(ctx, thisQuote.Provider, NewMicrotickCoinFromPremium(thisFill.PaidPremium))
        
        accountStatus := keeper.GetAccountStatus(ctx, thisQuote.Provider)
        
        // Adjust quote
        market.factorOut(thisQuote)
        
        var transferredBacking MicrotickCoin
        if thisFill.BoughtQuantity.GTE(thisQuote.Quantity.Amount) {
            transferredBacking = thisQuote.Backing
        } else {
            // For a partially filled quote, the backing transferred should be proportional to
            // the quantity purchased.
            transferredBacking = NewMicrotickCoinFromDec(thisQuote.Backing.Amount.Mul(thisFill.BoughtQuantity.Quo(thisQuote.Quantity.Amount)))
        }
        
        // Subtract out bought quantity and corresponding backing
        thisQuote.Quantity = NewMicrotickQuantityFromDec(thisQuote.Quantity.Amount.Sub(thisFill.BoughtQuantity))
        thisQuote.Backing = thisQuote.Backing.Minus(transferredBacking)
        
        if thisQuote.Quantity.Amount.IsZero() {
            // If no quantity is left, delete quote from market, active quote list, and
            // account active quote list
            market.DeleteQuote(thisQuote)
            keeper.DeleteActiveQuote(ctx, thisQuote.Id)
            accountStatus.ActiveQuotes.Delete(thisQuote.Id)
        } else {
            // else, factor quote back into market consensus
            market.factorIn(thisQuote)
            keeper.SetActiveQuote(ctx, thisQuote)
        }
        
        // Adjust trade
        matcher.Trade.Backing = matcher.Trade.Backing.Plus(transferredBacking)
        matcher.Trade.Premium = matcher.Trade.Premium.Plus(thisFill.PaidPremium)
        matcher.Trade.FilledQuantity = NewMicrotickQuantityFromDec(matcher.Trade.FilledQuantity.Amount.Add(thisFill.BoughtQuantity))
        
        // We save the current quote parameters in the trade because these may change
        // and we use them for historical and accounting purposes
        params := DataQuoteParams {
            Id: thisQuote.Id,
            Premium: thisQuote.Premium,
            Quantity: thisQuote.Quantity,
            Spot: thisQuote.Spot,
        }
        // Append this counter party fill to the trade counterparty list
        matcher.Trade.CounterParties = append(matcher.Trade.CounterParties, DataCounterParty {
            Backing: transferredBacking,
            PaidPremium: thisFill.PaidPremium,
            FilledQuantity: NewMicrotickQuantityFromDec(thisFill.BoughtQuantity),
            Short: thisQuote.Provider,
            Quoted: params,
        })
        
        // Update the account status of this counterparty
        accountStatus.ActiveTrades.Insert(NewListItem(matcher.Trade.Id, sdk.NewDec(matcher.Trade.Expiration.UnixNano())))
        accountStatus.QuoteBacking = accountStatus.QuoteBacking.Minus(transferredBacking)
        accountStatus.TradeBacking = accountStatus.TradeBacking.Plus(transferredBacking)
        
        // Save the counterparty account status in the store
        keeper.SetAccountStatus(ctx, thisQuote.Provider, accountStatus)
    }
}

func (matcher Matcher) hasQuantity() bool {
    return matcher.TotalQuantity.GT(sdk.ZeroDec())
}