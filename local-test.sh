#!/bin/bash

export PATH=$PATH:/usr/local/go/bin:/usr/local/go/bin:/root/go/bin

export ZkBNB=`cat ~/zkbnb-deploy/zkbnb-contract/info/addresses.json  | jq -r '.zkbnbProxy'`
export AssetGov=`cat ~/zkbnb-deploy/zkbnb-contract/info/addresses.json  | jq -r '.assetGovernance'`
export TestLogLevel=2
export L1EndPoint=https://data-seed-prebsc-1-s1.binance.org:8545
export L2EndPoint=http://127.0.0.1:8888

cd /tmp && rm -rf ./zkbnb-integration-test
git clone --branch main https://github.com/bnb-chain/zkbnb-integration-test.git
cd ./zkbnb-integration-test/tests

echo '1. start TestSetupSuite'
go test -v -run TestSetupSuite -timeout 30m

echo '2. start L1 test'
go test -v -run TestL1Suite -timeout 30m

echo '3. start L2 test'
go test -v -run TestL2Suite -timeout 30m
