# Placing a Quote

Every quote on the Microtick market has an amount of token backing, a spot price and a premium.

* Token backing = the amount of on-chain tokens staked on this quote.
* Spot price = an assertion of the real world or expected schelling point price for the underlying asset this market represents.
* Premium = an assertion of real world or expected volatility of the above spot price.

In layman's terms, a market maker placing a quote asserts that the real world price is:

1. At the given spot price,
2. Is expected to stay within the range [ spot - premium, spot + premium ] over the time duration of the quote, and
3. Is backed by the given amount of on-chain tokens which are used as collateral if the quote is traded.

Token backing and premium are also used to calculate the quantity of a quote. Specifically,

```
Quantity = Token Backing / Quoted Premium
```

Note that the more certain the market maker is of a quote, the less premium they will need to quote and therefore the quote will 
have more weight in the marketplace (quantity and weight are two terms used interchangeably on Microtick)

## Create Quote Transaction

To create a quote, use the following form:

```
$ mtcli tx microtick create-quote BTCUSD 5minute 10fox 5250spot 10premium --from mykey1
```

The values above used in the command give the market (BTCUSD), the standardized duration (5minute), the token backing (10fox),
the asserted spot or schelling point price (5250spot), and premium or uncertainty over the time period (10premium).

_In layman's terms, this quote says the market maker is asserting BTCUSD is at 5250, +/-10 over the next 5 minutes and is putting
10 fox tokens behind the quote to back it._

## Query Quote

To check the quote status:

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
PremiumAsCall: 10.000000000000000000premium
PremiumAsPut: 10.000000000000000000premium
```

Note that all the values are reflected in the quote status, and in addition we see Quantity, Premium as Call, and Premium as Put values.

* Quantity = Backing / Premium, with a Leverage of 10x that is built in to every quantity calculation (more on this later).
* Premium as Call = the actual price a trader would pay per unit of quantity to place a call trade against this quote's collateral.
* Premium as Put = the actual price a trader would pay per unit of quantity to place a put trade against this quote's collateral.

_Premium as Call and Premium as Put right now are equal to our quoted premium, but this is only because there are no other quotes on the marketplace and so our quote is the only source of consensus on the market._

You can also see our new quote's token backing and weight in the market and orderbook queries:

```
$ mtcli query microtick market BTCUSD
Market: BTCUSD
Consensus: 5250.000000000000000000spot
Orderbooks: [
  5minute:
    Sum Backing: 10.000000000000000000fox
    Sum Weight: 0.100000000000000000quantity 
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
Sum Backing: 10.000000000000000000fox
Sum Weight: 0.100000000000000000quantity
```

```
$ mtcli query microtick orderbook BTCUSD 5minute
Sum Backing: 10.000000000000000000fox
SumWeight: 0.100000000000000000quantity
Calls: [1]
Puts: [1]
```
