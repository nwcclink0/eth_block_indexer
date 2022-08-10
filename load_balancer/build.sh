#!/bin/bash
docker build -t yuantingwei/eth_block_indexer_http_api:v0.2 -f Dockerfile_http_api .
docker build -t yuantingwei/eth_block_indexer_indexer:v0.2 -f Dockerfile_indexer .
