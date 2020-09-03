package keeper

import (
    "fmt"
    "time"
    sdk "github.com/cosmos/cosmos-sdk/types"
    
    mt "github.com/mjackson001/mtzone/x/microtick/types"
)

type DataActiveTrade struct {
    Id mt.MicrotickId `json:"id"`
    Market mt.MicrotickMarket `json:"market"`
    // DurationName is only a tag at this point, not functional
    DurationName mt.MicrotickDurationName `json:"duration"`
    Order mt.MicrotickOrderType `json:"order"`
    Taker mt.MicrotickAccount `json:"taker"`
    Quantity mt.MicrotickQuantity `json:"quantity"`
    Legs []DataTradeLeg `json:"legs"`
    Start time.Time `json:"start"`
    Expiration time.Time `json:"expiration"`
    Strike mt.MicrotickSpot `json:"strike"`
    Commission mt.MicrotickCoin `json:"commission"`
    SettleIncentive mt.MicrotickCoin `json:"settleIncentive"`
}

func NewDataActiveTrade(now time.Time, market mt.MicrotickMarket, 
    dur mt.MicrotickDurationName, durSeconds mt.MicrotickDuration, 
    otype mt.MicrotickOrderType, taker mt.MicrotickAccount, 
    quantity mt.MicrotickQuantity, strike mt.MicrotickSpot,
    commission mt.MicrotickCoin, settleIncentive mt.MicrotickCoin) DataActiveTrade {
        
    expire, err := time.ParseDuration(fmt.Sprintf("%d", durSeconds) + "s")
    if err != nil {
        panic("invalid time")
    }
    return DataActiveTrade {
        Id: 0, // set actual trade ID later after premium has been verified
        Market: market,
        DurationName: dur,
        Order: otype,
        Taker: taker,
        Quantity: quantity,
        Start: now,
        Expiration: now.Add(expire),
        Strike: strike,
        Commission: commission,
        SettleIncentive: settleIncentive,
    }
}

type DataQuotedParams struct {
    Id mt.MicrotickId `json:"id"`
    Final bool `json:"final"`
    Premium mt.MicrotickPremium `json:"premium"`
    Spot mt.MicrotickSpot `json:"spot"`
}

func NewDataQuotedParams(id mt.MicrotickId, final bool, premium mt.MicrotickPremium, spot mt.MicrotickSpot) DataQuotedParams {
    return DataQuotedParams {
        Id: id,
        Final: final,
        Premium: premium,
        Spot: spot,
    }
}

type DataTradeLeg struct {
    LegId mt.MicrotickId `json:"leg_id"`
    Type mt.MicrotickLegType `json:"type"`
    Backing mt.MicrotickCoin `json:"backing"`
    Premium mt.MicrotickPremium `json:"premium"`
    Cost mt.MicrotickCoin `json:"cost"`
    Quantity mt.MicrotickQuantity `json:"quantity"`
    Long mt.MicrotickAccount `json:"long"`
    Short mt.MicrotickAccount `json:"short"`
    Quoted DataQuotedParams `json:"quoted"`
}

func NewDataTradeLeg(legId mt.MicrotickId, 
    ttype mt.MicrotickLegType, 
    backing mt.MicrotickCoin, 
    premium mt.MicrotickPremium,
    cost mt.MicrotickCoin, 
    quantity mt.MicrotickQuantity, 
    long mt.MicrotickAccount, 
    short mt.MicrotickAccount, 
    quoted DataQuotedParams) DataTradeLeg {
        
    return DataTradeLeg {
        LegId: legId,
        Type: ttype,
        Backing: backing,
        Premium: premium,
        Cost: cost,
        Quantity: quantity,
        Long: long,
        Short: short,
        Quoted: quoted,
    }
}

func (dtl *DataTradeLeg) CalculateValue(current sdk.Dec, strike sdk.Dec) sdk.Dec {
    var delta sdk.Dec
    if dtl.Type == mt.MicrotickLegCall {
        delta = current.Sub(strike)
    } else {
        delta = strike.Sub(current)
    }
    if delta.IsNegative() {
        return sdk.ZeroDec()
    }
    value := delta.Mul(dtl.Quantity.Amount)
    if value.GT(dtl.Backing.Amount) {
        value = dtl.Backing.Amount
    }
    return value
}

func (dat DataActiveTrade) CurrentValue(acct mt.MicrotickAccount, current mt.MicrotickSpot) sdk.Dec {
    strike := dat.Strike.Amount
    value := sdk.ZeroDec()
    for _, leg := range dat.Legs {
        if acct.Equals(leg.Long) {
            value = value.Add(leg.CalculateValue(current.Amount, strike))
        }
        if acct.Equals(leg.Short) {
            value = value.Sub(leg.CalculateValue(current.Amount, strike))
        }
    }
    return value
}

type TradeLegSettlement struct {
    LegId mt.MicrotickId
    Settle mt.MicrotickCoin
    Refund mt.MicrotickCoin
    SettleAccount mt.MicrotickAccount
    RefundAccount mt.MicrotickAccount
    Backing mt.MicrotickCoin
}

func (dat DataActiveTrade) CalculateLegSettlements(current mt.MicrotickSpot) []TradeLegSettlement {
    strike := dat.Strike.Amount
    var result []TradeLegSettlement
    for _, leg := range dat.Legs {
        settle := mt.NewMicrotickCoinFromDec(leg.CalculateValue(current.Amount, strike))
        refund := mt.NewMicrotickCoinFromDec(leg.Backing.Amount.Sub(settle.Amount))
        result = append(result, TradeLegSettlement {
            LegId: leg.LegId,
            Settle: settle,
            Refund: refund,
            SettleAccount: leg.Long,
            RefundAccount: leg.Short,
            Backing: leg.Backing,
        })
    }
    return result
}
