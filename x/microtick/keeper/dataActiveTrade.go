package keeper

import (
    "fmt"
    "time"
    sdk "github.com/cosmos/cosmos-sdk/types"
    
    mt "gitlab.com/microtick/mtzone/x/microtick/types"
)

type DataActiveTrade struct {
    Id mt.MicrotickId `json:"id"`
    Market mt.MicrotickMarket `json:"market"`
    // DurationName is only a tag at this point, not functional
    DurationName mt.MicrotickDurationName `json:"duration"`
    Type mt.MicrotickTradeType `json:"type"`
    CounterParties []DataCounterParty `json:"counterparties"`
    Long mt.MicrotickAccount `json:"long"`
    Backing mt.MicrotickCoin `json:"backing"`
    Cost mt.MicrotickCoin `json:"cost"`
    FilledQuantity mt.MicrotickQuantity `json:"quantity"`
    Start time.Time `json:"start"`
    Expiration time.Time `json:"expiration"`
    Strike mt.MicrotickSpot `json:"strike"`
    Commission mt.MicrotickCoin `json:"commission"`
    SettleIncentive mt.MicrotickCoin `json:"settleIncentive"`
}

func NewDataActiveTrade(now time.Time, market mt.MicrotickMarket, 
    dur mt.MicrotickDurationName, durSeconds mt.MicrotickDuration,
    ttype mt.MicrotickTradeType, long mt.MicrotickAccount, strike mt.MicrotickSpot,
    commission mt.MicrotickCoin, settleIncentive mt.MicrotickCoin) DataActiveTrade {
        
    expire, err := time.ParseDuration(fmt.Sprintf("%d", durSeconds) + "s")
    if err != nil {
        panic("invalid time")
    }
    return DataActiveTrade {
        Id: 0, // set actual trade ID later after premium has been verified
        Market: market,
        DurationName: dur,
        Type: ttype,
        Long: long,
        Backing: mt.NewMicrotickCoinFromExtCoinInt(0),
        Cost: mt.NewMicrotickCoinFromExtCoinInt(0),
        FilledQuantity: mt.NewMicrotickQuantityFromInt(0), // computed later
        Start: now,
        Expiration: now.Add(expire),
        Strike: strike,
        Commission: commission,
        SettleIncentive: settleIncentive,
    }
}

type DataQuoteParams struct {
    Id mt.MicrotickId `json:"id"`
    Premium mt.MicrotickPremium `json:"premium"`
    Quantity mt.MicrotickQuantity `json:"quantity"`
    Spot mt.MicrotickSpot `json:"spot"`
}

func NewDataQuoteParams(id mt.MicrotickId, premium mt.MicrotickPremium, quantity mt.MicrotickQuantity,
    spot mt.MicrotickSpot) DataQuoteParams {
    return DataQuoteParams {
        Id: id,
        Premium: premium,
        Quantity: quantity,
        Spot: spot,
    }
}

type DataCounterParty struct {
    Backing mt.MicrotickCoin `json:"backing"`
    Cost mt.MicrotickCoin `json:"premium"`
    FilledQuantity mt.MicrotickQuantity `json:"quantity"`
    FinalFill bool `json:"final"`
    Short mt.MicrotickAccount `json:"short"`
    Quoted DataQuoteParams `json:"quoted"`
}

func (dat DataActiveTrade) CurrentValue(current mt.MicrotickSpot) mt.MicrotickCoin {
    strike := dat.Strike.Amount
    var delta sdk.Dec
    if dat.Type == mt.MicrotickPut {
        // Put
        delta = strike.Sub(current.Amount)
    } else {
        // Call
        delta = current.Amount.Sub(strike)
    }
    if delta.IsNegative() {
        return mt.NewMicrotickCoinFromExtCoinInt(0)
    }
    value := delta.Mul(dat.FilledQuantity.Amount)
    if value.GT(dat.Backing.Amount) {
        value = dat.Backing.Amount
    }
    return mt.NewMicrotickCoinFromDec(value)
}

type CounterPartySettlement struct {
    Settle mt.MicrotickCoin
    Refund mt.MicrotickCoin
    RefundAddress mt.MicrotickAccount
    Backing mt.MicrotickCoin
}

func (dat DataActiveTrade) CounterPartySettlements(current mt.MicrotickSpot) []CounterPartySettlement {
    strike := dat.Strike.Amount
    var delta sdk.Dec
    if dat.Type == mt.MicrotickPut {
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
        settle := mt.NewMicrotickCoinFromDec(delta.Mul(cp.FilledQuantity.Amount))
        if settle.Amount.GT(cp.Backing.Amount) {
            settle.Amount = cp.Backing.Amount
        }
        refund := cp.Backing.Amount.Sub(settle.Amount)
        result = append(result, CounterPartySettlement {
            Settle: settle,
            Refund: mt.NewMicrotickCoinFromDec(refund),
            RefundAddress: cp.Short,
            Backing: cp.Backing,
        })
    }
    return result   
}
