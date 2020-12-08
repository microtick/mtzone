package keeper

import (
    "time"
    sdk "github.com/cosmos/cosmos-sdk/types"
    
    mt "gitlab.com/microtick/mtzone/x/microtick/types"
)

func NewDataMarket(market mt.MicrotickMarket, description string, durs []mt.MicrotickDurationName) DataMarket {
    orderBooks := make([]DataOrderBook, len(durs))
    for i := 0; i < len(durs); i++ {
        orderBooks[i] = newOrderBook(durs[i])
    }
    return DataMarket {
        Market: market,
        Description: description,
        Consensus: mt.NewMicrotickSpotFromInt(0),
        OrderBooks: orderBooks,
        TotalBacking: mt.NewMicrotickCoinFromInt(0),
        TotalSpots: sdk.ZeroDec(),
        TotalWeight: mt.NewMicrotickQuantityFromInt(0),
    }
}

func newOrderBook(name mt.MicrotickDurationName) DataOrderBook {
    return DataOrderBook {
        Name: name,
        CallAsks: NewOrderedList(),
        CallBids: NewOrderedList(),
        PutAsks: NewOrderedList(),
        PutBids: NewOrderedList(),
        SumBacking: mt.NewMicrotickCoinFromInt(0),
        SumWeight: mt.NewMicrotickQuantityFromInt(0),
    }
}

func (dm *DataMarket) GetOrderBook(name mt.MicrotickDurationName) DataOrderBook {
    for i := 0; i < len(dm.OrderBooks); i++ {
        if dm.OrderBooks[i].Name == name {
            return dm.OrderBooks[i]
        }
    }
    panic("Invalid duration name")
}

func (dm *DataMarket) SetOrderBook(name mt.MicrotickDurationName, ob DataOrderBook) {
    for i := 0; i < len(dm.OrderBooks); i++ {
        if dm.OrderBooks[i].Name == name {
            dm.OrderBooks[i] = ob
            return
        }
    }
    panic("Invalid duration name")
}

func (dm *DataMarket) FactorIn(quote DataActiveQuote, testInvariants bool) bool {
    // FactorIn is called from Tx's that originate from the quote provider, i.e.
    // create quote, update quote and from Tx's that result from a counter party, 
    // i.e. trade market, trade limit.
    // We only want to test invariants when we're not trade matching, because
    // the invariants apply only to creation or updates, not from the results
    // of market action over time.  (quotes can go stale, etc)

    dm.TotalBacking = dm.TotalBacking.Add(quote.Backing)
    dm.TotalSpots = dm.TotalSpots.Add(quote.Spot.Amount.Mul(quote.Quantity.Amount))
    dm.TotalWeight = dm.TotalWeight.Add(quote.Quantity)
    if dm.TotalWeight.Amount.IsPositive() {
        dm.Consensus = mt.NewMicrotickSpotFromDec(dm.TotalSpots.Quo(dm.TotalWeight.Amount))
    }
    
    // Test quote invariant:
    if testInvariants {
        // Spot 2x limitation
        // A quote cannot be placed or updated that will be a free call or put on the 
        // resulting order book (spot more than 2x premium from resulting consensus)
        // Purpose: protects market maker from damaging quotes
        if quote.Spot.Amount.Sub(quote.Ask.Amount.MulInt64(2)).GT(dm.Consensus.Amount) {
            //fmt.Printf("Failed Spot Invariant (1): %d\n", quote.Id)
            return false
        }
        if quote.Spot.Amount.Add(quote.Ask.Amount.MulInt64(2)).LT(dm.Consensus.Amount) {
            //fmt.Printf("Failed Spot Invariant (2): %d\n", quote.Id)
            return false
        }
        // Premium must be less than 1/2 spot, otherwise if consensus moves less than spot,
        // it would be possible for the premium to reach negative price territory
        if quote.Spot.Amount.QuoInt64(2).LT(quote.Ask.Amount) {
            return false
        }
    }
    
    orderBook := dm.GetOrderBook(quote.DurationName)
    orderBook.SumBacking = orderBook.SumBacking.Add(quote.Backing)
    orderBook.SumWeight = orderBook.SumWeight.Add(quote.Quantity)
    
    dm.SetOrderBook(quote.DurationName, orderBook)
    return true
}

func (dm *DataMarket) FactorOut(quote DataActiveQuote) {
    dm.TotalBacking = dm.TotalBacking.Sub(quote.Backing)
    dm.TotalSpots = dm.TotalSpots.Sub(quote.Spot.Amount.Mul(quote.Quantity.Amount))
    dm.TotalWeight = dm.TotalWeight.Sub(quote.Quantity)
    if dm.TotalWeight.Amount.IsPositive() {
        dm.Consensus = mt.NewMicrotickSpotFromDec(dm.TotalSpots.Quo(dm.TotalWeight.Amount))
    }
    
    orderBook := dm.GetOrderBook(quote.DurationName)
    orderBook.SumBacking = orderBook.SumBacking.Sub(quote.Backing)
    orderBook.SumWeight = orderBook.SumWeight.Sub(quote.Quantity)
    dm.SetOrderBook(quote.DurationName, orderBook)
}

func (dm *DataMarket) AddQuote(quote DataActiveQuote) {
    orderBook := dm.GetOrderBook(quote.DurationName)
    
    spotDiv2 := quote.Spot.Amount.QuoInt64(2)
    orderBook.CallAsks.Insert(NewListItem(quote.Id, quote.Ask.Amount.Add(spotDiv2)))
    orderBook.CallBids.Insert(NewListItem(quote.Id, quote.Bid.Amount.Add(spotDiv2)))
    orderBook.PutAsks.Insert(NewListItem(quote.Id, quote.Ask.Amount.Sub(spotDiv2)))
    orderBook.PutBids.Insert(NewListItem(quote.Id, quote.Bid.Amount.Sub(spotDiv2)))
    
    dm.Quotes.Insert(NewListItem(quote.Id, sdk.NewDec(quote.CanModify)))
    
    dm.SetOrderBook(quote.DurationName, orderBook)
}

func (dm *DataMarket) DeleteQuote(quote DataActiveQuote) {
    orderBook := dm.GetOrderBook(quote.DurationName)
    
    orderBook.CallAsks.Delete(quote.Id)
    orderBook.CallBids.Delete(quote.Id)
    orderBook.PutAsks.Delete(quote.Id)
    orderBook.PutBids.Delete(quote.Id)
    
    dm.Quotes.Delete(quote.Id)
    
    dm.SetOrderBook(quote.DurationName, orderBook)
}

func (dm *DataMarket) CanSettle(now time.Time) bool {
    if len(dm.Quotes.Data) == 0 {
        return false
    }
    if dm.Quotes.Data[0].Value.GT(sdk.NewDec(now.Unix())) {
        return false
    }
    return true
}
