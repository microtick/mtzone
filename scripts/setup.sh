#!/bin/sh

MTBINARY=mtm
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

export mtm_HOME=$HOME/chains/$1
export mtm_CHAIN_ID=$1
export mtm_KEYRING_BACKEND=test

echo "Working directory: $mtm_HOME"

if [ -d $mtm_HOME ]; then
  echo "Directory exists, exiting"
  exit 1
fi

echo "Initializing: $2"
redirect $MTBINARY init $2

echo "Setting chain params"
GENESIS=$mtm_HOME/config/genesis.json
TRANSFORMS='.app_state.slashing.params.signed_blocks_window="10000"'
TRANSFORMS+='|.app_state.microtick.params.mint_denom="stake"'
TRANSFORMS+='|.app_state.microtick.markets=[{name:"XBTUSD",description:"Crypto - Bitcoin"},{name:"ETHUSD",description:"Crypto - Ethereum"}]'
TRANSFORMS+='|.app_state.microtick.durations=[{name:"5minute",seconds:300},{name:"15minute",seconds:900},{name:"1hour",seconds:3600}]'
jq "$TRANSFORMS" $GENESIS > $mtm_HOME/tmp && mv $mtm_HOME/tmp $GENESIS

redirect $MTBINARY keys add validator
redirect $MTBINARY keys add bank

echo "Adding validator genesis accounts"
redirect $MTBINARY add-genesis-account validator 100000000000stake 
redirect $MTBINARY add-genesis-account bank 1000000000000udai

redirect $MTBINARY gentx validator 100000000000stake
redirect $MTBINARY collect-gentxs
