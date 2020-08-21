package keeper

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
    mt "github.com/mjackson001/mtzone/x/microtick/types"
)

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
            
            // calculate asks
            quotes := make(map[mt.MicrotickId]DataActiveQuote)
            for i := 0; i < len(orderBook.CallAsks.Data); i++ {
            	id := orderBook.CallAsks.Data[i].Id
                quotes[id], _ = k.GetActiveQuote(ctx, id)
            }
            asks := make([]DataSyntheticQuote, 0)
            qty := sdk.NewDec(0)
            askIndex := 0
            bidIndex := len(orderBook.PutBids.Data) - 1
            for askIndex < len(orderBook.CallAsks.Data) {
                askData := orderBook.CallAsks.Data[askIndex]
                askQuote := quotes[askData.Id]
                bidData := orderBook.PutBids.Data[bidIndex]
                bidQuote := quotes[bidData.Id]
                var fillAmount sdk.Dec
                spot := dm.Consensus.Amount.Add(askQuote.CallAsk(dm.Consensus).Amount).Sub(bidQuote.PutBid(dm.Consensus).Amount)
                if askQuote.Quantity.Amount.Equal(bidQuote.Quantity.Amount) {
                    if askData.Id == bidData.Id {
                        fillAmount = askQuote.Quantity.Amount.QuoInt64(2)
                    } else {
                        fillAmount = askQuote.Quantity.Amount
                    }
                    askIndex++
                    bidIndex--
                } else if askQuote.Quantity.Amount.GT(bidQuote.Quantity.Amount) {
                    fillAmount = bidQuote.Quantity.Amount
                    askQuote.Quantity = mt.NewMicrotickQuantityFromDec(askQuote.Quantity.Amount.Sub(bidQuote.Quantity.Amount))
                    quotes[askQuote.Id] = askQuote
                    bidIndex--
                } else {
                    fillAmount = askQuote.Quantity.Amount
                    bidQuote.Quantity = mt.NewMicrotickQuantityFromDec(bidQuote.Quantity.Amount.Sub(askQuote.Quantity.Amount))
                    quotes[bidQuote.Id] = bidQuote
                    askIndex++
                }
                qty = qty.Add(fillAmount)
                asks = append(asks, DataSyntheticQuote {
                    Spot: mt.NewMicrotickSpotFromDec(spot), 
                    Quantity: mt.NewMicrotickQuantityFromDec(fillAmount),
                })
            }
            
            // calculate bids
            quotes = make(map[mt.MicrotickId]DataActiveQuote)
            for i := 0; i < len(orderBook.PutAsks.Data); i++ {
            	id := orderBook.PutAsks.Data[i].Id
                quotes[id], _ = k.GetActiveQuote(ctx, id)
            }
            bids := make([]DataSyntheticQuote, 0)
            qty = sdk.NewDec(0)
            askIndex = 0
            bidIndex = len(orderBook.CallBids.Data) - 1
            for askIndex < len(orderBook.PutAsks.Data) && bidIndex >= 0 {
                askData := orderBook.PutAsks.Data[askIndex]
                askQuote := quotes[askData.Id]
                bidData := orderBook.CallBids.Data[bidIndex]
                bidQuote := quotes[bidData.Id]
                var fillAmount sdk.Dec
                spot := dm.Consensus.Amount.Sub(askQuote.PutAsk(dm.Consensus).Amount).Add(bidQuote.CallBid(dm.Consensus).Amount)
                if askQuote.Quantity.Amount.Equal(bidQuote.Quantity.Amount) {
                    if askData.Id == bidData.Id {
                        fillAmount = askQuote.Quantity.Amount.QuoInt64(2)
                    } else {
                        fillAmount = askQuote.Quantity.Amount
                    }
                    askIndex++
                    bidIndex--
                } else if askQuote.Quantity.Amount.GT(bidQuote.Quantity.Amount) {
                    fillAmount = bidQuote.Quantity.Amount
                    askQuote.Quantity = mt.NewMicrotickQuantityFromDec(askQuote.Quantity.Amount.Sub(bidQuote.Quantity.Amount))
                    quotes[askQuote.Id] = askQuote
                    bidIndex--
                } else {
                    fillAmount = askQuote.Quantity.Amount
                    bidQuote.Quantity = mt.NewMicrotickQuantityFromDec(bidQuote.Quantity.Amount.Sub(askQuote.Quantity.Amount))
                    quotes[bidQuote.Id] = bidQuote
                    askIndex++
                }
                qty = qty.Add(fillAmount)
                bids = append(bids, DataSyntheticQuote {
                    Spot: mt.NewMicrotickSpotFromDec(spot), 
                    Quantity: mt.NewMicrotickQuantityFromDec(fillAmount),
                })
            }
            
            return DataSyntheticBook {
                Name: name,
                Weight: mt.NewMicrotickQuantityFromDec(qty),
                Asks: asks,
                Bids: bids,
            }
        }
    }
    panic("Invalid duration name")
}
