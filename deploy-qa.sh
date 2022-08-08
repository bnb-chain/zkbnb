#!/bin/bash

# config
# GOBIN=/usr/local/bin/ go install  github.com/zeromicro/go-zero/tools/goctl@latest

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
rm -rf ~/zkbas-deploy-bak && mv ~/zkbas-deploy ~/zkbas-deploy-bak
mkdir zkbas-deploy && cd zkbas-deploy
git clone --branch develop  https://github.com/bnb-chain/zkbas-contract.git
git clone --branch develop https://github.com/bnb-chain/zkbas-crypto.git

# mv /home/ec2-user/zkbas ~/zkbas-deploy
branch=$1
git clone --branch $branch https://github.com/bnb-chain/zkbas.git

echo "new crypto env"
echo '2. start generate zkbas.vk and zkbas.pk'
cd ~/zkbas-deploy
cd zkbas-crypto && go test ./legend/circuit/bn254/solidity -timeout 99999s -run TestExportSol
cd ~/zkbas-deploy
sudo mkdir /home/.zkbas
cp -r ./zkbas-crypto/legend/circuit/bn254/solidity/* /home/.zkbas


echo '3. start verify_parse for ZkbasVerifier'
cd ~/zkbas-deploy/zkbas/service/cronjob/prover/
python3 verifier_parse.py /home/.zkbas/ZkbasVerifier1.sol,/home/.zkbas/ZkbasVerifier10.sol 1,10 ~/zkbas-deploy/zkbas-contract/contracts/ZkbasVerifier.sol



echo '4-1. get latest block number'
hexNumber=`curl -X POST 'https://data-seed-prebsc-1-s1.binance.org:8545' --header 'Content-Type: application/json' --data-raw '{"jsonrpc":"2.0", "method":"eth_blockNumber", "params": [], "id":1 }' | jq -r '.result'`
blockNumber=`echo $((${hexNumber}))`
echo 'latest block number = ' $blockNumber



echo '4-2. deploy contracts, register and deposit on BSC Testnet'
cd ~/zkbas-deploy
cd ./zkbas-contract && sudo npm install
npx hardhat --network BSCTestnet run ./scripts/deploy-keccak256/deploy.js
echo 'Recorded latest contract addresses into ~/zkbas-deploy/zkbas-contract/info/addresses.json'

npx hardhat --network BSCTestnet run ./scripts/deploy-keccak256/register.js
npx hardhat --network BSCTestnet run ./scripts/deploy-keccak256/deposit.js


echo '5. modify deployed contracts into zkbas config'
cd ~/zkbas-deploy/zkbas/common/model/init/
cp -r ./contractaddr.yaml.example ./contractaddr.yaml

ZkbasContractAddr=`cat ~/zkbas-deploy/zkbas-contract/info/addresses.json  | jq -r '.zkbasProxy'`
sed -i "s/ZkbasProxy: .*/ZkbasProxy: ${ZkbasContractAddr}/" ~/zkbas-deploy/zkbas/common/model/init/contractaddr.yaml

GovernanceContractAddr=`cat ~/zkbas-deploy/zkbas-contract/info/addresses.json  | jq -r '.governance'`
sed -i "s/Governance: .*/Governance: ${GovernanceContractAddr}/" ~/zkbas-deploy/zkbas/common/model/init/contractaddr.yaml

sed -i "s/BSC_Test_Network_RPC *= .*/BSC_Test_Network_RPC   = \"https\:\/\/data-seed-prebsc-1-s1.binance.org:8545\"/" ~/zkbas-deploy/zkbas/common/model/init/init.go



cd ~/zkbas-deploy/zkbas/
make app && make globalRPCProto
cd ~/zkbas-deploy/zkbas && go mod tidy



echo "6. init tables on database"
sed -i "s/password=.* dbname/password=Zkbas@123 dbname/" ~/zkbas-deploy/zkbas/common/model/basic/connection.go
cd ~/zkbas-deploy/zkbas/common/model/init/
go run .


cd ~/zkbas-deploy/zkbas/
make app && make globalRPCProto


echo "7. run prover"

echo -e "
Name: prover.cronjob
Postgres:
  DataSource: host=127.0.0.1 user=postgres password=Zkbas@123 dbname=zkbas port=5432 sslmode=disable

CacheRedis:
  - Host: 127.0.0.1:6379
    Type: node

KeyPath:
  ProvingKeyPath: [/home/.zkbas/zkbas1.pk, /home/.zkbas/zkbas10.pk]
  VerifyingKeyPath: [/home/.zkbas/zkbas1.vk, /home/.zkbas/zkbas10.vk]
  KeyTxCounts:    [1, 10]

