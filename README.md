# Microtick STARGATE

## Instructions for building

1. Ensure the go compiler version 1.15 or greater is installed on your system

```
$ go version
go version go1.16.3 linux/amd64
```

2. Build the Microtick 'mtm' executable:

```
$ make
```

## Instructions for running

4. Set up a directory for chain data and initialize

```
$ export mtm_HOME=<chain directory>
$ export mtm_CHAIN_ID=microtick_test
$ export mtm_KEYRING_BACKEND=test
$ rm -r <chain directory>
$ ./mtm init mylocalnode
```

5. Create a validator with staking tokens (stake) and a test account with test tokens (udai).  1 dai = 1000000 udai.

```
$ ./mtm keys add validator 
$ ./mtm keys add myaccount
$ ./mtm add-genesis-account validator 1000000000000stake
$ ./mtm add-genesis-account myaccount 1000000000000udai
$ ./mtm gentx validator 1000000000000stake
$ ./mtm collect-gentxs
```

6. Run your test chain:

```
$ ./mtm start
```
