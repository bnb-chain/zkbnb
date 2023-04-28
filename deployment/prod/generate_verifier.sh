#!/usr/bin/env bash

echo "Generate verifier contract"

# Import! set correct block sizes & key path
BLOCK_SIZES=1,2
KEY_PATH=~/zkbnb-deploy/zkbnb-crypto/circuit/solidity
TEMPLATE="https://raw.githubusercontent.com/bnb-chain/zkbnb-contract/qa/contracts/ZkBNBVerifier.sol"

BASE_PATH=$(cd `dirname $0`/../..; pwd)
cd ${BASE_PATH}/service/prover

contracts=()
keys=()
i=0
for size in $(echo $BLOCK_SIZES | tr ',' ' '); do
 contracts[$i]="${KEY_PATH}/ZkBNBVerifier${size}.sol"
 keys[$i]="${KEY_PATH}/zkbnb${size}"
 i=$((i+1))
done
VERIFIER_CONTRACTS=$(echo "${contracts[*]}" | tr ' ' ',')
PROVING_KEYS=$(echo "${keys[*]}" | tr ' ' ',')

curl -L $TEMPLATE -O $BASE_PATH/ZkBNBVerifier.sol
python3 verifier_parse.py ${VERIFIER_CONTRACTS} ${BLOCK_SIZES} ZkBNBVerifier.sol
rm -rf $BASE_PATH/ZkBNBVerifier.sol
