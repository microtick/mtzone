# Settling a Trade

Trades can't be settled for the time duration they are active. Go grab a cup of coffee and come back in 5 minutes...

Once the time period of the trade has elapsed, issue the following command to settle the trade. Either key will work to settle the trade,
since both accounts were counterparties to it.

## Settle Trade Transaction

```
$ mtcli tx microtick settle 1 --from mykey1
```

Settling a trade locks in any profits, and refunds the backing that is not used to settle the trade, back to the market maker who 
placed the original quote.

## Query Account

Let's query both accounts and see how they fared:

```
$ mtcli query microtick account $(mtcli keys show mykey1 -a)
Account: cosmos1hdq4rhfaz33plxh0qk49wr40edp8mhzfr4ckq8
Balance: 986.609615384615384615fox
Change: 0.609615384615384615fox
NumQuotes: 2
NumTrades: 0
ActiveQuotes: [1 2]
ActiveTrades: []
QuoteBacking: 13.000000000000000000fox
TradeBacking: 0.000000000000000000fox
```

```
$ mtcli query microtick account $(mtcli keys show mykey2 -a)
Account: cosmos179jq24nm45fku4n0mpqqe29f8etnp3zlhtc2tp
Balance: 1000.390384615384615385fox
Change: 0.390384615384615385fox
NumQuotes: 0
NumTrades: 1
ActiveQuotes: []
ActiveTrades: []
QuoteBacking: 0.000000000000000000fox
TradeBacking: 0.000000000000000000fox
```

Note that the trader (mykey2) is showing a profit of 0.39 tokens from his or her initial 1000fox balance. The market maker has 
lost the corresponding amount, once you add back the outstanding quote backing back into the account balance to get 999.6096 tokens.

