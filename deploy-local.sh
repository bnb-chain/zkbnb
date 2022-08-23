#!/bin/bash

# config
# GOBIN=/usr/local/bin/ go install  github.com/zeromicro/go-zero/tools/goctl@latest

# Attention: Set the following variables to the right one before running!!!
DEPLOY_PATH=~/zkbas-deploy
KEY_PATH=/Users/user/.zkbas
ZKBAS_REPO_PATH=/home/ec2-user/zkbas
CMC_TOKEN=cfce503f-fake-fake-fake-bbab5257dac8

export PATH=$PATH:/usr/local/go/bin:/usr/local/go/bin:/root/go/bin
echo '0. stop old database/redis and docker run new database/redis'
pm2 delete all
docker kill $(docker ps -q)
docker rm $(docker ps -a -q)
docker run -d --name zkbasredis -p 6379:6379 redis
docker run --name postgres -p 5432:5432 -e POSTGRES_PASSWORD=Zkbas@123 -e POSTGRES_USER=postgres -e POSTGRES_DB=zkbas -d postgres


echo '1. basic config and git clone repos'
#yum install jq -y
#npm install pm2 -g
export PATH=$PATH:/usr/local/go/bin/
cd ~
rm -rf ${DEPLOY_PATH}-bak && mv ${DEPLOY_PATH} ${DEPLOY_PATH}-bak
mkdir -p ${DEPLOY_PATH} && cd ${DEPLOY_PATH}
git clone --branch develop  https://github.com/bnb-chain/zkbas-contract.git
git clone --branch develop https://github.com/bnb-chain/zkbas-crypto.git
cp -r ${ZKBAS_REPO_PATH} ${DEPLOY_PATH}


flag=$1
if [ $flag = "new" ]; then
  echo "new crypto env"
  echo '2. start generate zkbas.vk and zkbas.pk'
  cd ${DEPLOY_PATH}
  cd zkbas-crypto && go test ./legend/circuit/bn254/solidity -timeout 99999s -run TestExportSol
  cd ${DEPLOY_PATH}
  mkdir -p $KEY_PATH
  cp -r ./zkbas-crypto/legend/circuit/bn254/solidity/* $KEY_PATH
fi



echo '3. start verify_parse for ZkbasVerifier'
cd ${DEPLOY_PATH}/zkbas/service/prover/
python3 verifier_parse.py ${KEY_PATH}/ZkbasVerifier1.sol,${KEY_PATH}/ZkbasVerifier10.sol 1,10 ${DEPLOY_PATH}/zkbas-contract/contracts/ZkbasVerifier.sol



echo '4-1. get latest block number'
hexNumber=`curl -X POST 'https://data-seed-prebsc-1-s1.binance.org:8545' --header 'Content-Type: application/json' --data-raw '{"jsonrpc":"2.0", "method":"eth_blockNumber", "params": [], "id":1 }' | jq -r '.result'`
blockNumber=`echo $((${hexNumber}))`
echo 'latest block number = ' $blockNumber



echo '4-2. deploy contracts, register and deposit on BSC Testnet'
cd ${DEPLOY_PATH}
cd ./zkbas-contract && sudo npm install
npx hardhat --network BSCTestnet run ./scripts/deploy-keccak256/deploy.js
echo 'Recorded latest contract addresses into ${DEPLOY_PATH}/zkbas-contract/info/addresses.json'

npx hardhat --network BSCTestnet run ./scripts/deploy-keccak256/register.js
npx hardhat --network BSCTestnet run ./scripts/deploy-keccak256/deposit.js


echo '5. modify deployed contracts into zkbas config'
cd ${DEPLOY_PATH}/zkbas/common/model/init/
cp -r ./contractaddr.yaml.example ./contractaddr.yaml

ZkbasContractAddr=`cat ${DEPLOY_PATH}/zkbas-contract/info/addresses.json  | jq -r '.zkbasProxy'`
sed -i -e "s/ZkbasProxy: .*/ZkbasProxy: ${ZkbasContractAddr}/" ${DEPLOY_PATH}/zkbas/common/model/init/contractaddr.yaml

GovernanceContractAddr=`cat ${DEPLOY_PATH}/zkbas-contract/info/addresses.json  | jq -r '.governance'`
sed -i -e "s/Governance: .*/Governance: ${GovernanceContractAddr}/" ${DEPLOY_PATH}/zkbas/common/model/init/contractaddr.yaml

