#!/bin/bash

# Preparation: Install following tools when you first run this script!!!
# GOBIN=/usr/local/bin/ go install  github.com/zeromicro/go-zero/tools/goctl@latest
# yum install jq -y
# npm install pm2 -g
# You should install nodejs above v14
# sh deploy-local.sh new  // append the new parameter to generate pk and vk data when you first run this script.
##
# Attention: Set the following variables to the right one before running!!!
DEPLOY_PATH=~/zkbnb-deploy
KEY_PATH=~/.zkbnb
ZkBNB_REPO_PATH=$(cd `dirname $0`; pwd)
CMC_TOKEN=cfce503f-fake-fake-fake-bbab5257dac8
NETWORK_RPC_SYS_CONFIG_NAME=LocalTestNetworkRpc # BscTestNetworkRpc or LocalTestNetworkRpc
BSC_TESTNET_RPC=HTTP://127.0.0.1:8545
BSC_TESTNET_PRIVATE_KEY=0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80
#use COMMIT_BLOCK_PRIVATE_KEY for submitting commit_block to bnb contract in sender application
#use VERIFY_BLOCK_PRIVATE_KEY for submitting verify_block to bnb contract in sender application
COMMIT_BLOCK_PRIVATE_KEY=0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80
VERIFY_BLOCK_PRIVATE_KEY=0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80
# security Council Members for upgrade approve
# FOR TEST
# generage by Mnemonic (account #17 ~ #19): giggle federal note disorder will close traffic air melody artefact taxi tissue
SECURITY_COUNCIL_MEMBERS_NUMBER_1=0x0000000000000000000000000000000000000000
SECURITY_COUNCIL_MEMBERS_NUMBER_2=0x0000000000000000000000000000000000000000
SECURITY_COUNCIL_MEMBERS_NUMBER_3=0x0000000000000000000000000000000000000000
# validator config, split by `,` the address of COMMIT_BLOCK_PRIVATE_KEY  and the address of VERIFY_BLOCK_PRIVATE_KEY,
VALIDATORS=
# treasury account address, the default value is the first validator's address
TREASURY_ACCOUNT_ADDRESS=
# gas account address, the default value is the second validator's address
GAS_ACCOUNT_ADDRESS=
ZKBNB_OPTIONAL_BLOCK_SIZES=1,10
ZKBNB_R1CS_BATCH_SIZE=10000

export PATH=$PATH:/usr/local/go/bin:/usr/local/go/bin:/root/go/bin
echo '0. stop old database/redis and docker run new database/redis'
pm2 delete all
ZKBNB_CONTAINERS=$(docker ps -a |grep zkbnb|awk '{print $1}')
[[ -z "${ZKBNB_CONTAINERS}" ]] || docker rm -f ${ZKBNB_CONTAINERS}
docker run -d --name zkbnb-redis -p 6379:6379 redis
docker run -d --name zkbnb-postgres -p 5432:5432 \
  -e PGDATA=/var/lib/postgresql/pgdata  \
  -e POSTGRES_PASSWORD=ZkBNB@123 \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_DB=zkbnb postgres


echo '1. basic config and git clone repos'
export PATH=$PATH:/usr/local/go/bin/
cd ~
rm -rf ${DEPLOY_PATH}-bak && mv ${DEPLOY_PATH} ${DEPLOY_PATH}-bak
mkdir -p ${DEPLOY_PATH} && cd ${DEPLOY_PATH}
git clone --branch bugfix/gnark-crypto-version  https://github.com/ruslangm/zkbnb-contract.git
git clone --branch feat/sha256-v0.8.0-load-opt https://github.com/ruslangm/zkbnb-crypto.git
cp -r ${ZkBNB_REPO_PATH} ${DEPLOY_PATH}


