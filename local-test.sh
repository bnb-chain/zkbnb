#!/bin/bash

export PATH=$PATH:/usr/local/go/bin:/usr/local/go/bin:/root/go/bin

export ZkBNB=`cat ~/zkbnb-deploy/zkbnb-contract/info/addresses.json  | jq -r '.zkbnbProxy'`
export AssetGov=`cat ~/zkbnb-deploy/zkbnb-contract/info/addresses.json  | jq -r '.assetGovernance'`
export TestLogLevel=2
export L1EndPoint=http://127.0.0.1:8545
export GovKey=ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80


DEPLOY_PATH=~/zkbnb-deploy

cd ${DEPLOY_PATH}
cd ./zkbnb-integration-test/tests

echo '2. start L1 test'
go test -v -run TestL1Suite -timeout 30m