package keeper

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
    mt "github.com/mjackson001/mtzone/x/microtick/types"
)

type LookupData struct {
    CallAsk sdk.Dec
    CallBid sdk.Dec
    PutAsk sdk.Dec
    PutBid sdk.Dec
}

type DataSyntheticQuote struct {
    Spot mt.MicrotickSpot `json:"spot"`
    Quantity mt.MicrotickQuantity `json:"quantity"`
}

type DataSyntheticBook struct {
    Name string `json:"name"`
    Weight mt.MicrotickQuantity `json:"weight"`
    Asks []DataSyntheticQuote `json:"asks"`
    Bids []DataSyntheticQuote `json:"bids"`
}

func (k Keeper) GetSyntheticBook(ctx sdk.Context, dm DataMarket, name string) DataSyntheticBook {
    for i := 0; i < len(dm.OrderBooks); i++ {
        if dm.OrderBooks[i].Name == name {
            // found
            orderBook := dm.OrderBooks[i]
            
            // create lookup tables
            lookup := make(map[mt.MicrotickId]LookupData)
            askQuantities := make(map[mt.MicrotickId]sdk.Dec)
            bidQuantities := make(map[mt.MicrotickId]sdk.Dec)
            for i := 0; i < len(orderBook.CallAsks.Data); i++ {
            	id := orderBook.CallAsks.Data[i].Id
                q, _ := k.GetActiveQuote(ctx, id)
                lookup[id] = LookupData {
                    CallAsk: q.CallAsk(dm.Consensus).Amount,
                    CallBid: q.CallBid(dm.Consensus).Amount,
                    PutAsk: q.PutAsk(dm.Consensus).Amount,
                    PutBid: q.PutBid(dm.Consensus).Amount,
                }
                askQuantities[id] = q.Quantity.Amount
                bidQuantities[id] = q.Quantity.Amount
            }
            
            // calculate asks
            asks := make([]DataSyntheticQuote, 0)
            askIndex := 0
            bidIndex := len(orderBook.PutBids.Data) - 1
            for askIndex < len(orderBook.CallAsks.Data) && bidIndex >= 0 {
                askId := orderBook.CallAsks.Data[askIndex].Id
                bidId := orderBook.PutBids.Data[bidIndex].Id
                if askQuantities[askId].IsPositive() {
                    if askQuantities[bidId].IsPositive() {
                        var fillAmount sdk.Dec
                        
                        if askId == bidId {
                            fillAmount = askQuantities[askId].QuoInt64(2)
                            askIndex++
                            bidIndex--
                        } else if askQuantities[askId].GT(askQuantities[bidId]) {
                            fillAmount = askQuantities[bidId]
                            askQuantities[askId] = askQuantities[askId].Sub(fillAmount)
                            askQuantities[bidId] = sdk.ZeroDec()
                            bidIndex--
                        } else {
                            fillAmount = askQuantities[askId]
                            askQuantities[bidId] = askQuantities[bidId].Sub(fillAmount)
                            askQuantities[askId] = sdk.ZeroDec()
                            askIndex++
                        }
                        
                        spot := dm.Consensus.Amount.Add(lookup[askId].CallAsk).Sub(lookup[bidId].PutBid)
                        asks = append(asks, DataSyntheticQuote {
                            Spot: mt.NewMicrotickSpotFromDec(spot), 
                            Quantity: mt.NewMicrotickQuantityFromDec(fillAmount),
                        })
                    } else {
                        bidIndex--
                    }
                } else {
                    askIndex++
                }
            }
            
            // calculate bids
            bids := make([]DataSyntheticQuote, 0)
            askIndex = 0
            bidIndex = len(orderBook.CallBids.Data) - 1
            for askIndex < len(orderBook.PutAsks.Data) && bidIndex >= 0 {
                askId := orderBook.PutAsks.Data[askIndex].Id
                bidId := orderBook.CallBids.Data[bidIndex].Id
                if bidQuantities[askId].IsPositive() {
                    if bidQuantities[bidId].IsPositive() {
                        var fillAmount sdk.Dec
                        
                        if askId == bidId {
                            fillAmount = bidQuantities[askId].QuoInt64(2)
                            askIndex++
                            bidIndex--
                        } else if bidQuantities[askId].GT(bidQuantities[bidId]) {
                            fillAmount = bidQuantities[bidId]
                            bidQuantities[askId] = bidQuantities[askId].Sub(fillAmount)
                            bidQuantities[bidId] = sdk.ZeroDec()
                            bidIndex--
                        } else {
                            fillAmount = bidQuantities[askId]
                            bidQuantities[bidId] = bidQuantities[bidId].Sub(fillAmount)
                            bidQuantities[askId] = sdk.ZeroDec()
                            askIndex++
                        }
                        
                        spot := dm.Consensus.Amount.Sub(lookup[askId].PutAsk).Add(lookup[bidId].CallBid)
                        bids = append(bids, DataSyntheticQuote {
                            Spot: mt.NewMicrotickSpotFromDec(spot), 
                            Quantity: mt.NewMicrotickQuantityFromDec(fillAmount),
                        })
                    } else {
                        bidIndex--
                    }
                } else {
                    askIndex++
                }
            }
            
            return DataSyntheticBook {
                Name: name,
                Weight: mt.NewMicrotickQuantityFromDec(orderBook.SumWeight.Amount.QuoInt64(2)),
                Asks: asks,
                Bids: bids,
            }
        }
    }
    panic("Invalid duration name")
}
