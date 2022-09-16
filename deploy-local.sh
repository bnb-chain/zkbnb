#!/bin/bash

# Preparation: Install following tools when you first run this script!!!
# GOBIN=/usr/local/bin/ go install  github.com/zeromicro/go-zero/tools/goctl@latest
# yum install jq -y
# npm install pm2 -g
# You should install nodejs above v14

# Attention: Set the following variables to the right one before running!!!
DEPLOY_PATH=~/zkbnb-deploy
KEY_PATH=~/.zkbnb
ZkBNB_REPO_PATH=$(cd `dirname $0`; pwd)
CMC_TOKEN=cfce503f-fake-fake-fake-bbab5257dac8

export PATH=$PATH:/usr/local/go/bin:/usr/local/go/bin:/root/go/bin
echo '0. stop old database/redis and docker run new database/redis'
pm2 delete all
docker kill $(docker ps -q)
docker rm $(docker ps -a -q)
docker run -d --name zkbnbredis -p 6379:6379 redis
docker run --name postgres -p 5432:5432 -e PGDATA=/var/lib/postgresql/pgdata  -e POSTGRES_PASSWORD=ZkBNB@123 -e POSTGRES_USER=postgres -e POSTGRES_DB=zkbnb -d postgres


echo '1. basic config and git clone repos'
export PATH=$PATH:/usr/local/go/bin/
cd ~
rm -rf ${DEPLOY_PATH}-bak && mv ${DEPLOY_PATH} ${DEPLOY_PATH}-bak
mkdir -p ${DEPLOY_PATH} && cd ${DEPLOY_PATH}
git clone --branch develop  https://github.com/bnb-chain/zkbnb-contract.git
git clone --branch develop https://github.com/bnb-chain/zkbnb-crypto.git
cp -r ${ZkBNB_REPO_PATH} ${DEPLOY_PATH}


