# State

## Notes

```
QuoteID = integer
TradeID = integer
All account balances and quantities are rounded (not floats)
```

## Description

* accounts
  * activeQuotes - array of currently active Quote IDs
  * activeTrades - array of currently active Trade IDs
  * balance (maintained by auth module)
  * id (Account ID maintained by auth module)
  * numQuotes - count of total quotes this account has created
  * numTrades - count of total trades this account has initiated
  * quoteBacking - total amount of tokens currently backing quotes for this account
  * tradeBacking - total amount of tokens currently backing trades for this account
* activeQuotes - map[Quote ID]
  * backing - amount of tokens backing this quote
  * canModify - timestamp when this quote may be modified again
  * commission - commission paid when placing this quote
  * dur - quote duration
  * id - self referencing Quote ID
  * market - quote market
  * modified - timestamp when this quote was last modified
  * premium - quote premium
  * provider - account ID of account that placed this quote
  * quantity - backing / premium, rounded
  * spot - quoted spot price
* activeTrades - map[Trade ID]
  * commission - commission paid when placing this trade
  * counterparties - array
    * backing - amount of backing this counterparty contributed to trade
    * final - boolean: true if this trade used all the available quantity from this quote
    * premium - premium paid to this counterparty
    * quoteParams - this struct contains the quoted values at the time the trade was initiated
      * premium - quoted premium at the time of trade initiation
      * quantity - quoted quantity at the time of trade initiation
      * spot - quoted spot at the time of trade initiation
    * quantity - amount of quantity this counterparty is short
    * quoteId - counterparty Quote ID
    * short - counterparty Account ID
    * startBalance - counterparty account balance after trade was placed (this might be just for debug)
  * dur - trade duration
  * expiration - timestamp of trade expiration
  * id - self referencing Trade ID
  * long - Account ID of account placing the trade
  * market - Market ID of market this trade is in reference to
  * premium - total trade premium
  * quantity - total trade quantity
  * start - timestamp of trade start time
  * startBalance - long account balance after trade was placed
  * strike - trade strike price = market consensus price at time trade was initiated
  * type - call / put
* totalCommission - total commissions paid
* config - adjustable settings
  * commissionQuote - default 0.0005
  * commissionTrade - default 0.05
  * commissionUpdate - default 0.00005
  * settleTime - default 30 seconds
* markets - map[marketId]
  * consensus - current consensus price for market
  * orderbooks - array of all available durations
    * calls - ordered array of quote IDs
    * puts - ordered array of quote IDs
    * sumWeight - sum of quantities of all quotes currently active
  * sumBacking - sum of backing for this market
  * sumSpots - sum of spot prices * quantity of all quotes. used for calculating weighted average
  * sumWeight - sum of quote quantity of all quotes
* nextQuoteId - counter for assigning a unique quote ID
* nextTradeId - counter for assigning a unique trade ID
* tradeExpirations - array of trade IDs that can currently be settled

