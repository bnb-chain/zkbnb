#ZkBNB_REPO_PATH=$(cd `dirname $0`; pwd)

ZKBNB_OPTIONAL_BLOCK_SIZES=8,16,32,64
ZKBNB_R1CS_BATCH_SIZE=100000

mkdir deploy
cd deploy
mkdir r1cs_keys
mkdir gen_sol
DEPLOY_PATH=$(pwd)
KEY_PATH=${DEPLOY_PATH}/r1cs_keys
SOL_PATH=${DEPLOY_PATH}/gen_sol

git clone --branch testnet https://github.com/node-real/zkbnb.git
git clone --branch qa https://github.com/bnb-chain/zkbnb-contract.git
git clone --branch testnet https://github.com/bnb-chain/zkbnb-crypto.git

echo "new pk vk env"
echo '2-1. start generate zkbnb.vk and zkbnb.pk'
cd ${DEPLOY_PATH}
cd zkbnb-crypto && go test -v ./circuit/solidity -timeout 99999s -run TestExportSol -blocksizes=${ZKBNB_OPTIONAL_BLOCK_SIZES} -batchsize=${ZKBNB_R1CS_BATCH_SIZE}
cd ${DEPLOY_PATH}
mkdir -p $KEY_PATH
cp -r ./zkbnb-crypto/circuit/solidity/* ${KEY_PATH}
rm ${KEY_PATH}/*.go

cp -r ${KEY_PATH}/*.sol > ${SOL_PATH}

cd ${DEPLOY_PATH}/zkbnb/service/prover

contracts=()
keys=()
i=0
for size in $(echo $ZKBNB_OPTIONAL_BLOCK_SIZES | tr ',' ' '); do
  contracts[$i]="${SOL_PATH}/ZkBNBVerifier${size}.sol"
  keys[$i]="${KEY_PATH}/zkbnb${size}"
  i=$((i+1))
done
VERIFIER_CONTRACTS=$(echo "${contracts[*]}" | tr ' ' ',')
PROVING_KEYS=$(echo "${keys[*]}" | tr ' ' ',')
python3 verifier_parse.py ${VERIFIER_CONTRACTS} ${ZKBNB_OPTIONAL_BLOCK_SIZES} ${DEPLOY_PATH}/zkbnb-contract/contracts/ZkBNBVerifier.sol

cp ${DEPLOY_PATH}/zkbnb-contract/contracts/ZkBNBVerifier.sol ${SOL_PATH}/ZkBNBVerifier.sol


