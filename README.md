# Microtick Zone

This is the proof of concept port of Microtick to the [Cosmos SDK](https://github.com/cosmos/cosmos-sdk).

## Instructions for building

You must set GOPATH to a working directory somewhere and have your PATH set to point to its bin directory.

```
$ git clone https://github.com/mjackson001/mtzone.git
$ cd mtzone
$ export GOPATH=<some absolute path to a working directory>
$ export PATH=$PATH:$GOPATH/bin
$ make
```

## Instructions for running

1. Clean out any existing working data
```
$ rm -r $HOME/.microtick
```

2. Initialize working data
```
$ mtd init --chain-id=<your chain id>
```

3. Create a validator with staking tokens (stake) and a test account with test tokens (udai).  1 dai = 1000000 udai.
```
$ mtcli config chain-id <your chain id>
$ mtcli config output text
$ mtcli config trust-node true
$ mtcli keys add validator
$ mtcli keys add test
$ mtd add-genesis-account $(mtcli keys show validator -a) 1000000000000stake
$ mtd add-genesis-account $(mtcli keys show test -a) 1000000000000udai
$ mtd gentx --name validator
$ mtd collect-gentxs
```

5. Run your test chain:
```
$ mtd start
```
