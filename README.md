# Microtick STARGATE

## Instructions for building

1. Set GOPATH, GOBIN, and PATH to include GOBIN appropriately.

```
export GOPATH=...
$ export GOBIN=$GOPATH/bin
$ export PATH=$GOBIN:$PATH
```

2. Follow the instructions to install grpc-gateway binaries in your GOBIN:

https://pkg.go.dev/mod/github.com/grpc-ecosystem/grpc-gateway

3. Build the Microtick 'mtm' executable:

```
$ make proto
$ make
```

## Instructions for running

4. Remove any existing data and initialize

```
$ rm -r ~/.microtick
$ ./mtm init localnet --chain-id=microtick_test
```

5. Create a validator with staking tokens (stake) and a test account with test tokens (udai).  1 dai = 1000000 udai.

```
$ ./mtm keys add validator --keyring-backend=test
$ ./mtm keys add myaccount --keyring-backend=test
$ ./mtm add-genesis-account validator 1000000000000stake --keyring-backend=test
$ ./mtm add-genesis-account myaccount 1000000000000udai --keyring-backend=test
$ ./mtm gentx validator --keyring-backend=test --chain-id=microtick_test
$ ./mtm collect-gentxs
```

6. Run your test chain:
```
$ ./mtm start
```