flag=$1
if [ $flag = "new" ]; then
  echo "new crypto env"
  echo '2. start generate zkbnb.vk and zkbnb.pk'
  cd ${DEPLOY_PATH}
  cd zkbnb-crypto && go test ./circuit/solidity -timeout 99999s -run TestExportSol -blocksizes=${ZKBNB_OPTIONAL_BLOCK_SIZES}
  cd ${DEPLOY_PATH}
  mkdir -p $KEY_PATH
  cp -r ./zkbnb-crypto/circuit/solidity/* $KEY_PATH
fi



echo '3. start verify_parse for ZkBNBVerifier'
cd ${DEPLOY_PATH}/zkbnb/service/prover/
contracts=()
keys=()
i=0
for size in $(echo $ZKBNB_OPTIONAL_BLOCK_SIZES | tr ',' ' '); do
  contracts[$i]="${KEY_PATH}/ZkBNBVerifier${size}.sol"
  keys[$i]="${KEY_PATH}/zkbnb${size}"
  i=$((i+1))
done
VERIFIER_CONTRACTS=$(echo "${contracts[*]}" | tr ' ' ',')
PROVING_KEYS=$(echo "${keys[*]}" | tr ' ' ',')
python3 verifier_parse.py ${VERIFIER_CONTRACTS} ${ZKBNB_OPTIONAL_BLOCK_SIZES} ${DEPLOY_PATH}/zkbnb-contract/contracts/ZkBNBVerifier.sol

echo '4-1. get latest block number'
hexNumber=`curl -X POST ${BSC_TESTNET_RPC} --header 'Content-Type: application/json' --data-raw '{"jsonrpc":"2.0", "method":"eth_blockNumber", "params": [], "id":1 }' | jq -r '.result'`
blockNumber=`echo $((${hexNumber}))`
echo 'latest block number = ' $blockNumber



echo '4-2. deploy contracts, register and deposit on BSC Testnet'
cd ${DEPLOY_PATH}
cd ./zkbnb-contract
cp -r .env.example .env
sed -i -e "s~BSC_TESTNET_RPC=.*~BSC_TESTNET_RPC=${BSC_TESTNET_RPC}~" .env
sed -i -e "s/BSC_TESTNET_PRIVATE_KEY=.*/BSC_TESTNET_PRIVATE_KEY=${BSC_TESTNET_PRIVATE_KEY}/" .env
sed -i -e "s/SECURITY_COUNCIL_MEMBERS_NUMBER_1=.*/SECURITY_COUNCIL_MEMBERS_NUMBER_1=${SECURITY_COUNCIL_MEMBERS_NUMBER_1}/" .env
sed -i -e "s/SECURITY_COUNCIL_MEMBERS_NUMBER_2=.*/SECURITY_COUNCIL_MEMBERS_NUMBER_2=${SECURITY_COUNCIL_MEMBERS_NUMBER_2}/" .env
sed -i -e "s/SECURITY_COUNCIL_MEMBERS_NUMBER_3=.*/SECURITY_COUNCIL_MEMBERS_NUMBER_3=${SECURITY_COUNCIL_MEMBERS_NUMBER_3}/" .env
sed -i -e "s/VALIDATORS=.*/VALIDATORS=${VALIDATORS}/" .env
sed -i -e "s/TREASURY_ACCOUNT_ADDRESS=.*/TREASURY_ACCOUNT_ADDRESS=${TREASURY_ACCOUNT_ADDRESS}/" .env
sed -i -e "s/GAS_ACCOUNT_ADDRESS=.*/GAS_ACCOUNT_ADDRESS=${GAS_ACCOUNT_ADDRESS}/" .env
yarn install
npx hardhat --network BSCTestnet run ./scripts/deploy-keccak256/deploy.js
echo 'Recorded latest contract addresses into ${DEPLOY_PATH}/zkbnb-contract/info/addresses.json'

npx hardhat --network BSCTestnet run ./scripts/deploy-keccak256/register.js
npx hardhat --network BSCTestnet run ./scripts/deploy-keccak256/deposit.js


echo '5. modify deployed contracts into zkbnb config'
cd ${DEPLOY_PATH}/zkbnb/tools/dbinitializer/
cp -r ./contractaddr.yaml.example ./contractaddr.yaml

