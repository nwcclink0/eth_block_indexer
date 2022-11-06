#!/bin/bash

TARGET=$1
GO_OSARCH=
GO_BUILD_TAGS=
OUTPUT_SUBDIR=
export GOOS=linux
#export CGO_ENABLED=1
export GOARCH=amd64
GO_BUILD_TAGS=""
if [ "$TARGET" = "github_action" ]; then
  go build -v -o="./load_balancer/release/eth_block_indexer"
else
  go build -v -o="./load_balancer/release/eth_block_indexer"
  docker build -t eth_block_indexer_indexer:v0.3 -f load_balancer/Dockerfile_indexer .
  docker build -t eth_block_indexer_http_api:v0.3 -f load_balancer/Dockerfile_http_api .
fi
