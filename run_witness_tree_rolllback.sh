#!/bin/bash

blockHeight=$1
#
#Production Environment
# don't stop api server service
# /api/v1/sendTx and /api/v1/updateNftByIndex stop accepting requests，

# get config from apollo
go run ./cmd/zkbnb/main.go rollbackwitnesssmt --height ${blockHeight}

# get config from file
#go run ./cmd/zkbnb/main.go rollbackwitnesssmt --config ./tools/rollbackwitnesssmt/etc/config.yaml --height ${blockHeight}

#  sh run_witness_tree_rolllback.sh > run_witness_tree_rolllback.log 2>&1