ZkBNBContractAddr=`cat ${DEPLOY_PATH}/zkbnb-contract/info/addresses.json  | jq -r '.zkbnbProxy'`
sed -i -e "s/ZkBNBProxy: .*/ZkBNBProxy: ${ZkBNBContractAddr}/" ${DEPLOY_PATH}/zkbnb/tools/dbinitializer/contractaddr.yaml

GovernanceContractAddr=`cat ${DEPLOY_PATH}/zkbnb-contract/info/addresses.json  | jq -r '.governance'`
sed -i -e "s/Governance: .*/Governance: ${GovernanceContractAddr}/" ${DEPLOY_PATH}/zkbnb/tools/dbinitializer/contractaddr.yaml

BUSDContractAddr=`cat ${DEPLOY_PATH}/zkbnb-contract/info/addresses.json  | jq -r '.BUSDToken'`
sed -i -e "s/BUSDToken: .*/BUSDToken: ${BUSDContractAddr}/" ${DEPLOY_PATH}/zkbnb/tools/dbinitializer/contractaddr.yaml

DefaultNftFactoryAddr=`cat ${DEPLOY_PATH}/zkbnb-contract/info/addresses.json  | jq -r '.DefaultNftFactory'`
sed -i -e "s/DefaultNftFactory: .*/DefaultNftFactory: ${DefaultNftFactoryAddr}/" ${DEPLOY_PATH}/zkbnb/tools/dbinitializer/contractaddr.yaml

 cd ${DEPLOY_PATH}/zkbnb/
 make api-server
cd ${DEPLOY_PATH}/zkbnb && go mod tidy

echo "6. init tables on database"
go run ./cmd/zkbnb/main.go db initialize --dsn "host=127.0.0.1 user=postgres password=ZkBNB@123 dbname=zkbnb port=5432 sslmode=disable" --contractAddr ${DEPLOY_PATH}/zkbnb/tools/dbinitializer/contractaddr.yaml

sleep 10s

echo "7. run prover"

echo -e "
Name: prover
Postgres:
  MasterDataSource: host=127.0.0.1 user=postgres password=ZkBNB@123 dbname=zkbnb port=5432 sslmode=disable
  SlaveDataSource: host=127.0.0.1 user=postgres password=ZkBNB@123 dbname=zkbnb port=5432 sslmode=disable

CacheRedis:
  - Host: 127.0.0.1:6379
    Type: node

KeyPath: [${PROVING_KEYS}]

BlockConfig:
  OptionalBlockSizes: [${ZKBNB_OPTIONAL_BLOCK_SIZES}]
  R1CSBatchSize: ${ZKBNB_R1CS_BATCH_SIZE}

TreeDB:
  Driver: memorydb
  AssetTreeCacheSize: 512000
" > ${DEPLOY_PATH}/zkbnb/service/prover/etc/config.yaml

echo -e "
go run ./cmd/zkbnb/main.go prover --config ${DEPLOY_PATH}/zkbnb/service/prover/etc/config.yaml --pprof --pprof.addr 127.0.0.1 --pprof.port 6060 --metrics --metrics.addr 127.0.0.1 --metrics.port 6060
" > run_prover.sh
# remove the fist line if it includes -e
sed -i '' -e '/-e/,1d' ${DEPLOY_PATH}/zkbnb/service/prover/etc/config.yaml
sed -i '' -e '/-e/,1d' run_prover.sh
pm2 start --name prover "./run_prover.sh"

echo "8. run witness"

echo -e "
Name: witness

Postgres:
  MasterDataSource: host=127.0.0.1 user=postgres password=ZkBNB@123 dbname=zkbnb port=5432 sslmode=disable
  SlaveDataSource: host=127.0.0.1 user=postgres password=ZkBNB@123 dbname=zkbnb port=5432 sslmode=disable

CacheRedis:
  - Host: 127.0.0.1:6379
    Type: node

TreeDB:
  Driver: memorydb
  AssetTreeCacheSize: 512000
" > ${DEPLOY_PATH}/zkbnb/service/witness/etc/config.yaml

