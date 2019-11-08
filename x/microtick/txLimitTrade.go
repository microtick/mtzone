package microtick

import (
    "fmt"
    "time"
    
    "github.com/cosmos/cosmos-sdk/codec"
    sdk "github.com/cosmos/cosmos-sdk/types"
)

type TxLimitTrade struct {
    Market MicrotickMarket
    Duration MicrotickDuration
    Buyer MicrotickAccount
    TradeType MicrotickTradeType
    Limit MicrotickPremium
    MaxCost MicrotickCoin
}

func NewTxLimitTrade(market MicrotickMarket, dur MicrotickDuration, buyer sdk.AccAddress,
    tradeType MicrotickTradeType, limit MicrotickPremium, maxCost MicrotickCoin) TxLimitTrade {
        
    return TxLimitTrade {
        Market: market,
        Duration: dur,
        Buyer: buyer,
        TradeType: tradeType,
        
        Limit: limit,
        MaxCost: maxCost,
    }
}

type LimitTradeData struct {
    Originator string `json:"originator"`
    Trade DataActiveTrade `json:"trade"`
    Consensus MicrotickSpot `json:"consensus"`
    Time time.Time `json:"time"`
}

func (msg TxLimitTrade) Route() string { return "microtick" }

func (msg TxLimitTrade) Type() string { return "trade_limit" }

func (msg TxLimitTrade) ValidateBasic() sdk.Error {
    if len(msg.Market) == 0 {
        return sdk.ErrInternal("Unknown market")
    }
    if msg.Buyer.Empty() {
        return sdk.ErrInvalidAddress(msg.Buyer.String())
    }
    return nil
}

func (msg TxLimitTrade) GetSignBytes() []byte {
    return sdk.MustSortJSON(msgCdc.MustMarshalJSON(msg))
}

func (msg TxLimitTrade) GetSigners() []sdk.AccAddress {
    return []sdk.AccAddress{ msg.Buyer }
}

// Handler

func handleTxLimitTrade(ctx sdk.Context, keeper Keeper, msg TxLimitTrade) sdk.Result {
    params := keeper.GetParams(ctx)
     
    if !keeper.HasDataMarket(ctx, msg.Market) {
        return sdk.ErrInternal("No such market: " + msg.Market).Result()
    }
    
    if !ValidMicrotickDuration(msg.Duration) {
        return sdk.ErrInternal(fmt.Sprintf("Invalid duration: %d", msg.Duration)).Result()
    }
    
    // Step 1 - Obtain the strike spot price and create trade struct
    market, err2 := keeper.GetDataMarket(ctx, msg.Market)
    if err2 != nil {
        return sdk.ErrInternal("Error fetching market").Result()
    }
    commission := NewMicrotickCoinFromDec(params.CommissionTradeFixed)
    settleIncentive := NewMicrotickCoinFromDec(params.SettleIncentive)
    now := ctx.BlockHeader().Time
    trade := NewDataActiveTrade(now, msg.Market, msg.Duration, msg.TradeType,
        msg.Buyer, market.Consensus, commission, settleIncentive)
        
    matcher := NewMatcher(trade, func (id MicrotickId) DataActiveQuote {
        quote, err := keeper.GetActiveQuote(ctx, id)
        if err != nil {
            // This function should always be called with an active quote
            panic("Invalid quote ID")
        }
        return quote
    })
        
    // Step 2 - Compute premium for quantity requested
    market.MatchByLimit(&matcher, msg.Limit, msg.MaxCost)
    
    if matcher.hasQuantity() {
        
        // Step 3 - Deduct premium from buyer account and add it to provider account
        // We do this first because if the funds aren't there we abort
        total := NewMicrotickCoinFromDec(matcher.TotalCost.Add(trade.Commission.Amount).Add(settleIncentive.Amount))
        keeper.WithdrawMicrotickCoin(ctx, msg.Buyer, total)
        //fmt.Printf("Trade Commission: %s\n", trade.Commission.String())
        //fmt.Printf("Settle Incentive: %s\n", settleIncentive.String())
        keeper.PoolCommission(ctx, trade.Commission)
    
        // Step 4 - Finalize trade 
        matcher.Trade.Id = keeper.GetNextActiveTradeId(ctx)
        
        matcher.AssignCounterparties(ctx, keeper, &market)
        
        // Update the account status for the buyer
        accountStatus := keeper.GetAccountStatus(ctx, msg.Buyer)
        accountStatus.ActiveTrades.Insert(NewListItem(matcher.Trade.Id, sdk.NewDec(matcher.Trade.Expiration.UnixNano())))
        accountStatus.SettleBacking = accountStatus.SettleBacking.Add(settleIncentive)
        accountStatus.NumTrades++
        
        // Commit changes
        keeper.SetAccountStatus(ctx, msg.Buyer, accountStatus)
        keeper.SetDataMarket(ctx, market)
        
        matcher.Trade.Balance = keeper.GetTotalBalance(ctx, msg.Buyer)
        keeper.SetActiveTrade(ctx, matcher.Trade)
    
        ctx.EventManager().EmitEvent(
            sdk.NewEvent(
                sdk.EventTypeMessage,
                sdk.NewAttribute("mtm.NewTrade", fmt.Sprintf("%d", matcher.Trade.Id)),
                sdk.NewAttribute(fmt.Sprintf("trade.%d", matcher.Trade.Id), "event.create"),
                sdk.NewAttribute(fmt.Sprintf("acct.%s", msg.Buyer), "trade.long"),
                sdk.NewAttribute("mtm.MarketTick", msg.Market),
            ),
        )
        
        for i := 0; i < len(matcher.FillInfo); i++ {
            thisFill := matcher.FillInfo[i]
            
            ctx.EventManager().EmitEvent(
                sdk.NewEvent(
                    sdk.EventTypeMessage,
                    sdk.NewAttribute(fmt.Sprintf("acct.%s", thisFill.Quote.Provider), "trade.short"),
                ),
            )
            
            quoteKey := fmt.Sprintf("quote.%d", thisFill.Quote.Id)
            if thisFill.FinalFill {
                ctx.EventManager().EmitEvent(
                    sdk.NewEvent(
                        sdk.EventTypeMessage,
                        sdk.NewAttribute(quoteKey, "event.final"),
                    ),
                )
            } else {
                // should never get here, but in case logic changes for filling limit order
                ctx.EventManager().EmitEvent(
                    sdk.NewEvent(
                        sdk.EventTypeMessage,
                        sdk.NewAttribute(quoteKey, "event.match"),
                    ),
                )
            }
        }
        
        // Data
        data := LimitTradeData {
            Originator: "limitTrade",
            Consensus: market.Consensus,
            Time: now,
            Trade: matcher.Trade,
        }
        bz, _ := codec.MarshalJSONIndent(keeper.cdc, data)
            
        return sdk.Result {
            Data: bz,
            Events: ctx.EventManager().Events(),
        }
        
    } else {
        
        // No liquidity available
        return sdk.ErrInternal("No liquidity available").Result()
        
    }
}
