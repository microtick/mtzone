#!/usr/bin/env bash

set -eo pipefail

proto_dirs=$(find ./proto -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  .cache/bin/protoc \
  -I "proto" \
  -I "vendor/github.com/regen-network/cosmos-proto" \
  -I "vendor/github.com/tendermint/tendermint/proto" \
  -I "vendor/github.com/cosmos/cosmos-sdk/proto" \
  -I "vendor/github.com/cosmos/cosmos-sdk/third_party/proto" \
  -I "vendor/github.com/gogo/protobuf" \
  -I ".cache/include" \
  --grpc-gateway_out=logtostderr=true:. \
  --gocosmos_out=plugins=interfacetype+grpc,\
Mgoogle/protobuf/any.proto=github.com/cosmos/cosmos-sdk/codec/types:. \
  $(find "${dir}" -maxdepth 1 -name '*.proto')
done

# move proto files to the right places
cp -r github.com/microtick/mtzone/* ./
rm -rf github.com