echo -e "
go run ./cmd/zkbnb/main.go witness --config ${DEPLOY_PATH}/zkbnb/service/witness/etc/config.yaml --pprof --pprof.addr 127.0.0.1 --pprof.port 6061 --metrics --metrics.addr 127.0.0.1 --metrics.port 6061
" > run_witness.sh
# remove the fist line if it includes -e
sed -i '' -e '/-e/,1d' ${DEPLOY_PATH}/zkbnb/service/witness/etc/config.yaml
sed -i '' -e '/-e/,1d' run_witness.sh
pm2 start --name witness "./run_witness.sh"

echo "9. run monitor"

echo -e "
Name: monitor

Postgres:
  MasterDataSource: host=127.0.0.1 user=postgres password=ZkBNB@123 dbname=zkbnb port=5432 sslmode=disable
  SlaveDataSource: host=127.0.0.1 user=postgres password=ZkBNB@123 dbname=zkbnb port=5432 sslmode=disable

CacheRedis:
  - Host: 127.0.0.1:6379
    Type: node
AccountCacheSize: 100000

ChainConfig:
  NetworkRPCSysConfigName: "${NETWORK_RPC_SYS_CONFIG_NAME}"
  #NetworkRPCSysConfigName: "LocalTestNetworkRpc"
  StartL1BlockHeight: $blockNumber
  ConfirmBlocksCount: 0
  MaxHandledBlocksCount: 5000
  KeptHistoryBlocksCount: 100000

TreeDB:
  Driver: memorydb
  AssetTreeCacheSize: 512000
" > ${DEPLOY_PATH}/zkbnb/service/monitor/etc/config.yaml

echo -e "
go run ./cmd/zkbnb/main.go monitor --config ${DEPLOY_PATH}/zkbnb/service/monitor/etc/config.yaml --pprof --pprof.addr 127.0.0.1 --pprof.port 6062 --metrics --metrics.addr 127.0.0.1 --metrics.port 6062
" > run_monitor.sh
# remove the fist line if it includes -e
sed -i '' -e '/-e/,1d' ${DEPLOY_PATH}/zkbnb/service/monitor/etc/config.yaml
sed -i '' -e '/-e/,1d' run_monitor.sh
pm2 start --name monitor "./run_monitor.sh"


echo "10. run committer"

echo -e "
Name: committer

Postgres:
  MasterDataSource: host=127.0.0.1 user=postgres password=ZkBNB@123 dbname=zkbnb port=5432 sslmode=disable
  SlaveDataSource: host=127.0.0.1 user=postgres password=ZkBNB@123 dbname=zkbnb port=5432 sslmode=disable

CacheRedis:
  - Host: 127.0.0.1:6379
    Type: node

BlockConfig:
  OptionalBlockSizes: [${ZKBNB_OPTIONAL_BLOCK_SIZES}]

IpfsUrl:
  10.23.23.40:5001

TreeDB:
  Driver: memorydb
  AssetTreeCacheSize: 512000
" > ${DEPLOY_PATH}/zkbnb/service/committer/etc/config.yaml

echo -e "
go run ./cmd/zkbnb/main.go committer --config ${DEPLOY_PATH}/zkbnb/service/committer/etc/config.yaml --pprof --pprof.addr 127.0.0.1 --pprof.port 6063 --metrics --metrics.addr 127.0.0.1 --metrics.port 6063
" > run_committer.sh
# remove the fist line if it includes -e
sed -i '' -e '/-e/,1d' ${DEPLOY_PATH}/zkbnb/service/committer/etc/config.yaml
sed -i '' -e '/-e/,1d' run_committer.sh
pm2 start --name committer "./run_committer.sh"


echo "11. run sender"

echo -e "
Name: sender

Postgres:
  MasterDataSource: host=127.0.0.1 user=postgres password=ZkBNB@123 dbname=zkbnb port=5432 sslmode=disable
  SlaveDataSource: host=127.0.0.1 user=postgres password=ZkBNB@123 dbname=zkbnb port=5432 sslmode=disable

CacheRedis:
  - Host: 127.0.0.1:6379
    Type: node