## Sample state dump from current Tendermint javascript-based Dapp
```
[
  {
    "accounts": {
      "0x39b7cd816c09f25f5928e5158cc399e9d7fdb723": {
        "activeQuotes": [],
        "activeTrades": {
          "long": [],
          "short": []
        },
        "balance": 1016.087073,
        "id": "0x39b7cd816c09f25f5928e5158cc399e9d7fdb723",
        "numQuotes": 0,
        "numTrades": 1,
        "quoteBacking": 0,
        "tradeBacking": 0
      },
      "0x46850d6ca14cae1e7ef91d5a26382446abf3acc2": {
        "activeQuotes": [],
        "activeTrades": {
          "long": [],
          "short": []
        },
        "balance": 984.851519,
        "id": "0x46850d6ca14cae1e7ef91d5a26382446abf3acc2",
        "numQuotes": 0,
        "numTrades": 1,
        "quoteBacking": 0,
        "tradeBacking": 0
      },
      "0x4eb022aea7ea8acf636f644faf652970adee958e": {
        "activeQuotes": [
          1,
          2,
          3,
          4,
          5,
          6,
          7,
          8,
          9,
          11,
          12,
          13,
          14,
          15,
          16
        ],
        "activeTrades": {
          "long": [],
          "short": [
            89
          ]
        },
        "balance": 86619.45673,
        "id": "0x4eb022aea7ea8acf636f644faf652970adee958e",
        "numQuotes": 16,
        "numTrades": 0,
        "quoteBacking": 7500,
        "tradeBacking": 236.842105
      },
      "0x4f0ab595c2f8249c63a8008c0b954e75e7131a31": {
        "activeQuotes": [],
        "activeTrades": {
          "long": [],
          "short": []
        },
        "balance": 848.831259,
        "id": "0x4f0ab595c2f8249c63a8008c0b954e75e7131a31",
        "numQuotes": 0,
        "numTrades": 30,
        "quoteBacking": 0,
        "tradeBacking": 0
      },
      "0x52deb381d241df28b93f71ce0a2cc78d143fa880": {
        "activeQuotes": [],
        "activeTrades": {
          "long": [],
          "short": []
        },
        "balance": 1000,
        "id": "0x52deb381d241df28b93f71ce0a2cc78d143fa880",
        "numQuotes": 0,
        "numTrades": 0,
        "quoteBacking": 0,
        "tradeBacking": 0
      },
      "0x57fb077caf7952ee1944d14b21ab2125414d0a96": {
        "activeQuotes": [],
        "activeTrades": {
          "long": [],
          "short": []
        },
        "balance": 1042.148356,
        "id": "0x57fb077caf7952ee1944d14b21ab2125414d0a96",
        "numQuotes": 0,
        "numTrades": 4,
        "quoteBacking": 0,
        "tradeBacking": 0
      },
      "0x5d7b3da095a2df6d4aa6e9710682e0698408fd84": {
        "activeQuotes": [],
        "activeTrades": {
          "long": [
            89
          ],
          "short": []
        },
        "balance": 975.028426,
        "id": "0x5d7b3da095a2df6d4aa6e9710682e0698408fd84",
        "numQuotes": 0,
        "numTrades": 1,
        "quoteBacking": 0,
        "tradeBacking": 0
      },
      "0x63916a1d1fecf8b190e2fb0a424bf7f69aa852a5": {
        "activeQuotes": [],
        "activeTrades": {
          "long": [],
          "short": []
        },
        "balance": 873.635604,
        "id": "0x63916a1d1fecf8b190e2fb0a424bf7f69aa852a5",
        "numQuotes": 0,
        "numTrades": 12,
        "quoteBacking": 0,
        "tradeBacking": 0
      },
      "0x687a1495c508e7ce4c0b1388785ebb123edf6702": {
        "activeQuotes": [],
        "activeTrades": {
          "long": [],
          "short": []
        },
        "balance": 924.484019,
        "id": "0x687a1495c508e7ce4c0b1388785ebb123edf6702",
        "numQuotes": 0,
        "numTrades": 4,
        "quoteBacking": 0,
        "tradeBacking": 0
      },
      "0x7716c6990dec6f9592d9f23fd0276f63f7da0fb5": {
        "activeQuotes": [],
        "activeTrades": {
          "long": [],
          "short": []
        },
        "balance": 1118.564289,
        "id": "0x7716c6990dec6f9592d9f23fd0276f63f7da0fb5",
        "numQuotes": 0,
        "numTrades": 3,
        "quoteBacking": 0,
        "tradeBacking": 0
      },
      "0x7dab3bd166fe25e9c4ee1e66c73bb5d4842f71f2": {
        "activeQuotes": [],
        "activeTrades": {
          "long": [],
          "short": []
        },
        "balance": 645.004635,
        "id": "0x7dab3bd166fe25e9c4ee1e66c73bb5d4842f71f2",
        "numQuotes": 0,
        "numTrades": 17,
        "quoteBacking": 0,
        "tradeBacking": 0
      },
      "0x81f9d8584f82798b65e246a9c8a0fe9212b46f2d": {
        "activeQuotes": [],
        "activeTrades": {
          "long": [],
          "short": []
        },
        "balance": 1000,
        "id": "0x81f9d8584f82798b65e246a9c8a0fe9212b46f2d",
        "numQuotes": 0,
        "numTrades": 0,
        "quoteBacking": 0,
        "tradeBacking": 0
      },
      "0x8bea02d74e0df99d7262787e80a0e5b4b8e33cfb": {
        "activeQuotes": [],
        "activeTrades": {
          "long": [],
          "short": []
        },
        "balance": 976.711267,
        "id": "0x8bea02d74e0df99d7262787e80a0e5b4b8e33cfb",
        "numQuotes": 0,
        "numTrades": 1,
        "quoteBacking": 0,
        "tradeBacking": 0
      },
      "0xaccae102e6fbd2b199e19e1e808a05c381f03846": {
        "activeQuotes": [],
        "activeTrades": {
          "long": [],
          "short": []
        },
        "balance": 948.64345,
        "id": "0xaccae102e6fbd2b199e19e1e808a05c381f03846",
        "numQuotes": 0,
        "numTrades": 2,
        "quoteBacking": 0,
        "tradeBacking": 0
      },
      "0xb3c339b95cd10c614a3f978ff87dfd2c74b654f3": {
        "activeQuotes": [],
        "activeTrades": {
          "long": [],
          "short": []
        },
        "balance": 1066.258065,
        "id": "0xb3c339b95cd10c614a3f978ff87dfd2c74b654f3",
        "numQuotes": 0,
        "numTrades": 11,
        "quoteBacking": 0,
        "tradeBacking": 0
      },
      "0xcc6565ce44c43dc1a71fc1abad8dceccf81dc057": {
        "activeQuotes": [],
        "activeTrades": {
          "long": [],
          "short": []
        },
        "balance": 992.269316,
        "id": "0xcc6565ce44c43dc1a71fc1abad8dceccf81dc057",
        "numQuotes": 0,
        "numTrades": 1,
        "quoteBacking": 0,
        "tradeBacking": 0
      },
      "0xcca77960f040d6b5d9739e6193944d12353bc49a": {
        "activeQuotes": [],
        "activeTrades": {
          "long": [],
          "short": []
        },
        "balance": 1000,
        "id": "0xcca77960f040d6b5d9739e6193944d12353bc49a",
        "numQuotes": 0,
        "numTrades": 0,
        "quoteBacking": 0,
        "tradeBacking": 0
      },
      "0xd0055dd466b6d57aca65434572292d20df59faa1": {
        "activeQuotes": [],
        "activeTrades": {
          "long": [],
          "short": []
        },
        "balance": 1000,
        "id": "0xd0055dd466b6d57aca65434572292d20df59faa1",
        "numQuotes": 0,
        "numTrades": 0,
        "quoteBacking": 0,
        "tradeBacking": 0
      },
      "0xd7b3a20c601de8ae9e3adc0b79ac719c59ae2ce7": {
        "activeQuotes": [],
        "activeTrades": {
          "long": [],
          "short": []
        },
        "balance": 992.819734,
        "id": "0xd7b3a20c601de8ae9e3adc0b79ac719c59ae2ce7",
        "numQuotes": 0,
        "numTrades": 1,
        "quoteBacking": 0,
        "tradeBacking": 0
      },
      "0xeb9a501e1d17b80c5b88a4f3c1fc884b42227112": {
        "activeQuotes": [],
        "activeTrades": {
          "long": [],
          "short": []
        },
        "balance": 1000000000,
        "id": "0xeb9a501e1d17b80c5b88a4f3c1fc884b42227112",
        "numQuotes": 0,
        "numTrades": 0,
        "quoteBacking": 0,
        "tradeBacking": 0
      }
    },
    "activeQuotes": {
      "1": {
        "backing": 500,
        "canModify": 1553115746,
        "commission": 0.25,
        "dur": 300,
        "id": 1,
        "market": "ETHUSD",
        "modified": 1553115716,
        "premium": 0.102605,
        "provider": "0x4eb022aea7ea8acf636f644faf652970adee958e",
        "quantity": 487.305687,
        "spot": 137.67
      },
      "2": {
        "backing": 500,
        "canModify": 1553115746,
        "commission": 0.25,
        "dur": 900,
        "id": 2,
        "market": "ETHUSD",
        "modified": 1553115716,
        "premium": 0.102605,
        "provider": "0x4eb022aea7ea8acf636f644faf652970adee958e",
        "quantity": 487.305687,
        "spot": 137.67
      },
      "3": {
        "backing": 500,
        "canModify": 1553114004,
        "commission": 0.25,
        "dur": 3600,
        "id": 3,
        "market": "ETHUSD",
        "modified": 1553113974,
        "premium": 0.127897,
        "provider": "0x4eb022aea7ea8acf636f644faf652970adee958e",
        "quantity": 390.939584,
        "spot": 137.74
      },
      "4": {
        "backing": 500,
        "canModify": 1553114452,
        "commission": 0.25,
        "dur": 43200,
        "id": 4,
        "market": "ETHUSD",
        "modified": 1553114422,
        "premium": 0.439313,
        "provider": "0x4eb022aea7ea8acf636f644faf652970adee958e",
        "quantity": 113.814069,
        "spot": 137.75
      },
      "5": {
        "backing": 500,
        "canModify": 1553114870,
        "commission": 0.25,
        "dur": 14400,
        "id": 5,
        "market": "ETHUSD",
        "modified": 1553114840,
        "premium": 0.250476,
        "provider": "0x4eb022aea7ea8acf636f644faf652970adee958e",
        "quantity": 199.619924,
        "spot": 137.76
      },
      "6": {
        "backing": 500,
        "canModify": 1553115721,
        "commission": 0.25,
        "dur": 300,
        "id": 6,
        "market": "XBTUSD",
        "modified": 1553115691,
        "premium": 0.920343,
        "provider": "0x4eb022aea7ea8acf636f644faf652970adee958e",
        "quantity": 54.327571,
        "spot": 4008.3
      },
      "7": {
        "backing": 500,
        "canModify": 1553115721,
        "commission": 0.25,
        "dur": 900,
        "id": 7,
        "market": "XBTUSD",
        "modified": 1553115691,
        "premium": 1.594077,
        "provider": "0x4eb022aea7ea8acf636f644faf652970adee958e",
        "quantity": 31.366113,
        "spot": 4008.3
      },
      "8": {
        "backing": 500,
        "canModify": 1553115147,
        "commission": 0.25,
        "dur": 3600,
        "id": 8,
        "market": "XBTUSD",
        "modified": 1553115117,
        "premium": 3.237424,
        "provider": "0x4eb022aea7ea8acf636f644faf652970adee958e",
        "quantity": 15.444378,
        "spot": 4008.2
      },
      "9": {
        "backing": 500,
        "canModify": 1553114724,
        "commission": 0.25,
        "dur": 14400,
        "id": 9,
        "market": "XBTUSD",
        "modified": 1553114694,
        "premium": 6.254526,
        "provider": "0x4eb022aea7ea8acf636f644faf652970adee958e",
        "quantity": 7.994211,
        "spot": 4010.2
      },
      "11": {
        "backing": 500,
        "canModify": 1553115479,
        "commission": 0.25,
        "dur": 300,
        "id": 11,
        "market": "LTCUSD",
        "modified": 1553115449,
        "premium": 0.017854,
        "provider": "0x4eb022aea7ea8acf636f644faf652970adee958e",
        "quantity": 2800.492887,
        "spot": 59.83
      },
      "12": {
        "backing": 500,
        "canModify": 1553115011,
        "commission": 0.25,
        "dur": 900,
        "id": 12,
        "market": "LTCUSD",
        "modified": 1553114981,
        "premium": 0.032791,
        "provider": "0x4eb022aea7ea8acf636f644faf652970adee958e",
        "quantity": 1524.808637,
        "spot": 59.8
      },
      "13": {
        "backing": 500,
        "canModify": 1553113500,
        "commission": 0.25,
        "dur": 3600,
        "id": 13,
        "market": "LTCUSD",
        "modified": 1553113470,
        "premium": 0.065541,
        "provider": "0x4eb022aea7ea8acf636f644faf652970adee958e",
        "quantity": 762.88125,
        "spot": 59.83
      },
      "14": {
        "backing": 500,
        "canModify": 1553112232,
        "commission": 0.25,
        "dur": 14400,
        "id": 14,
        "market": "LTCUSD",
        "modified": 1553112202,
        "premium": 0.128056,
        "provider": "0x4eb022aea7ea8acf636f644faf652970adee958e",
        "quantity": 390.454176,
        "spot": 59.79
      },
      "15": {
        "backing": 500,
        "canModify": 1553112050,
        "commission": 0.25,
        "dur": 43200,
        "id": 15,
        "market": "LTCUSD",
        "modified": 1553112020,
        "premium": 0.223054,
        "provider": "0x4eb022aea7ea8acf636f644faf652970adee958e",
        "quantity": 224.160966,
        "spot": 59.81
      },
      "16": {
        "backing": 500,
        "canModify": 1553114090,
        "commission": 0.25,
        "dur": 43200,
        "id": 16,
        "market": "XBTUSD",
        "modified": 1553114060,
        "premium": 10.277871,
        "provider": "0x4eb022aea7ea8acf636f644faf652970adee958e",
        "quantity": 4.864821,
        "spot": 4016.2
      }
    },
    "activeTrades": {
      "89": {
        "commission": 0.05,
        "counterparties": [
          {
            "backing": 236.842105,
            "final": false,
            "premium": 24.921574,
            "qparams": {
              "premium": 0.14036,
              "quantity": 356.226845,
              "spot": 137.58
            },
            "quantity": 168.739032,
            "quoteId": 3,
            "short": "0x4eb022aea7ea8acf636f644faf652970adee958e",
            "startBalance": 86860.967256
          }
        ],
        "dur": 3600,
        "expiration": 1553116074,
        "id": 89,
        "long": "0x5d7b3da095a2df6d4aa6e9710682e0698408fd84",
        "market": "ETHUSD",
        "premium": 24.921574,
        "quantity": 168.739032,
        "start": 1553112474,
        "startBalance": 975.028426,
        "strike": 137.565334,
        "type": 0
      }
    },
    "ca": 6238.364153,
    "config": {
      "commissionQuote": 0.0005,
      "commissionTrade": 0.05,
      "commissionUpdate": 0.00005,
      "settleTime": 30
    },
    "markets": {
      "ETHUSD": {
        "consensus": 137.702423,
        "orderbooks": {
          "300": {
            "calls": [
              1
            ],
            "puts": [
              1
            ],
            "sumWeight": 487.305687
          },
          "900": {
            "calls": [
              2
            ],
            "puts": [
              2
            ],
            "sumWeight": 487.305687
          },
          "3600": {
            "calls": [
              3
            ],
            "puts": [
              3
            ],
            "sumWeight": 390.939584
          },
          "14400": {
            "calls": [
              5
            ],
            "puts": [
              5
            ],
            "sumWeight": 199.619924
          },
          "43200": {
            "calls": [
              4
            ],
            "puts": [
              4
            ],
            "sumWeight": 113.814069
          }
        },
        "sumBacking": 2500,
        "sumSpots": 231200.296529,
        "sumWeight": 1678.984951
      },
      "LTCUSD": {
        "consensus": 59.818454,
        "orderbooks": {
          "300": {
            "calls": [
              11
            ],
            "puts": [
              11
            ],
            "sumWeight": 2800.492887
          },
          "900": {
            "calls": [
              12
            ],
            "puts": [
              12
            ],
            "sumWeight": 1524.808637
          },
          "3600": {
            "calls": [
              13
            ],
            "puts": [
              13
            ],
            "sumWeight": 762.88125
          },
          "14400": {
            "calls": [
              14
            ],
            "puts": [
              14
            ],
            "sumWeight": 390.454176
          },
          "43200": {
            "calls": [
              15
            ],
            "puts": [
              15
            ],
            "sumWeight": 224.160966
          }
        },
        "sumBacking": 2500,
        "sumSpots": 341132.554976,
        "sumWeight": 5702.797916
      },
      "XBTUSD": {
        "consensus": 4008.756866,
        "orderbooks": {
          "300": {
            "calls": [
              6
            ],
            "puts": [
              6
            ],
            "sumWeight": 54.327571
          },
          "900": {
            "calls": [
              7
            ],
            "puts": [
              7
            ],
            "sumWeight": 31.366113
          },
          "3600": {
            "calls": [
              8
            ],
            "puts": [
              8
            ],
            "sumWeight": 15.444378
          },
          "14400": {
            "calls": [
              9
            ],
            "puts": [
              9
            ],
            "sumWeight": 7.994211
          },
          "43200": {
            "calls": [
              16
            ],
            "puts": [
              16
            ],
            "sumWeight": 4.864821
          }
        },
        "sumBacking": 2500,
        "sumSpots": 456986.63333,
        "sumWeight": 113.997094
      }
    },
    "nextQuoteId": 17,
    "nextTradeId": 90,
    "tradeExpirations": [
      {
        "expiration": 1553116074,
        "id": 89
      }
    ]
  }
]
```