flag=$1
if [ $flag = "new" ]; then
  echo "new crypto env"
  echo '2. start generate zkbnb.vk and zkbnb.pk'
  cd ${DEPLOY_PATH}
  cd zkbnb-crypto && go test ./legend/circuit/bn254/solidity -timeout 99999s -run TestExportSol
  cd ${DEPLOY_PATH}
  mkdir -p $KEY_PATH
  cp -r ./zkbnb-crypto/legend/circuit/bn254/solidity/* $KEY_PATH
fi



echo '3. start verify_parse for ZkBNBVerifier'
cd ${DEPLOY_PATH}/zkbnb/service/prover/
python3 verifier_parse.py ${KEY_PATH}/ZkBNBVerifier1.sol,${KEY_PATH}/ZkBNBVerifier10.sol 1,10 ${DEPLOY_PATH}/zkbnb-contract/contracts/ZkBNBVerifier.sol



echo '4-1. get latest block number'
hexNumber=`curl -X POST 'https://data-seed-prebsc-1-s1.binance.org:8545' --header 'Content-Type: application/json' --data-raw '{"jsonrpc":"2.0", "method":"eth_blockNumber", "params": [], "id":1 }' | jq -r '.result'`
blockNumber=`echo $((${hexNumber}))`
echo 'latest block number = ' $blockNumber



echo '4-2. deploy contracts, register and deposit on BSC Testnet'
cd ${DEPLOY_PATH}
cd ./zkbnb-contract &&  yarn install
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



cd ${DEPLOY_PATH}/zkbnb/
make api-server
cd ${DEPLOY_PATH}/zkbnb && go mod tidy


echo "6. init tables on database"
go run ./cmd/zkbnb/main.go db initialize --dsn "host=localhost user=postgres password=ZkBNB@123 dbname=zkbnb port=5432 sslmode=disable" --contractAddr ${DEPLOY_PATH}/zkbnb/tools/dbinitializer/contractaddr.yaml


cd ${DEPLOY_PATH}/zkbnb/
make api-server


sleep 30s


echo "7. run prover"

echo -e "
Name: prover
Postgres:
  DataSource: host=127.0.0.1 user=postgres password=ZkBNB@123 dbname=zkbnb port=5432 sslmode=disable

CacheRedis:
  - Host: 127.0.0.1:6379
    Type: node

KeyPath:
  ProvingKeyPath: [${KEY_PATH}/zkbnb1.pk, ${KEY_PATH}/zkbnb10.pk]
  VerifyingKeyPath: [${KEY_PATH}/zkbnb1.vk, ${KEY_PATH}/zkbnb10.vk]

BlockConfig:
  OptionalBlockSizes: [1, 10]

TreeDB:
  Driver: memorydb
" > ${DEPLOY_PATH}/zkbnb/service/prover/etc/config.yaml

echo -e "
go run ./cmd/zkbnb/main.go prover --config ${DEPLOY_PATH}/zkbnb/service/prover/etc/config.yaml
" > run_prover.sh
pm2 start --name prover "./run_prover.sh"





echo "8. run witness"

echo -e "
Name: witness

Postgres:
  DataSource: host=127.0.0.1 user=postgres password=ZkBNB@123 dbname=zkbnb port=5432 sslmode=disable

CacheRedis:
  - Host: 127.0.0.1:6379
    Type: node

TreeDB:
  Driver: memorydb
" > ${DEPLOY_PATH}/zkbnb/service/witness/etc/config.yaml

echo -e "
go run ./cmd/zkbnb/main.go witness --config ${DEPLOY_PATH}/zkbnb/service/witness/etc/config.yaml
" > run_witness.sh
pm2 start --name witness "./run_witness.sh"


echo "9. run monitor"

echo -e "
Name: monitor

Postgres:
  DataSource: host=127.0.0.1 user=postgres password=ZkBNB@123 dbname=zkbnb port=5432 sslmode=disable

CacheRedis:
  - Host: 127.0.0.1:6379
    Type: node

ChainConfig:
  NetworkRPCSysConfigName: "BscTestNetworkRpc"
  #NetworkRPCSysConfigName: "LocalTestNetworkRpc"
  StartL1BlockHeight: $blockNumber
  ConfirmBlocksCount: 0
  MaxHandledBlocksCount: 5000
  KeptHistoryBlocksCount: 100000

TreeDB:
  Driver: memorydb
" > ${DEPLOY_PATH}/zkbnb/service/monitor/etc/config.yaml

echo -e "
go run ./cmd/zkbnb/main.go monitor --config ${DEPLOY_PATH}/zkbnb/service/monitor/etc/config.yaml
" > run_monitor.sh
pm2 start --name monitor "./run_monitor.sh"


echo "10. run committer"

echo -e "
Name: committer

Postgres:
  DataSource: host=127.0.0.1 user=postgres password=ZkBNB@123 dbname=zkbnb port=5432 sslmode=disable

CacheRedis:
  - Host: 127.0.0.1:6379
    Type: node

BlockConfig:
  OptionalBlockSizes: [1, 10]

TreeDB:
  Driver: memorydb
" > ${DEPLOY_PATH}/zkbnb/service/committer/etc/config.yaml

echo -e "
go run ./cmd/zkbnb/main.go committer --config ${DEPLOY_PATH}/zkbnb/service/committer/etc/config.yaml
" > run_committer.sh
pm2 start --name committer "./run_committer.sh"


echo "11. run sender"

echo -e "
Name: sender

Postgres:
  DataSource: host=127.0.0.1 user=postgres password=ZkBNB@123 dbname=zkbnb port=5432 sslmode=disable

CacheRedis:
  - Host: 127.0.0.1:6379
    Type: node

ChainConfig:
  NetworkRPCSysConfigName: "BscTestNetworkRpc"
  #NetworkRPCSysConfigName: "LocalTestNetworkRpc"
  ConfirmBlocksCount: 0
  MaxWaitingTime: 120
  MaxBlockCount: 4
  Sk: "acbaa269bd7573ff12361be4b97201aef019776ea13384681d4e5ba6a88367d9"
  GasLimit: 5000000

TreeDB:
  Driver: memorydb
" > ${DEPLOY_PATH}/zkbnb/service/sender/etc/config.yaml

echo -e "
go run ./cmd/zkbnb/main.go sender --config ${DEPLOY_PATH}/zkbnb/service/sender/etc/config.yaml
" > run_sender.sh
pm2 start --name sender "./run_sender.sh"


echo "12. run api-server"

echo -e "
Name: api-server
Host: 0.0.0.0
Port: 8888

Prometheus:
  Host: 0.0.0.0
  Port: 9091
  Path: /metrics

Postgres:
  DataSource: host=127.0.0.1 user=postgres password=ZkBNB@123 dbname=zkbnb port=5432 sslmode=disable

CacheRedis:
  - Host: 127.0.0.1:6379
    Type: node

LogConf:
  ServiceName: api-server
  Mode: console
  Path: ./log/api-server
  StackCooldownMillis: 500
  Level: error

CoinMarketCap:
  Url: https://pro-api.coinmarketcap.com/v1/cryptocurrency/quotes/latest?symbol=
  Token: ${CMC_TOKEN}

MemCache:
  AccountExpiration: 200
  AssetExpiration:   600
  BlockExpiration:   400
  TxExpiration:      400
  PriceExpiration:   1000
  " > ${DEPLOY_PATH}/zkbnb/service/apiserver/etc/config.yaml

echo -e "
go run ./cmd/zkbnb/main.go apiserver --config ${DEPLOY_PATH}/zkbnb/service/apiserver/etc/config.yaml
" > run_apiserver.sh
pm2 start --name apiserver "./run_apiserver.sh"
