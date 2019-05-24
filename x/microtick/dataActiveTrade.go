package microtick

import (
    "fmt"
    "time"
    sdk "github.com/cosmos/cosmos-sdk/types"
)

type DataActiveTrade struct {
    Id MicrotickId `json:"id"`
    Market MicrotickMarket `json:"market"`
    // Duration is only a tag at this point, not functional
    Duration MicrotickDurationName `json:"duration"`
    Type MicrotickTradeType `json:"type"`
    CounterParties []DataCounterParty `json:"counterParties"`
    Long MicrotickAccount `json:"long"`
    Backing MicrotickCoin `json:"backing"`
    Cost MicrotickCoin `json:"cost"`
    FilledQuantity MicrotickQuantity `json:"quantity"`
    Start time.Time `json:"start"`
    Expiration time.Time `json:"expiration"`
    Strike MicrotickSpot `json:"strike"`
    Commission MicrotickCoin `json:"commission"`
    SettleIncentive MicrotickCoin `json:"settleIncentive"`
}

func NewDataActiveTrade(now time.Time, market MicrotickMarket, dur MicrotickDuration,
    ttype MicrotickTradeType, long MicrotickAccount, strike MicrotickSpot,
    commission MicrotickCoin, settleIncentive MicrotickCoin) DataActiveTrade {
        
    expire, err := time.ParseDuration(fmt.Sprintf("%d", dur) + "s")
    if err != nil {
        panic("invalid time")
    }
    return DataActiveTrade {
        Id: 0, // set actual trade ID later after premium has been verified
        Market: market,
        Duration: MicrotickDurationNameFromDur(dur),
        Type: ttype,
        Long: long,
        Backing: NewMicrotickCoinFromInt(0),
        Cost: NewMicrotickCoinFromInt(0),
        FilledQuantity: NewMicrotickQuantityFromInt(0), // computed later
        Start: now,
        Expiration: now.Add(expire),
        Strike: strike,
        Commission: commission,
        SettleIncentive: settleIncentive,
    }
}

type DataQuoteParams struct {
    Id MicrotickId `json:"quoteId"`
    Premium MicrotickPremium `json:"premium"`
    Quantity MicrotickQuantity `json:"quantity"`
    Spot MicrotickSpot `json:"spot"`
}

func NewDataQuoteParams(id MicrotickId, premium MicrotickPremium, quantity MicrotickQuantity,
    spot MicrotickSpot) DataQuoteParams {
    return DataQuoteParams {
        Id: id,
        Premium: premium,
        Quantity: quantity,
        Spot: spot,
    }
}

type DataCounterParty struct {
    Backing MicrotickCoin `json:"backing"`
    Cost MicrotickCoin `json:"premium"`
    FilledQuantity MicrotickQuantity `json:"quantity"`
    Short MicrotickAccount `json:"short"`
    Quoted DataQuoteParams `json:"quoted"`
}

func NewDataCounterParty(backing MicrotickCoin, final bool, cost MicrotickCoin, 
    quantity MicrotickQuantity)  DataCounterParty {
    return DataCounterParty {
        Backing: backing,
        Cost: cost,
        FilledQuantity: quantity,
    }
}

func (dat DataActiveTrade) CurrentValue(current MicrotickSpot) MicrotickCoin {
    strike := dat.Strike.Amount
    var delta sdk.Dec
    if dat.Type {
        // Put
        delta = strike.Sub(current.Amount)
    } else {
        // Call
        delta = current.Amount.Sub(strike)
    }
    if delta.IsNegative() {
        return NewMicrotickCoinFromInt(0)
    }
    value := delta.Mul(dat.FilledQuantity.Amount)
    if value.GT(dat.Backing.Amount) {
        value = dat.Backing.Amount
    }
    return NewMicrotickCoinFromDec(value)
}

type CounterPartySettlement struct {
    Settle MicrotickCoin
    Refund MicrotickCoin
    RefundAddress MicrotickAccount
    Backing MicrotickCoin
}

func (dat DataActiveTrade) CounterPartySettlements(current MicrotickSpot) []CounterPartySettlement {
    strike := dat.Strike.Amount
    var delta sdk.Dec
    if dat.Type {
        // Put
        delta = strike.Sub(current.Amount)
    } else {
        // Call
        delta = current.Amount.Sub(strike)
    }
    if delta.IsNegative() {
        delta = sdk.ZeroDec()
    }
    var result []CounterPartySettlement
    for i := 0; i < len(dat.CounterParties); i++ {
        cp := dat.CounterParties[i]
        settle := delta.Mul(cp.FilledQuantity.Amount)
        if settle.GT(cp.Backing.Amount) {
            settle = cp.Backing.Amount
        }
        refund := cp.Backing.Amount.Sub(settle)
        result = append(result, CounterPartySettlement {
            Settle: NewMicrotickCoinFromDec(settle),
            Refund: NewMicrotickCoinFromDec(refund),
            RefundAddress: cp.Short,
            Backing: cp.Backing,
        })
    }
    return result   
}
