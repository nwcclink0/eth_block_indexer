#!/bin/bash

TARGET=$1
GO_OSARCH=
GO_BUILD_TAGS=
OUTPUT_SUBDIR=
export GOOS=linux
#export CGO_ENABLED=1
if [ "$TARGET" = "cloud" ]; then
  export GOARCH=amd64
  GO_BUILD_TAGS=""
  go build -v -o="./load_balancer/release/eth_block_indexer"
else
  go build -v -o="./load_balancer/release/eth_block_indexer"
fi
