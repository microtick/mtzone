To demo IBC token backing for Microtick, complete the following steps:

1.  Ensure gaiad and rly are built and in your path. There is a bug fixed relayer version here: https://github.com/mjackson001/relayer

2.  Run ./two-chainz

3.  Set up the funding path from the gaiad chain to the microtick chain

```
$ rly tx link funding -d -o 3s
```

4.  Query the token balances:

```
$ rly q bal gaiad
$ rly q bal microtick
```

5.  Transfer some backing over:

```
$ rly tx xfer gaiad microtick 1000000udai $(rly chains address microtick)
```

6.  Relay the IBC packet:

```
$ rly tx relay funding -d
```

7.  Query the token balances again:

```
$ rly q bal gaiad
$ rly q bal microtick
```

The token balance for microtick is in denomination: ibc/BC599B88586F8C22E408569D7F6FAD40AEBF808A67D2051B86958CBB5F0A16B0.  This does not show up as token backing yet:

```
$ mtm query microtick account $(rly chains address microtick)
balances:
- amount: "0.000000000000000000"
  denom: backing
```

So let's create a governance proposal and swith the token backing to the ibc denomination:

```
$ ./change-backing-proposal
```

And before 5 minutes are finished, vote "yes" for the new proposal:

```
$ mtm tx gov vote 1 yes --from validator --home ./data/microtick --keyring-backend test --chain-id microtick
```

Finally, wait 5 minutes for the proposal to pass and re-query the microtick account to see your funds from the gaiad chain available as token backing and ready to trade on Microtick!

```
$ mtm query microtick account $(rly chains address microtick)
balances:
- amount: "1.000000000000000000"
  denom: backing
```
