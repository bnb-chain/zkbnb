#!/bin/bash

# config
# GOBIN=/usr/local/bin/ go install  github.com/zeromicro/go-zero/tools/goctl@latest

export PATH=$PATH:/usr/local/go/bin:/usr/local/go/bin:/root/go/bin
echo '0. stop old database/redis and docker run new database/redis'
pm2 delete all
docker kill $(docker ps -q)
docker rm $(docker ps -a -q)
docker run -d --name zkbnbredis -p 6379:6379 redis
docker run --name postgres -p 5432:5432 -e POSTGRES_PASSWORD=ZkBNB@123 -e POSTGRES_USER=postgres -e POSTGRES_DB=zkbnb -d postgres


echo '1. basic config and git clone repos'
#yum install jq -y
#npm install pm2 -g
export PATH=$PATH:/usr/local/go/bin/
cd ~
rm -rf ~/zkbnb-deploy-bak && mv ~/zkbnb-deploy ~/zkbnb-deploy-bak
mkdir zkbnb-deploy && cd zkbnb-deploy
git clone --branch develop  https://github.com/bnb-chain/zkbnb-contract.git
git clone --branch develop https://github.com/bnb-chain/zkbnb-crypto.git

# mv /home/ec2-user/zkbnb ~/zkbnb-deploy
branch=$1
git clone --branch $branch https://github.com/bnb-chain/zkbnb.git

echo "new crypto env"
echo '2. start generate zkbnb.vk and zkbnb.pk'
cd ~/zkbnb-deploy
cd zkbnb-crypto && go test ./circuit/solidity -timeout 99999s -run TestExportSol
cd ~/zkbnb-deploy
sudo mkdir /home/.zkbnb
cp -r ./zkbnb-crypto/circuit/solidity/* /home/.zkbnb


echo '3. start verify_parse for ZkBNBVerifier'
cd ~/zkbnb-deploy/zkbnb/service/cronjob/prover/
python3 verifier_parse.py /home/.zkbnb/ZkBNBVerifier1.sol,/home/.zkbnb/ZkBNBVerifier10.sol 1,10 ~/zkbnb-deploy/zkbnb-contract/contracts/ZkBNBVerifier.sol



echo '4-1. get latest block number'
hexNumber=`curl -X POST 'https://data-seed-prebsc-1-s1.binance.org:8545' --header 'Content-Type: application/json' --data-raw '{"jsonrpc":"2.0", "method":"eth_blockNumber", "params": [], "id":1 }' | jq -r '.result'`
blockNumber=`echo $((${hexNumber}))`
echo 'latest block number = ' $blockNumber



echo '4-2. deploy contracts, register and deposit on BSC Testnet'
cd ${DEPLOY_PATH}/zkbnb-contract
cat <<EOF >.env
BSC_TESTNET_PRIVATE_KEY=acbaa269bd7573ff12361be4b97201aef019776ea13384681d4e5ba6a88367d9
EOF
yarn install
npx hardhat --network BSCTestnet run ./scripts/deploy-keccak256/deploy.js
echo 'Recorded latest contract addresses into ~/zkbnb-deploy/zkbnb-contract/info/addresses.json'

npx hardhat --network BSCTestnet run ./scripts/deploy-keccak256/register.js
npx hardhat --network BSCTestnet run ./scripts/deploy-keccak256/deposit.js


echo '5. modify deployed contracts into zkbnb config'
cd ~/zkbnb-deploy/zkbnb/tools/dbinitializer/
cp -r ./contractaddr.yaml.example ./contractaddr.yaml

ZkBNBContractAddr=`cat ~/zkbnb-deploy/zkbnb-contract/info/addresses.json  | jq -r '.zkbnbProxy'`
sed -i "s/ZkBNBProxy: .*/ZkBNBProxy: ${ZkBNBContractAddr}/" ~/zkbnb-deploy/zkbnb/tools/dbinitializer/contractaddr.yaml

GovernanceContractAddr=`cat ~/zkbnb-deploy/zkbnb-contract/info/addresses.json  | jq -r '.governance'`
sed -i "s/Governance: .*/Governance: ${GovernanceContractAddr}/" ~/zkbnb-deploy/zkbnb/tools/dbinitializer/contractaddr.yaml

sed -i "s/BSC_Test_Network_RPC *= .*/BSC_Test_Network_RPC   = \"https\:\/\/data-seed-prebsc-1-s1.binance.org:8545\"/" ~/zkbnb-deploy/zkbnb/tools/dbinitializer/main.go



cd ~/zkbnb-deploy/zkbnb/
make app && make globalRPCProto
cd ~/zkbnb-deploy/zkbnb && go mod tidy



echo "6. init tables on database"
sed -i "s/password=.* dbname/password=ZkBNB@123 dbname/" ~/zkbnb-deploy/zkbnb/tools/dbinitializer/main.go
cd ~/zkbnb-deploy/zkbnb/tools/dbinitializer/
go run .


cd ~/zkbnb-deploy/zkbnb/
make app && make globalRPCProto

sleep 30s

echo "7. run prover"

echo -e "
Name: prover.cronjob
Postgres:
  DataSource: host=127.0.0.1 user=postgres password=ZkBNB@123 dbname=zkbnb port=5432 sslmode=disable

CacheRedis:
  - Host: 127.0.0.1:6379
    Type: node

KeyPath:
  ProvingKeyPath: [/home/.zkbnb/zkbnb1.pk, /home/.zkbnb/zkbnb10.pk]
  VerifyingKeyPath: [/home/.zkbnb/zkbnb1.vk, /home/.zkbnb/zkbnb10.vk]
  KeyTxCounts:    [1, 10]