sed -i -e "s/BSC_Test_Network_RPC *= .*/BSC_Test_Network_RPC   = \"https\:\/\/data-seed-prebsc-1-s1.binance.org:8545\"/" ${DEPLOY_PATH}/zkbas/common/model/init/init.go



cd ${DEPLOY_PATH}/zkbas/
make api-server
cd ${DEPLOY_PATH}/zkbas && go mod tidy



echo "6. init tables on database"
sed -i -e "s/password=.* dbname/password=Zkbas@123 dbname/" ${DEPLOY_PATH}/zkbas/common/model/basic/connection.go
cd ${DEPLOY_PATH}/zkbas/common/model/init/
go run .


cd ${DEPLOY_PATH}/zkbas/
make api-server


sleep 30s


echo "7. run prover"

echo -e "
Name: prover
Postgres:
  DataSource: host=127.0.0.1 user=postgres password=Zkbas@123 dbname=zkbas port=5432 sslmode=disable

CacheRedis:
  - Host: 127.0.0.1:6379
    Type: node

KeyPath:
  ProvingKeyPath: [${KEY_PATH}/zkbas1.pk, ${KEY_PATH}/zkbas10.pk]
  VerifyingKeyPath: [${KEY_PATH}/zkbas1.vk, ${KEY_PATH}/zkbas10.vk]

BlockConfig:
  OptionalBlockSizes: [1, 10]

TreeDB:
  Driver: memorydb
" > ${DEPLOY_PATH}/zkbas/service/prover/etc/config.yaml

cd ${DEPLOY_PATH}/zkbas/service/prover/
pm2 start --name prover "go run ./main.go"




echo "8. run witness"

echo -e "
Name: witness

Postgres:
  DataSource: host=127.0.0.1 user=postgres password=Zkbas@123 dbname=zkbas port=5432 sslmode=disable

CacheRedis:
  - Host: 127.0.0.1:6379
    Type: node

TreeDB:
  Driver: memorydb
" > ${DEPLOY_PATH}/zkbas/service/witness/etc/config.yaml

cd ${DEPLOY_PATH}/zkbas/service/witness/
pm2 start --name witness "go run ./main.go"


echo "9. run monitor"

echo -e "
Name: monitor

Postgres:
  DataSource: host=127.0.0.1 user=postgres password=Zkbas@123 dbname=zkbas port=5432 sslmode=disable

CacheRedis:
  - Host: 127.0.0.1:6379
    Type: node

ChainConfig:
  NetworkRPCSysConfigName: "BscTestNetworkRpc"
  #NetworkRPCSysConfigName: "LocalTestNetworkRpc"
  StartL1BlockHeight: $blockNumber
  ConfirmBlocksCount: 0
  MaxHandledBlocksCount: 5000

TreeDB:
  Driver: memorydb
" > ${DEPLOY_PATH}/zkbas/service/monitor/etc/config.yaml

cd ${DEPLOY_PATH}/zkbas/service/monitor/
pm2 start --name monitor "go run ./main.go"



echo "10. run committer"

echo -e "
Name: committer

Postgres:
  DataSource: host=127.0.0.1 user=postgres password=Zkbas@123 dbname=zkbas port=5432 sslmode=disable

CacheRedis:
  - Host: 127.0.0.1:6379
    Type: node

BlockConfig:
  OptionalBlockSizes: [1, 10]

TreeDB:
  Driver: memorydb
" > ${DEPLOY_PATH}/zkbas/service/committer/etc/config.yaml

cd ${DEPLOY_PATH}/zkbas/service/committer/
pm2 start --name committer "go run ./main.go"


echo "11. run sender"

echo -e "
Name: sender

Postgres:
  DataSource: host=127.0.0.1 user=postgres password=Zkbas@123 dbname=zkbas port=5432 sslmode=disable

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
" > ${DEPLOY_PATH}/zkbas/service/sender/etc/config.yaml

cd ${DEPLOY_PATH}/zkbas/service/sender/
pm2 start --name sender "go run ./main.go"


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
  DataSource: host=127.0.0.1 user=postgres password=Zkbas@123 dbname=zkbas port=5432 sslmode=disable

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
  PriceExpiration:   200
  " > ${DEPLOY_PATH}/zkbas/service/apiserver/etc/config.yaml

cd ${DEPLOY_PATH}/zkbas/service/apiserver
pm2 start --name api-server "go run ./server.go"
