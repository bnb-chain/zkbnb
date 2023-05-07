#!/bin/bash

blockHeight=$1
#
#Production Environment
# don't stop api server service
# /api/v1/sendTx and /api/v1/updateNftByIndex stop accepting requests，

echo "stop services"
#pm2 stop apiserver
pm2 stop sender
pm2 stop prover
pm2 stop monitor
pm2 stop committer
pm2 stop witness

pm2 stop nft-server

echo "sleep 30s"
sleep 30s

echo "start rollback"

# get config from apollo
go run ./cmd/zkbnb/main.go rollback --height ${blockHeight}

# get config from file
#go run ./cmd/zkbnb/main.go rollback --config ./tools/rollback/etc/config.yaml --height ${blockHeight}

echo "start services"
#pm2 start apiserver
pm2 start committer
pm2 start witness
pm2 start prover
pm2 start sender
pm2 start monitor

pm2 start nft-server


sh run_whole_rolllback.sh > run_whole_rolllback.log 2>&1