#!/bin/bash

# Preparation: Install following tools when you first run this script!!!
# GOBIN=/usr/local/bin/ go install  github.com/zeromicro/go-zero/tools/goctl@latest
# yum install jq -y
# npm install pm2 -g
# You should install nodejs above v14
# sh deploy-local.sh new  // append the new parameter to generate pk and vk data when you first run this script.

# Attention: Set the following variables to the right one before running!!!
#DEPLOY_PATH=/zkbnb
DEPLOY_PATH=~/zkbnb-deploy
KEY_PATH=~/.zkbnb
ZkBNB_REPO_PATH=$(cd `dirname $0`; pwd)

ZKBNB_OPTIONAL_BLOCK_SIZES=1
ZKBNB_R1CS_BATCH_SIZE=100000

export PATH=$PATH:/usr/local/go/bin:/usr/local/go/bin:/root/go/bin

echo '1. basic config and git clone repos'
export PATH=$PATH:/usr/local/go/bin/
cd ~
mkdir -p ${DEPLOY_PATH} && cd ${DEPLOY_PATH}
cp -r ${ZkBNB_REPO_PATH} ${DEPLOY_PATH}
cp -r  /Users/user/Documents/work/source/bnb-chain/zkbnb-contract ${DEPLOY_PATH}
cp -r  /Users/user/Documents/work/source/bnb-chain/zkbnb-crypto ${DEPLOY_PATH}


flag=$1
#if [ $flag = "new" ]; then
  echo "new crypto env"
  echo '2. start generate zkbnb.desert.vk and zkbnb.desert.pk'
  cd ${DEPLOY_PATH}
  cd zkbnb-crypto && go test ./circuit/solidity -timeout 99999s -run TestExportDesertSol -batchsize=${ZKBNB_R1CS_BATCH_SIZE}
  cd ${DEPLOY_PATH}
  mkdir -p $KEY_PATH
  cp -r ./zkbnb-crypto/circuit/solidity/* $KEY_PATH
#fi



echo '3. start verify_parse for DesertVerifier'
cd ${DEPLOY_PATH}/zkbnb/service/prover/
contracts=()
keys=()
i=0
for size in $(echo $ZKBNB_OPTIONAL_BLOCK_SIZES | tr ',' ' '); do
  contracts[$i]="${KEY_PATH}/DesertVerifier${size}.sol"
  keys[$i]="${KEY_PATH}/zkbnb.desert${size}"
  i=$((i+1))
done
VERIFIER_CONTRACTS=$(echo "${contracts[*]}" | tr ' ' ',')
PROVING_KEYS=$(echo "${keys[*]}" | tr ' ' ',')
python3 desert_verifier_parse.py ${VERIFIER_CONTRACTS}  ${DEPLOY_PATH}/zkbnb-contract/contracts/DesertVerifier.sol