TreeDB:
  Driver: memorydb
  AssetTreeCacheSize: 512000
" > ~/zkbnb-deploy/zkbnb/service/cronjob/prover/etc/prover.yaml

cd ~/zkbnb-deploy/zkbnb/service/cronjob/prover/
pm2 start --name prover "go run ./prover.go"




echo "8. run witnessGenerator"

echo -e "
Name: witnessGenerator.cronjob

Postgres:
  DataSource: host=127.0.0.1 user=postgres password=ZkBNB@123 dbname=zkbnb port=5432 sslmode=disable

CacheRedis:
  - Host: 127.0.0.1:6379
    Type: node

TreeDB:
  Driver: memorydb
  AssetTreeCacheSize: 512000
" > ~/zkbnb-deploy/zkbnb/service/cronjob/witnessGenerator/etc/witnessGenerator.yaml

cd ~/zkbnb-deploy/zkbnb/service/cronjob/witnessGenerator/
pm2 start --name witnessGenerator "go run ./witnessgenerator.go"






echo "9. run monitor"

echo -e "
Name: monitor.cronjob

Postgres:
  DataSource: host=127.0.0.1 user=postgres password=ZkBNB@123 dbname=zkbnb port=5432 sslmode=disable

CacheRedis:
  - Host: 127.0.0.1:6379
    Type: node

ChainConfig:
  NetworkRPCSysConfigName: "BscTestNetworkRpc"
  #NetworkRPCSysConfigName: "LocalTestNetworkRpc"
  ZkBNBContractAddrSysConfigName: "ZkBNBContract"
  GovernanceContractAddrSysConfigName: "GovernanceContract"
  StartL1BlockHeight: $blockNumber
  PendingBlocksCount: 0
  MaxHandledBlocksCount: 5000

TreeDB:
  Driver: memorydb
  AssetTreeCacheSize: 512000
" > ~/zkbnb-deploy/zkbnb/service/cronjob/monitor/etc/monitor.yaml

cd ~/zkbnb-deploy/zkbnb/service/cronjob/monitor/
pm2 start --name monitor "go run ./monitor.go"



echo "10. run committer"

echo -e "
Name: committer.cronjob

Postgres:
  DataSource: host=127.0.0.1 user=postgres password=ZkBNB@123 dbname=zkbnb port=5432 sslmode=disable

CacheRedis:
  - Host: 127.0.0.1:6379
    Type: node

KeyPath:
  KeyTxCounts: [1, 10]

TreeDB:
  Driver: memorydb
  AssetTreeCacheSize: 512000
" >> ~/zkbnb-deploy/zkbnb/service/cronjob/committer/etc/committer.yaml

cd ~/zkbnb-deploy/zkbnb/service/cronjob/committer/
pm2 start --name committer "go run ./committer.go"




echo "11. run sender"

echo -e "
Name: sender.cronjob

Postgres:
  DataSource: host=127.0.0.1 user=postgres password=ZkBNB@123 dbname=zkbnb port=5432 sslmode=disable

CacheRedis:
  - Host: 127.0.0.1:6379
    Type: node

ChainConfig:
  NetworkRPCSysConfigName: "BscTestNetworkRpc"
  #NetworkRPCSysConfigName: "LocalTestNetworkRpc"
  ZkBNBContractAddrSysConfigName: "ZkBNBContract"
  MaxWaitingTime: 120
  MaxBlockCount: 4
  Sk: "acbaa269bd7573ff12361be4b97201aef019776ea13384681d4e5ba6a88367d9"
  GasLimit: 5000000
  GasPrice: 0

TreeDB:
  Driver: memorydb
  AssetTreeCacheSize: 512000
" > ~/zkbnb-deploy/zkbnb/service/cronjob/sender/etc/sender.yaml

cd ~/zkbnb-deploy/zkbnb/service/cronjob/sender/
pm2 start --name sender "go run ./sender.go"





echo "12. run globalRPC"

echo -e "
Name: global.rpc
ListenOn: 127.0.0.1:8080

Postgres:
  DataSource: host=127.0.0.1 user=postgres password=ZkBNB@123 dbname=zkbnb port=5432 sslmode=disable

CacheRedis:
  - Host: 127.0.0.1:6379
    Type: node

LogConf:
  ServiceName: global.rpc
  Mode: console
  Path: ./log/globalrpc
  StackCooldownMillis: 500

TreeDB:
  Driver: memorydb
  AssetTreeCacheSize: 512000
" > ~/zkbnb-deploy/zkbnb/service/rpc/globalRPC/etc/config.yaml

cd ~/zkbnb-deploy/zkbnb/service/rpc/globalRPC/
pm2 start --name globalRPC "go run ./globalrpc.go"



echo "13. run app"

echo -e "
Name: appService-api
Host: 0.0.0.0
Port: 8888
Postgres:
  DataSource: host=127.0.0.1 user=postgres password=ZkBNB@123 dbname=zkbnb port=5432 sslmode=disable

CacheRedis:
  - Host: 127.0.0.1:6379
    Type: node

GlobalRpc:
  Endpoints:
    - 127.0.0.1:8080

LogConf:
  ServiceName: appservice
  Mode: console
  Path: ./log/appService
  StackCooldownMillis: 500

TreeDB:
  Driver: memorydb
  AssetTreeCacheSize: 512000
  " > ~/zkbnb-deploy/zkbnb/service/api/app/etc/app.yaml

cd ~/zkbnb-deploy/zkbnb/service/api/app
pm2 start --name app "go run ./app.go"
