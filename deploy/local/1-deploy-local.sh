#!/bin/bash

# if first deploying, should install dependencies before
# yum install -y docker jq gcc && systemctl start docker
# curl -fsSL https://rpm.nodesource.com/setup_16.x | sudo bash - && sudo yum install -y nodejs
# npm install pm2 -g


# only for `local`
# config
echo '0. stop old database/redis and start new database/redis'
pm2 delete all
docker kill $(docker ps -q)
docker rm $(docker ps -a -q)

# start redis service
docker run -d --name zecreyredis -p 6379:6379 redis
# start postgres service
docker run --name postgres -p 5432:5432 -e POSTGRES_PASSWORD=ZecreyProtocolDB@123 -e POSTGRES_USER=postgres -e POSTGRES_DB=zecreylegend -d postgres


echo '1. basic config and git clone repos'
export PATH=$PATH:/usr/local/go/bin/
cd ~
rm -rf ~/zkbas-deploy-bak && mv ~/zkbas-deploy ~/zkbas-deploy-bak
mkdir zkbas-deploy && cd zkbas-deploy
git clone --branch develop https://github.com/bnb-chain/zkbas
git clone --branch develop https://github.com/bnb-chain/zkbas-contract
git clone --branch develop https://github.com/bnb-chain/zkbas-crypto


cd ~/zkbas-deploy
echo '2. start generate zkbas.vk and zkbas.pk'
cd zkbas-crypto && go test ./legend/circuit/bn254/solidity -run TestExportSol

cd ~/zkbas-deploy
cp ./zkbas-crypto/legend/circuit/bn254/solidity/zkbas.vk ./zkbas && cp ./zkbas-crypto/legend/circuit/bn254/solidity/zkbas.pk ./zkbas
cp ./zkbas-crypto/legend/circuit/bn254/solidity/zkbas.vk /tmp && cp ./zkbas-crypto/legend/circuit/bn254/solidity/zkbas.pk /tmp


echo '3. start verify_parse for ZecreyVerifier'
cd ~/zkbas-deploy
python3 ./zkbas/deploy/verify_parse.py ./zkbas-crypto/legend/circuit/bn254/solidity/ZecreyVerifier.sol ./zkbas-contract/contracts/ZecreyVerifier.sol


echo '4. deploy contracts, register and deposit on BSC Testnet'
cd ~/zkbas-deploy
cd ./zkbas-contract && npm install
npx hardhat --network testnet run ./scripts/deploy-keccak256/deploy.js
echo 'Recorded latest contract addresses into ~/zkbas-deploy/zkbas-contract/info/addresses.json'

npx hardhat --network testnet run ./scripts/deploy-keccak256/register.js
npx hardhat --network testnet run ./scripts/deploy-keccak256/deposit.js


echo "5. update contracts in '~/zkbas-deploy/zkbas/common/model/init/sysconfig.go'"
cd ~/zkbas-deploy/zkbas
ZkbasContractAddr=`cat ~/zkbas-deploy/zkbas-contract/info/addresses.json  | jq -r '.zecreyLegendProxy'`
sed -i "s/ZkbasContractAddr      = .*/ZkbasContractAddr      = \"${ZkbasContractAddr}\"/" ~/zkbas-deploy/zkbas/common/model/init/sysconfig.go
GovernanceContractAddr=`cat ~/zkbas-deploy/zkbas-contract/info/addresses.json  | jq -r '.governance'`
sed -i "s/GovernanceContractAddr = .*/GovernanceContractAddr = \"${GovernanceContractAddr}\"/" ~/zkbas-deploy/zkbas/common/model/init/sysconfig.go
VerifierContractAddr=`cat ~/zkbas-deploy/zkbas-contract/info/addresses.json  | jq -r '.verifierProxy'`
sed -i "s/VerifierContractAddr   = .*/VerifierContractAddr   = \"${VerifierContractAddr}\"/" ~/zkbas-deploy/zkbas/common/model/init/sysconfig.go


echo "6. init tables on database"
cd ~/zkbas-deploy
cd ./zkbas
cd common/model/init/
go run .

echo "7. run governanceMonitor"
cd ~/zkbas-deploy/zkbas
pm2 start --name governanceMonitor "go run service/cronjob/governanceMonitor/governancemonitor.go"

echo "8. run blockMonitor"
cd ~/zkbas-deploy/zkbas
pm2 start --name blockMonitor "go run service/cronjob/blockMonitor/blockmonitor.go"

echo "9. run mempoolMonitor"
cd ~/zkbas-deploy/zkbas
pm2 start --name mempoolMonitor "go run service/cronjob/mempoolMonitor/mempoolmonitor.go"

echo "10. run l2BlockMonitor"
cd ~/zkbas-deploy/zkbas
pm2 start --name l2BlockMonitor "go run service/cronjob/l2BlockMonitor/l2blockmonitor.go"

echo "11. run globalRPC"
cd ~/zkbas-deploy/zkbas
pm2 start --name globalRPC "go run service/rpc/globalRPC/globalrpc.go"

echo "12. run committer"
cd ~/zkbas-deploy/zkbas
pm2 start --name committer "go run service/cronjob/committer/committer.go"


echo "13. run proverHub"
cd ~/zkbas-deploy/zkbas
pm2 start --name proverHub "go run service/rpc/proverHub/proverhub.go"

echo "14. run proverClient"
cd ~/zkbas-deploy/zkbas
pm2 start --name proverClient "go run service/cronjob/proverClient/proverclient.go"

echo "15. run sender"
cd ~/zkbas-deploy/zkbas
pm2 start --name sender "go run service/cronjob/sender/sender.go"


echo "16. run explorer"
cd ~/zkbas-deploy/zkbas
pm2 start --name explorer "go run service/api/explorer/explorer.go"





