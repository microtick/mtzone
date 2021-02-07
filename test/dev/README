## Running a test / dev local setup

This directory contains the code to start up a local test setup of two chains: gaiad and microtick with all the bank tokens on the gaiad chain initially. You can then fund accounts on the microtick chain using IBC.

1. Start in this directory and make sure the versions of mtm (stargate branch), gaiad and rly (the cosmos relayer) are in your path.

2. You can edit the initial account balances, and settings in the config.tom, app.toml and genesis.json for gaiad and microtick in the corresponding .json files.

3. Run the script "./run".  This will setup the chains running from a local directory "./data"

4. If you need to restart the chains, just re-run "./run". It will stop the existing chains and reconfigure / restart. If you need to stop the chains, run "./stop".

5. To move tokens from gaiad to microtick, run "./fund <micro-account> 1000000000udai" where micro-account is your microtick chain account.

That's it!


## Misc

Command to run grpcui (browser-based GPRC client) from mtzone root directory:

$ grpcui -proto ./proto/microtick/msg/GRPC.proto -import-path ./vendor/github.com/cosmos/cosmos-sdk/third_party/proto -import-path ./proto -import-path ./vendor/github.com/cosmos/cosmos-sdk/proto -plaintext localhost:9090
