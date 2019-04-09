# Settling a Trade

For this step, trades can't be settled for the time duration they are active. Go grab a cup of coffee and come back in 5 minutes...!

Once the time period of the trade has elapsed, issue the following command to settle the trade. Either key for the accounts involved will work to settle the trade,
since both accounts were counterparties to it.

## Settle Trade Transaction

```
$ mtcli tx microtick settle 1 --from trader
```

Settling a trade locks in any profits, and refunds the backing that is not used to settle the trade back to the market maker who placed the original quote.

## Query Account

Let's query both accounts and see how they fared:

```
$ mtcli query microtick account $(mtcli keys show marketmaker1 -a)
Account: cosmos1qwu9f6zk5klej0tfs8p40j6uu8j86nh80nm3t4
Balance: 996.609615384615384615fox
Change: 0.609615384615384615fox
NumQuotes: 1
NumTrades: 0
ActiveQuotes: [1]
ActiveTrades: []
QuoteBacking: 3.000000000000000000fox
TradeBacking: 0.000000000000000000fox
```

```
$ mtcli query microtick account $(mtcli keys show marketmaker2 -a)
Account: cosmos18ljxxk48nsx0u8trm0rylgndxrk7t9pvqh0knw
Balance: 990.000000000000000000fox
Change: 0.000000000000000000fox
NumQuotes: 1
NumTrades: 0
ActiveQuotes: [2]
ActiveTrades: []
QuoteBacking: 10.000000000000000000fox
TradeBacking: 0.000000000000000000fox
```

```
$ mtcli query microtick account $(mtcli keys show trader -a)
Account: cosmos1ved24h424mhqa072dclwyjgjewy8cxj0tmsk34
Balance: 1000.390384615384615385fox
Change: 0.390384615384615385fox
NumQuotes: 0
NumTrades: 1
ActiveQuotes: []
ActiveTrades: []
QuoteBacking: 0.000000000000000000fox
TradeBacking: 0.000000000000000000fox
```

Note that the trader is showing a profit of 0.39 tokens from his or her initial 1000fox balance. The first market maker has lost the corresponding amount, once you add back the outstanding quote backing back into the account balance to get 999.6096 tokens.

This is as it should be - it was the first market maker who's quote was "more wrong" as compared to the consensus price and the real world observed price.  The second market maker still has full backing for their original quote (quote #2):

```
$ mtcli query microtick quote 2
Quote Id: 2
Provider: cosmos18ljxxk48nsx0u8trm0rylgndxrk7t9pvqh0knw
Market: BTCUSD
Duration: 5minute
Backing: 10.000000000000000000fox
Spot: 5280.000000000000000000spot
Premium: 10.000000000000000000premium
Quantity: 0.100000000000000000quantity
PremiumAsCall: 13.461538461538461538premium
PremiumAsPut: 6.538461538461538462premium
```

While the first quote is now has less backing because its backing was used as collateral in the trade.

```
$ mtcli query microtick quote 1
Quote Id: 1
Provider: cosmos1qwu9f6zk5klej0tfs8p40j6uu8j86nh80nm3t4
Market: BTCUSD
Duration: 5minute
Backing: 3.000000000000000000fox
Spot: 5250.000000000000000000spot
Premium: 10.000000000000000000premium
Quantity: 0.030000000000000000quantity
PremiumAsCall: 0.000000000000000000premium
PremiumAsPut: 21.538461538461538461premium
```

As a result of the first quote losing its backing to the trade, the consensus moved up closer to the second quote's spot. This change in spot price is what made the trade profitable for the trader.

```
$ mtcli query microtick consensus BTCUSD
Market: BTCUSD
Consensus: 5273.076923076923076923spot
Sum Backing: 13.000000000000000000fox
Sum Weight: 0.130000000000000000quantity
```

