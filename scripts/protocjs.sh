#!/usr/bin/env bash

rm -rf ./js
mkdir ./js
cd js 
protos=$(find ../proto -name "*.proto")
cosmos_protos=$(find ../vendor/github.com/cosmos/cosmos-sdk/proto -name "*.proto")
gogo_protos=$(find ../vendor/github.com/cosmos/cosmos-sdk/third_party/proto/gogoproto -name "*.proto")
cosmos_protos2=$(find ../vendor/github.com/cosmos/cosmos-sdk/third_party/proto/cosmos_proto -name "*.proto")

../.cache/bin/protoc \
-I "../proto" \
-I "../vendor/github.com/regen-network/cosmos-proto" \
-I "../vendor/github.com/tendermint/tendermint/proto" \
-I "../vendor/github.com/cosmos/cosmos-sdk/proto" \
-I "../vendor/github.com/cosmos/cosmos-sdk/third_party/proto" \
-I "../vendor/github.com/gogo/protobuf" \
-I "../.cache/include" \
--js_out=import_style=commonjs,binary:. \
$protos \
$cosmos_protos \
$gogo_protos \
$cosmos_protos2 

# Closure style

#--js_out=library=mtproto,binary:. \
# Common-JS
#--js_out=import_style=commonjs,binary:. \

# create tarball for use w/ mtapi
cd ..
tar cvfz mtprotojs.tar.gz js
