#!/usr/bin/env bash

# Preparation: Install following tools when you first run this script!!!
# GOBIN=/usr/local/bin/ go install  github.com/zeromicro/go-zero/tools/goctl@latest
# yum install jq -y
# You should install nodejs above v14

# Attention: Set the following variables to the right one before running!!!
WORKDIR=$(cd `dirname $0`/..; pwd)
KEY_PATH=${WORKDIR}/.zkbnb
ZkBNB_CONTRACT_REPO=https://github.com/bnb-chain/zkbnb-contract.git
ZkBNB_CRYPTO_REPO=https://github.com/bnb-chain/zkbnb-crypto.git
BSC_TESTNET_ENDPOINT=https://data-seed-prebsc-1-s1.binance.org:8545
ZKBNB_CRYPTO_BRANCH=$(cat $WORKDIR/../go.mod | grep github.com/bnb-chain/zkbnb-crypto | awk -F" " '{print $2}' | awk -F"-" '{if ($3 != "") print $3;else print $1;}')

export PATH=$PATH:/usr/local/go/bin:/usr/local/go/bin:/root/go/bin

function prepare() {
    echo 'basic config and git clone repos ...'
    rm -rf ${WORKDIR}/dependency
    mkdir -p ${WORKDIR}/dependency && cd ${WORKDIR}/dependency

    git clone --branch develop ${ZkBNB_CONTRACT_REPO}
    git clone --branch develop ${ZkBNB_CRYPTO_REPO}
    cd ${WORKDIR}/dependency/zkbnb-crypto && git checkout ${ZKBNB_CRYPTO_BRANCH}

    if [ ! -z $1 ] && [ "$1" = "new" ]; then
        echo "new crypto env"
        echo 'start generate zkbnb.vk and zkbnb.pk ...'
        cd ${WORKDIR}/dependency/zkbnb-crypto
        go test ./circuit/solidity -timeout 99999s -run TestExportSol
        mkdir -p $KEY_PATH
        cp -r ./circuit/solidity/* $KEY_PATH/
    fi

    echo 'start verify_parse for ZkBNBVerifier ...'
    cd ${WORKDIR}/../service/prover/
    python3 verifier_parse.py ${KEY_PATH}/ZkBNBVerifier1.sol 1 ${WORKDIR}/dependency/zkbnb-contract/contracts/ZkBNBVerifier.sol
}

function getLatestBlockHeight() {
    hexNumber=$(curl -X POST $BSC_TESTNET_ENDPOINT --header 'Content-Type: application/json' --data-raw '{"jsonrpc":"2.0", "method":"eth_blockNumber", "params": [], "id":1 }' | jq -r '.result')
    blockNumber=`echo $((${hexNumber}))`
    
    echo $blockNumber
}

function deployContracts() {
    echo 'deploy contracts, register and deposit on BSC Testnet'
    cd ${WORKDIR}/dependency/zkbnb-contract && npm install
    cp /server/.env ./
    npx hardhat --network BSCTestnet run ./scripts/deploy-keccak256/deploy.js
    echo "Recorded latest contract addresses into ${WORKDIR}/dependency/zkbnb-contract/info/addresses.json"
    npx hardhat --network BSCTestnet run ./scripts/deploy-keccak256/register.js
    npx hardhat --network BSCTestnet run ./scripts/deploy-keccak256/deposit.js

    mkdir -p ${WORKDIR}/configs/
    echo 'modify deployed contracts into zkbnb config ...'
    cp -r ${WORKDIR}/../tools/dbinitializer/contractaddr.yaml.example ${WORKDIR}/configs/contractaddr.yaml

    ZkBNBContractAddr=`cat ${WORKDIR}/dependency/zkbnb-contract/info/addresses.json  | jq -r '.zkbnbProxy'`
    sed -i -e "s/ZkBNBProxy: .*/ZkBNBProxy: ${ZkBNBContractAddr}/" ${WORKDIR}/configs/contractaddr.yaml

    GovernanceContractAddr=`cat ${WORKDIR}/dependency/zkbnb-contract/info/addresses.json  | jq -r '.governance'`
    sed -i -e "s/Governance: .*/Governance: ${GovernanceContractAddr}/" ${WORKDIR}/configs/contractaddr.yaml
}

CMD=$1
case ${CMD} in
prepare)
    prepare $2
    ;;
blockHeight)
    blockNumber=$(getLatestBlockHeight)
    echo "$blockNumber"
    ;;
deployContracts)
    deployContracts
    ;;
all)
    prepare $2
    deployContracts
    ;;
*)
    echo "Usage: tool.sh prepare | blockHeight | deployContracts | all "
    ;;
esac
