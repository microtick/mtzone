#!/bin/sh

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
MTROOT=$WORK mtd init $2 > /dev/null 2>&1

echo "Setting chain id"
GENESIS=$WORK/mtd/config/genesis.json
TRANSFORMS='.chain_id="'$1'"'
TRANSFORMS+='|.app_state.slashing.params.signed_blocks_window="10000"'
TRANSFORMS+='|.app_state.microtick.markets=[{name:"XBTUSD",description:"Crypto - Bitcoin"},{name:"ETHUSD",description:"Crypto - Ethereum"}]'
TRANSFORMS+='|.app_state.microtick.durations=[{name:"5minute",seconds:300},{name:"15minute",seconds:900},{name:"1hour",seconds:3600},{name:"4hour",seconds:14400},{name:"12hour",seconds:43200}]'
jq "$TRANSFORMS" $WORK/mtd/config/genesis.json > $WORK/tmp && mv $WORK/tmp $WORK/mtd/config/genesis.json
sed -i 's/stake/utick/g' $WORK/mtd/config/genesis.json

MTROOT=$WORK mtcli config chain-id $1 > /dev/null 2>&1
if [ -z "$PROD"]; then
  MTROOT=$WORK mtcli config keyring-backend test > /dev/null 2>&1

  # Create password file
  echo "temp1234" > $WORK/pass

  MTROOT=$WORK mtcli keys show validator -a > /dev/null 2>&1
  if [ $? -ne 0 ]; then
    echo "Creating validator key"
    MTROOT=$WORK mtcli keys add validator < $WORK/pass > /dev/null 2>&1
  fi

  MTROOT=$WORK mtcli keys show microtick -a > /dev/null 2>&1
  if [ $? -ne 0 ]; then
    echo "Creating microtick key"
    MTROOT=$WORK mtcli keys add microtick < $WORK/pass > /dev/null 2>&1
  fi

  echo "Adding validator genesis account"
  MTROOT=$WORK mtd add-genesis-account validator 1000000000000utick --keyring-backend=test > /dev/null 2>&1

  echo "Adding microtick genesis account"
  MTROOT=$WORK mtd add-genesis-account microtick 1000000000000udai --keyring-backend=test > /dev/null 2>&1

  echo "Creating genesis transaction"
  MTROOT=$WORK mtd gentx --amount 1000000000000utick --name validator --keyring-backend=test < $WORK/pass > /dev/null 2>&1

  echo "Collecting genesis transactions"
  MTROOT=$WORK mtd collect-gentxs > /dev/null 2>&1

  rm $WORK/pass
fi
