# Command line usage

Once the software is built and the blockchain is running, you can start to issue commands to better understand the operation of 
the Microtick marketplace / BFT price feed.

Microtick is a decentralized byzantine fault tolerant marketplace for realtime price discovery that:

* Does not require liquidity or trading of the underlying asset
* Achieves an on-chain realtime global consensus price for the underlying asset that is usable by an Dapp for free
* Creates market-based incentives that reward participants that converge on a [Schelling Point](https://en.wikipedia.org/wiki/Focal_point_(game_theory)) and punishes those that diverge from it

## Step 1. Create a Market

You can name a market whatever you want, but it makes sense to name it something that:

* Describes the asset or asset class you are pricing
* Makes it easy for other market participants to find

```
mtcli tx microtick create-market BTCUSD --from mykey1
```

Once created, you can verify the market exists:

```
$ mtcli query microtick market BTCUSD
Market: BTCUSD
Consensus: 0.000000000000000000spot
Orderbooks: [
  5minute:
    Sum Backing: 0.000000000000000000fox
    Sum Weight: 0.000000000000000000quantity 
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
Sum Backing: 0.000000000000000000fox
Sum Weight: 0.000000000000000000quantity
```

All values in Microtick are labeled according to the value they represent. The labels following the numbers above represent:

* **fox** = the underlying token that is used for backing quotes, placing trades, and staking
* **quantity** = a derived measure of units of trading that will be described in more detail later

Every market in Microtick is standardized into different time durations for the quotes and trades running on the marketplace.
These standardized durations concentrate market activity to facilitate trading.

* **5minute**
* **15minute**
* **1hour**
* **4hour**
* **12hour**

You can see more detail on a particular orderbook using the following command:

```
$ mtcli query microtick orderbook ETHUSD 5minute
Sum Backing: 0.000000000000000000fox
SumWeight: 0.000000000000000000quantity
Calls: []
Puts: []
```
