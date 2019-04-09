# Placing a Trade

Suppose you're a trader and you're watching the Microtick marketplace. You see the consensus price at $5265, and two quotes
with call prices of 2.5premium and 17.5premium.

Watching your external data feed on the price of Bitcoin, you see a real world price of $5278. Therefore, you decide you want to 
buy a call on the Microtick marketplace, expecting that the price of $5265 is lagging and will go up.

You don't want to risk a lot but the call price of 2.5premium is much less than the difference between 5278 and 5265, so you decide
to buy 0.07quantity of calls on BTCUSD:

## Create Trade Transaction

```
$ mtcli tx microtick trade-market BTCUSD 5minute call 0.07quantity --from trader
```

(note that you must use a different account to buy these calls because otherwise you'd be the counterparty to your own trade and
the marketplace would reject it)

## Query Trade

```
$ mtcli query microtick trade 1
Trade Id: 1
Long: cosmos1ved24h424mhqa072dclwyjgjewy8cxj0tmsk34
Market: BTCUSD
Duration: 5minute
Type: call
Start: 2019-04-09 01:34:48.427146414 +0000 UTC
Expiration: 2019-04-09 01:39:48.427146414 +0000 UTC
Filled Quantity: 0.070000000000000000quantity
Backing: 7.000000000000000000fox
Cost: 0.175000000000000000fox
Commission: 0.000000000000000000fox
Counter Parties: [
    Short: cosmos1qwu9f6zk5klej0tfs8p40j6uu8j86nh80nm3t4
    Backing: 7.000000000000000000fox
    Cost: 0.175000000000000000fox
    FilledQuantity: 0.070000000000000000quantity]
Strike: 5265.000000000000000000spot 
Current Spot: 5273.076923076923076923spot
Current Value: 0.565384615384615385fox
```

There are a bunch of interesting things happening here. First, your order matched the cheapest quote first and was filled against
that quote automatically. Second, because the first quote had a lower quoted price than the second quote, and now has less quantity
(weight in the marketplace), the consensus price has moved up to 5273.0769 (Current Spot). What's more, because the calls were so
cheap, **the current value of the trade (0.565fox) is almost 3 times the price you paid (0.175fox)!!!**

You're making money already and what's more, your bullish outlook on the price based on your real world observation of a higher price
has caused the market to move higher as a result!

You can also see that the trade start and trade expiration were assigned automatically.

Next: [Settle the Trade](https://github.com/mjackson001/mtzone/blob/master/doc/settletrade.md)

