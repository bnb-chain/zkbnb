#!/usr/bin/env bash

# Import! set correct block sizes & key path
OPTIONAL_BLOCK_SIZES=${BLOCK_SIZES:-"8,16,32,64"}
KEY_PATH=~/zkbnb-deploy/zkbnb-crypto/circuit/solidity
TEMPLATE="https://raw.githubusercontent.com/bnb-chain/zkbnb-contract/qa/contracts/ZkBNBVerifier.sol"

echo "Generate verifier contract for block sizes: ${OPTIONAL_BLOCK_SIZES}"

BASE_PATH=$(cd `dirname $0`/../..; pwd)
cd ${BASE_PATH}/service/prover

contracts=()
keys=()
i=0
for size in $(echo $OPTIONAL_BLOCK_SIZES | tr ',' ' '); do
 contracts[$i]="${KEY_PATH}/ZkBNBVerifier${size}.sol"
 keys[$i]="${KEY_PATH}/zkbnb${size}"
 i=$((i+1))
done
VERIFIER_CONTRACTS=$(echo "${contracts[*]}" | tr ' ' ',')
PROVING_KEYS=$(echo "${keys[*]}" | tr ' ' ',')

curl -L $TEMPLATE -O $BASE_PATH/ZkBNBVerifier.sol
python3 verifier_parse.py ${VERIFIER_CONTRACTS} ${OPTIONAL_BLOCK_SIZES} ZkBNBVerifier.sol
rm -rf $BASE_PATH/ZkBNBVerifier.sol
cp ZkBNBVerifier.sol $BASE_PATH/ZkBNBVerifier.sol