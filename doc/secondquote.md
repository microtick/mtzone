# Placing a Second Quote

Now it's time to make things a bit more interesting. We're going to place a second quote on the market and see how that changes things.

```
$ mtcli tx microtick create-quote BTCUSD 5minute 10fox 5280spot 10premium --from mykey1
```

Let's query the first quote and this new quote, and check the market as well:

```
$ mtcli query microtick quote 1
Quote Id: 1
Provider: cosmos1hdq4rhfaz33plxh0qk49wr40edp8mhzfr4ckq8
Market: BTCUSD
Duration: 5minute
Backing: 10.000000000000000000fox
Spot: 5250.000000000000000000spot
Premium: 10.000000000000000000premium
Quantity: 0.100000000000000000quantity
PremiumAsCall: 2.500000000000000000premium
PremiumAsPut: 17.500000000000000000premium
```

```
$ mtcli query microtick quote 2
Quote Id: 2
Provider: cosmos1hdq4rhfaz33plxh0qk49wr40edp8mhzfr4ckq8
Market: BTCUSD
Duration: 5minute
Backing: 10.000000000000000000fox
Spot: 5280.000000000000000000spot
Premium: 10.000000000000000000premium
Quantity: 0.100000000000000000quantity
PremiumAsCall: 17.500000000000000000premium
PremiumAsPut: 2.500000000000000000premium
```

```
$ mtcli query microtick market BTCUSD
Market: BTCUSD
Consensus: 5265.000000000000000000spot
Orderbooks: [
  5minute:
    Sum Backing: 20.000000000000000000fox
    Sum Weight: 0.200000000000000000quantity 
  15minute:
    Sum Backing: 0.000000000000000000fox
    Sum Weight: 0.000000000000000000quantity 
  1hour:
    Sum Backing: 0.000000000000000000fox
    Sum Weight: 0.000000000000000000quantity 
  4hour:
    Sum Backing: 0.000000000000000000fox
    Sum Weight: 0.000000000000000000quantity 
  12hour:
    Sum Backing: 0.000000000000000000fox
    Sum Weight: 0.000000000000000000quantity]
Sum Backing: 20.000000000000000000fox
Sum Weight: 0.200000000000000000quantity
```

Everything looks pretty much the same, and we can recognize our quoted values for Market, Duration, Backing, Spot and Premium.

But wait! Premium as Call and Premium as Put have changed for the first quote, and for the second quote they do not match the quoted premium!
Note that the first quote is much cheaper to buy as a call than the second, and the second quote is much cheaper as a put than the first.

What's happening here is the marketplace is doing several things at once:
1. It calculates a weighted average of quote 1 and quote 2 based on their spot and quantity values. This price is called the **Consensus Price**. In this case, both quotes have 
the same quantities so the consensus is an average of the two (5265spot).
2. The premium of the first quote is discounted as a call, but increased as a put, because the first quote's spot (5250spot) is less than the consensus price.
3. The premium of the second quote is increased as a call, but decreased as a put, because the second quote's spot (5280spot) is greater than the consensus price.

Next: [Place a Trade](https://github.com/mjackson001/mtzone/blob/master/doc/placetrade.md)
