package keeper

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
    
    mt "github.com/mjackson001/mtzone/x/microtick/types"
)

type QuoteFillInfo struct {
    Quote DataActiveQuote
    BoughtQuantity sdk.Dec
    Cost mt.MicrotickCoin
    FinalFill bool
}

type FetchQuoteFunc func(mt.MicrotickId) DataActiveQuote

type Matcher struct {
    Trade DataActiveTrade
    TotalQuantity sdk.Dec
    TotalCost sdk.Dec
    FillInfo []QuoteFillInfo
    FetchQuote FetchQuoteFunc
}

func NewMatcher(trade DataActiveTrade, fetchQuoteFunc FetchQuoteFunc) Matcher {
    return Matcher {
        Trade: trade,
        TotalQuantity: sdk.ZeroDec(),
        TotalCost: sdk.ZeroDec(),
        FetchQuote: fetchQuoteFunc,
    }
}

func (matcher *Matcher) AssignCounterparties(ctx sdk.Context, keeper MicrotickKeeper, market *DataMarket) {
    for i := 0; i < len(matcher.FillInfo); i++ {
        thisFill := matcher.FillInfo[i]
        thisQuote := thisFill.Quote
        
        // We save the current quote parameters in the trade because these may change
        // and we use them for historical and accounting purposes
        params := DataQuoteParams {
            Id: thisQuote.Id,
            Premium: thisQuote.Premium,
            Quantity: thisQuote.Quantity,
            Spot: thisQuote.Spot,
        }
        
        // Pay premium
        keeper.DepositMicrotickCoin(ctx, thisQuote.Provider, thisFill.Cost)
        
        accountStatus := keeper.GetAccountStatus(ctx, thisQuote.Provider)
        
        // Adjust quote
        market.FactorOut(thisQuote)
        
        var transferredBacking mt.MicrotickCoin
        if thisFill.BoughtQuantity.GTE(thisQuote.Quantity.Amount) {
            transferredBacking = thisQuote.Backing
        } else {
            // For a partially filled quote, the backing transferred should be proportional to
            // the quantity purchased.
            transferredBacking = mt.NewMicrotickCoinFromDec(thisQuote.Backing.Amount.Mul(thisFill.BoughtQuantity.Quo(thisQuote.Quantity.Amount)))
        }
        
        // Subtract out bought quantity and corresponding backing
        thisQuote.Quantity = mt.NewMicrotickQuantityFromDec(thisQuote.Quantity.Amount.Sub(thisFill.BoughtQuantity))
        thisQuote.Backing = thisQuote.Backing.Sub(transferredBacking)
        
        if thisQuote.Quantity.Amount.IsZero() {
            // If no quantity is left, delete quote from market, active quote list, and
            // account active quote list
            market.DeleteQuote(thisQuote)
            keeper.DeleteActiveQuote(ctx, thisQuote.Id)
            accountStatus.ActiveQuotes.Delete(thisQuote.Id)
        } else {
            // else, factor quote back into market consensus
            market.FactorIn(thisQuote)
            keeper.SetActiveQuote(ctx, thisQuote)
        }
        
        // Adjust trade
        matcher.Trade.Backing = matcher.Trade.Backing.Add(transferredBacking)
        matcher.Trade.Cost = matcher.Trade.Cost.Add(thisFill.Cost)
        matcher.Trade.FilledQuantity = mt.NewMicrotickQuantityFromDec(matcher.Trade.FilledQuantity.Amount.Add(thisFill.BoughtQuantity))
        
        // Update the account status of this counterparty
        accountStatus.ActiveTrades.Insert(NewListItem(matcher.Trade.Id, sdk.NewDec(matcher.Trade.Expiration.UnixNano())))
        accountStatus.QuoteBacking = accountStatus.QuoteBacking.Sub(transferredBacking)
        accountStatus.TradeBacking = accountStatus.TradeBacking.Add(transferredBacking)
        
        // Save the counterparty account status in the store
        keeper.SetAccountStatus(ctx, thisQuote.Provider, accountStatus)
        
        // Append this counter party fill to the trade counterparty list
        matcher.Trade.CounterParties = append(matcher.Trade.CounterParties, DataCounterParty {
            Backing: transferredBacking,
            Cost: thisFill.Cost,
            FilledQuantity: mt.NewMicrotickQuantityFromDec(thisFill.BoughtQuantity),
            Short: thisQuote.Provider,
            Quoted: params,
            Balance: keeper.GetTotalBalance(ctx, thisQuote.Provider),
        })
    }
}

func (matcher Matcher) HasQuantity() bool {
    return matcher.TotalQuantity.GT(sdk.ZeroDec())
}