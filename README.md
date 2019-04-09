# Microtick Zone

This is the proof of concept port of Microtick to the [Cosmos SDK](https://github.com/cosmos/cosmos-sdk).

## Instructions for building

These instructions assuming a working Go installation and GOPATH

```
$ cd $GOPATH/src
$ mkdir -p github.com/mjackson001/mtzone
$ cd github.com/mjackson001/mtzone
$ make
```

## Instructions for running

1. Remove any existing configuration directories:
```
$ rm -rf ~/.mtd
$ rm -rf ~/.mtcli
```

2. ```$ mtd init --chain-id=mtzone```

3. Create several keys and add the keys to the genesis file
```
$ mtcli keys add mykey1
$ mtcli keys add mykey2
$ mtd add-genesis-account $(mtcli keys show mykey1 -a) 1000fox
$ mtd add-genesis-account $(mtcli keys show mykey2 -a) 1000fox
```

4. Set up the command line tool:
```
$ mtcli config chain-id mtzone
$ mtcli config output text
$ mtcli config trust-node true
```

5. Run the chain:
```
$ mtd start
```

Next step: [Creating a Market](https://github.com/mjackson001/mtzone/blob/master/doc/createmarket.md)
