#!/bin/sh

MTBINARY=./mtm
TESTOPTS=--keyring-backend=test
SILENT=1

redirect() {
  if [ "$SILENT" -eq 1 ]; then
    "$@" > /dev/null 2>&1
  else
    "$@"
  fi
}

if [[ "$#" -ne 2 ]]; then
  BASE=$(basename $0)
  echo "Usage: $BASE: <chain_id> <moniker>"
  exit 1
fi

if [ -z "$MTROOT" ]; then
  WORK=$HOME/.microtick
else
  WORK=$MTROOT
fi

echo "Working directory: $WORK"

if [ -d $WORK ]; then
  echo "Directory exists, exiting"
  exit 1
fi

mkdir -p $WORK
echo "Initializing: $2"
MTROOT=$WORK redirect $MTBINARY init $2

echo "Setting chain id"
GENESIS=$WORK/config/genesis.json
TRANSFORMS='.chain_id="'$1'"'
TRANSFORMS+='|.app_state.slashing.params.signed_blocks_window="10000"'
TRANSFORMS+='|.app_state.microtick.params.mint_denom="stake"'
TRANSFORMS+='|.app_state.microtick.markets=[{name:"XBTUSD",description:"Crypto - Bitcoin"},{name:"ETHUSD",description:"Crypto - Ethereum"}]'
TRANSFORMS+='|.app_state.microtick.durations=[{name:"5minute",seconds:300},{name:"15minute",seconds:900},{name:"1hour",seconds:3600},{name:"4hour",seconds:14400},{name:"12hour",seconds:43200}]'
jq "$TRANSFORMS" $GENESIS > $WORK/tmp && mv $WORK/tmp $GENESIS

if [ -z "$PROD"]; then
  echo "Adding validator genesis accounts"
  MTROOT=$WORK redirect $MTBINARY add-genesis-account micro1fvx2wxvalg7pwdh8lgnsm20t7ntk8u9lg2w6sw 10000000000stake $TESTOPTS 
  MTROOT=$WORK redirect $MTBINARY add-genesis-account micro195v8l6xaf62m3h3lsdhe78q4ngz0pv5f88y9jk 10000000000stake $TESTOPTS 
  MTROOT=$WORK redirect $MTBINARY add-genesis-account micro12sgxpwgn0fvyagx5j0duazyvj6hfc65ezy09ke 10000000000stake $TESTOPTS 
  MTROOT=$WORK redirect $MTBINARY add-genesis-account micro19qlaxcg5mtgxfgc6l3s7h8cp42c4l92ahzfypq 10000000000stake $TESTOPTS 

  echo "Adding supply genesis account"
  MTROOT=$WORK redirect $MTBINARY add-genesis-account micro1tp56xndmrgwnqspmsl6qqul073fcmxsv6672v3 1000000000000udai,1000000000000stake $TESTOPTS
fi
