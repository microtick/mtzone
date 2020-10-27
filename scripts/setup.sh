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

echo "Setting chain id: $1"
GENESIS=$WORK/config/genesis.json
TRANSFORMS='.chain_id="'$1'"'
TRANSFORMS+='|.app_state.slashing.params.signed_blocks_window="10000"'
TRANSFORMS+='|.app_state.microtick.markets=[{name:"XBTUSD",description:"Crypto - Bitcoin"},{name:"ETHUSD",description:"Crypto - Ethereum"}]'
TRANSFORMS+='|.app_state.microtick.durations=[{name:"5minute",seconds:300},{name:"15minute",seconds:900},{name:"1hour",seconds:3600},{name:"4hour",seconds:14400},{name:"12hour",seconds:43200}]'
jq "$TRANSFORMS" $GENESIS > $WORK/tmp && mv $WORK/tmp $GENESIS
sed -i 's/stake/utick/g' $GENESIS

if [ -z "$PROD"]; then
  echo "Checking validator key"
  MTROOT=$WORK $MTBINARY keys show validator -a $TESTOPTS > /dev/null 2>&1
  if [ $? -ne 0 ]; then
    echo "Creating validator key"
    MTROOT=$WORK redirect $MTBINARY keys add validator $TESTOPTS
  fi

  echo "Checking microtick key"
  MTROOT=$WORK $MTBINARY keys show microtick -a $TESTOPTS > /dev/null 2>&1
  if [ $? -ne 0 ]; then
    echo "Creating microtick key"
    MTROOT=$WORK redirect $MTBINARY keys add microtick $TESTOPTS
  fi

  VALIDATOR=$($MTBINARY keys show validator -a --home=$WORK --keyring-backend=test)
  echo "Adding validator genesis account: $VALIDATOR"
  MTROOT=$WORK redirect $MTBINARY add-genesis-account $VALIDATOR 1000000000000utick

  MICROTICK=$($MTBINARY keys show microtick -a --home=$WORK --keyring-backend=test)
  echo "Adding microtick genesis account: $MICROTICK"
  MTROOT=$WORK redirect $MTBINARY add-genesis-account $MICROTICK 1000000000000udai

  echo "Creating genesis transaction"
  MTROOT=$WORK redirect $MTBINARY gentx validator --amount 1000000000000utick --chain-id $1 $TESTOPTS

  echo "Collecting genesis transactions"
  MTROOT=$WORK redirect $MTBINARY collect-gentxs
fi