ChainConfig:
  NetworkRPCSysConfigName: "${NETWORK_RPC_SYS_CONFIG_NAME}"
  #NetworkRPCSysConfigName: LocalTestNetworkRpc
  ConfirmBlocksCount: 0
  SendSignatureMode: PrivateKeySignMode
  MaxWaitingTime: 120
  MaxBlockCount: 4
  GasLimit: 5000000
  GasPrice: 0

Apollo:
  AppID:             zkbnb-cloud
  Cluster:           prod
  ApolloIp:          http://internal-tf-cm-test-apollo-config-alb-2119591301.ap-northeast-1.elb.amazonaws.com:9028
  Namespace:         applicationDev
  IsBackupConfig:    true

AuthConfig:
  CommitBlockSk: "${COMMIT_BLOCK_PRIVATE_KEY}"
  VerifyBlockSk: "${VERIFY_BLOCK_PRIVATE_KEY}"

TreeDB:
  Driver: memorydb
  AssetTreeCacheSize: 512000
" > ${DEPLOY_PATH}/zkbnb/service/sender/etc/config.yaml

echo -e "
go run ./cmd/zkbnb/main.go sender --config ${DEPLOY_PATH}/zkbnb/service/sender/etc/config.yaml --pprof --pprof.addr 127.0.0.1 --pprof.port 6064 --metrics --metrics.addr 127.0.0.1 --metrics.port 6064
" > run_sender.sh
# remove the fist line if it includes -e
sed -i '' -e '/-e/,1d' ${DEPLOY_PATH}/zkbnb/service/sender/etc/config.yaml
sed -i '' -e '/-e/,1d' run_sender.sh
pm2 start --name sender "./run_sender.sh"


echo "12. run api-server"

echo -e "
Name: api-server
Host: 127.0.0.1
Port: 8888

TxPool:
  MaxPendingTxCount: 10000

Postgres:
  MasterDataSource: host=127.0.0.1 user=postgres password=ZkBNB@123 dbname=zkbnb port=5432 sslmode=disable
  SlaveDataSource: host=127.0.0.1 user=postgres password=ZkBNB@123 dbname=zkbnb port=5432 sslmode=disable
  MaxConn: 1000
  MaxIdle: 10

CacheRedis:
  - Host: 127.0.0.1:6379
    Type: node

LogConf:
  ServiceName: api-server
  Mode: console
  Path: ./log/api-server
  StackCooldownMillis: 500
  Level: error

Apollo:
  AppID:             zkbnb-cloud
  Cluster:           prod
  ApolloIp:          http://internal-tf-cm-test-apollo-config-alb-2119591301.ap-northeast-1.elb.amazonaws.com:9028
  Namespace:         applicationDev
  IsBackupConfig:    true

IpfsUrl:
  10.23.23.40:5001

CoinMarketCap:
  Url: https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest?symbol=
  Token: ${CMC_TOKEN}

BinanceOracle:
  Url: http://cloud-oracle-gateway.qa1fdg.net:9902
  Apikey: b11f867a6b8fed571720fbb8155f65b5f589f291c35148c41c2f7b81b9177c47
  ApiSecret: 7a1f315f47aea8f8a451d5f5a8bfa7dc7dea292fff7c8ed27a6294a03ec4f974

MemCache:
  AccountExpiration: 200
  AssetExpiration:   600
  BlockExpiration:   400
  TxExpiration:      400
  PriceExpiration:   3600000
  MaxCounterNum:     100000
  MaxKeyNum:         10000
  " > ${DEPLOY_PATH}/zkbnb/service/apiserver/etc/config.yaml

echo -e "
go run ./cmd/zkbnb/main.go apiserver --config ${DEPLOY_PATH}/zkbnb/service/apiserver/etc/config.yaml --pprof --pprof.addr 127.0.0.1 --pprof.port 6065 --metrics --metrics.addr 127.0.0.1 --metrics.port 6065
" > run_apiserver.sh
# remove the fist line if it includes -e
sed -i '' -e '/-e/,1d' ${DEPLOY_PATH}/zkbnb/service/apiserver/etc/config.yaml
sed -i '' -e '/-e/,1d' run_apiserver.sh
pm2 start --name apiserver "./run_apiserver.sh"