TreeDB:
  Driver: memorydb
" > ~/zkbas-deploy/zkbas/service/cronjob/prover/etc/prover.yaml

cd ~/zkbas-deploy/zkbas/service/cronjob/prover/
pm2 start --name prover "go run ./prover.go"




echo "8. run witnessGenerator"

echo -e "
Name: witnessGenerator.cronjob

Postgres:
  DataSource: host=127.0.0.1 user=postgres password=Zkbas@123 dbname=zkbas port=5432 sslmode=disable

CacheRedis:
  - Host: 127.0.0.1:6379
    Type: node

TreeDB:
  Driver: memorydb
" > ~/zkbas-deploy/zkbas/service/cronjob/witnessGenerator/etc/witnessGenerator.yaml

cd ~/zkbas-deploy/zkbas/service/cronjob/witnessGenerator/
pm2 start --name witnessGenerator "go run ./witnessgenerator.go"






echo "9. run monitor"

echo -e "
Name: monitor.cronjob

Postgres:
  DataSource: host=127.0.0.1 user=postgres password=Zkbas@123 dbname=zkbas port=5432 sslmode=disable

CacheRedis:
  - Host: 127.0.0.1:6379
    Type: node

ChainConfig:
  NetworkRPCSysConfigName: "BscTestNetworkRpc"
  #NetworkRPCSysConfigName: "LocalTestNetworkRpc"
  ZkbasContractAddrSysConfigName: "ZkbasContract"
  GovernanceContractAddrSysConfigName: "GovernanceContract"
  StartL1BlockHeight: $blockNumber
  PendingBlocksCount: 0
  MaxHandledBlocksCount: 5000

TreeDB:
  Driver: memorydb
" > ~/zkbas-deploy/zkbas/service/cronjob/monitor/etc/monitor.yaml

cd ~/zkbas-deploy/zkbas/service/cronjob/monitor/
pm2 start --name monitor "go run ./monitor.go"



echo "10. run committer"

echo -e "
Name: committer.cronjob

Postgres:
  DataSource: host=127.0.0.1 user=postgres password=Zkbas@123 dbname=zkbas port=5432 sslmode=disable

CacheRedis:
  - Host: 127.0.0.1:6379
    Type: node

KeyPath:
  KeyTxCounts: [1, 10]

TreeDB:
  Driver: memorydb
" >> ~/zkbas-deploy/zkbas/service/cronjob/committer/etc/committer.yaml

cd ~/zkbas-deploy/zkbas/service/cronjob/committer/
pm2 start --name committer "go run ./committer.go"




echo "11. run sender"

echo -e "
Name: sender.cronjob

Postgres:
  DataSource: host=127.0.0.1 user=postgres password=Zkbas@123 dbname=zkbas port=5432 sslmode=disable

CacheRedis:
  - Host: 127.0.0.1:6379
    Type: node

ChainConfig:
  NetworkRPCSysConfigName: "BscTestNetworkRpc"
  #NetworkRPCSysConfigName: "LocalTestNetworkRpc"
  ZkbasContractAddrSysConfigName: "ZkbasContract"
  MaxWaitingTime: 120
  MaxBlockCount: 4
  Sk: "acbaa269bd7573ff12361be4b97201aef019776ea13384681d4e5ba6a88367d9"
  GasLimit: 5000000
  L1ChainId: \"97\"

TreeDB:
  Driver: memorydb
" > ~/zkbas-deploy/zkbas/service/cronjob/sender/etc/sender.yaml

cd ~/zkbas-deploy/zkbas/service/cronjob/sender/
pm2 start --name sender "go run ./sender.go"





echo "12. run globalRPC"

echo -e "
Name: global.rpc
ListenOn: 127.0.0.1:8080

Postgres:
  DataSource: host=127.0.0.1 user=postgres password=Zkbas@123 dbname=zkbas port=5432 sslmode=disable

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
" > ~/zkbas-deploy/zkbas/service/rpc/globalRPC/etc/config.yaml

cd ~/zkbas-deploy/zkbas/service/rpc/globalRPC/
pm2 start --name globalRPC "go run ./globalrpc.go"



echo "13. run app"

echo -e "
Name: appService-api
Host: 0.0.0.0
Port: 8888
Postgres:
  DataSource: host=127.0.0.1 user=postgres password=Zkbas@123 dbname=zkbas port=5432 sslmode=disable

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
  " > ~/zkbas-deploy/zkbas/service/api/app/etc/app.yaml

cd ~/zkbas-deploy/zkbas/service/api/app
pm2 start --name app "go run ./app.go"
