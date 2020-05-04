#!/bin/sh

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
echo "Creating testnet"
MTROOT=$WORK mtd init testnet > /dev/null 2>&1

echo "Setting chain id"
GENESIS=$WORK/mtd/config/genesis.json
jq '.chain_id="mtlocal"' $WORK/mtd/config/genesis.json > $WORK/tmp && mv $WORK/tmp $WORK/mtd/config/genesis.json
MTROOT=$WORK mtcli config chain-id mtlocal > /dev/null 2>&1
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
MTROOT=$WORK mtd add-genesis-account validator 1000000000000stake --keyring-backend=test > /dev/null 2>&1

echo "Adding microtick genesis account"
MTROOT=$WORK mtd add-genesis-account microtick 1000000000000udai --keyring-backend=test > /dev/null 2>&1

echo "Creating genesis transaction"
MTROOT=$WORK mtd gentx --name validator --keyring-backend=test < $WORK/pass > /dev/null 2>&1

echo "Collecting genesis transactions"
MTROOT=$WORK mtd collect-gentxs > /dev/null 2>&1

rm $WORK/pass


