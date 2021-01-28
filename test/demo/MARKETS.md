## Setting up a Microtick market

With the instructions complete from the initial setup from the README, we can now create a market and place
a quote:

1.  Create a governance proposal to add an XBTUSD market

```
$ mtm tx gov submit-proposal microtick-add-markets --proposal ./add-market-proposal.json --home ./data/microtick --from user --keyring-backend test --chain-id microtick --deposit 10000000stake -y
```

2.  Use your validator's stake to vote for the proposal, before 5 minutes passes:

```
$ mtm tx gov vote 2 yes --from validator --home ./data/microtick --keyring-backend test --chain-id microtick -y
```

After the governance period (5 minutes) elapses, you should see a XBTUSD market on-chain:

```
$ mtm query microtick market XBTUSD
```

Assuming you followed the instructions in the README, the "user" account should have 1 DAI as available backing.  So let's create a quote. This will require some data from you.  Look up the current price for XBTUSD: https://coincap.io/assets/bitcoin

* Use the current price as your spot price.  Example: 35000spot
* Use the difference between the daily high / low divided by two to approximate a 12-hour volatility. (Don't use this algorithm in production please!).  Example: 1000premium

3.  Create a quote:

```
$ mtm tx microtick create XBTUSD 12hour 0.5backing 35000spot 1000premium --from user --home ./data/microtick --keyring-backend test --chain-id microtick --gas 1000000 -y
```

You should now see that the XBTUSD market has a consensus spot price:

```
$ mtm query microtick market XBTUSD:
consensus:
  amount: "35000.000000000000000000"
  denom: spot
```

