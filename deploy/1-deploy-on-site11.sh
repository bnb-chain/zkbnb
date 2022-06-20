#!/bin/bash

# only for `root@tf_bsc_qa_bsc_zecrey_site11_ec2`
# config
echo '0. stop old database/redis and docker run new database/redis'
docker rm $ (docker ps -a -q)
docker run -d --name zecreyredis -p 6379:6379 redis
docker run --name postgres -p 5432:5432 -e POSTGRES_PASSWORD=ZecreyProtocolDB@123 -e POSTGRES_USER=postgres -e POSTGRES_DB=zecreylegend -d postgres


echo '1. basic config and git clone repos'
yum install jq -y
npm install pm2 -g
export PATH=$PATH:/usr/local/go/bin/
cd ~
rm -rf ~/zecreylegend-bak && mv ~/zecreylegend ~/zecreylegend-bak
mkdir zecreylegend && cd zecreylegend
git clone --branch qa_testnet_with_keccak256 https://github.com/bnb-chain/zecrey-legend
git clone --branch qa_testnet_with_keccak256 https://github.com/bnb-chain/zecrey-legend-contract
git clone --branch qa_testnet_with_keccak256 https://github.com/bnb-chain/zecrey-crypto


cd ~/zecreylegend
echo '2. start generate zecrey-legend.vk and zecrey-legend.pk'
cd zecrey-crypto && go test ./zecrey-legend/circuit/bn254/solidity -run TestExportSol

cd ~/zecreylegend
cp ./zecrey-crypto/zecrey-legend/circuit/bn254/solidity/zecrey-legend.vk ./zecrey-legend && cp ./zecrey-crypto/zecrey-legend/circuit/bn254/solidity/zecrey-legend.pk ./zecrey-legend


echo '3. upload zecrey-legend.vk and zecrey-legend.pk'
cd ~/zecreylegend
upload_vk_url=`curl --upload-file ./zecrey-legend/zecrey-legend.vk https://transfer.toolsfdg.net/zecrey-legend.vk`
upload_pk_url=`curl --upload-file ./zecrey-legend/zecrey-legend.pk https://transfer.toolsfdg.net/zecrey-legend.pk`
echo "upload_vk_url $upload_vk_url"
echo "upload_pk_url $upload_pk_url"
echo -e "$upload_vk_url" > ~/zecreylegend/upload_vk_url.txt
echo -e "$upload_pk_url" > ~/zecreylegend/upload_pk_url.txt


echo '4. start verify_parse for ZecreyVerifier'
cd ~/zecreylegend
python3 ./zecrey-legend/deploy/verify_parse.py ./zecrey-crypto/zecrey-legend/circuit/bn254/solidity/ZecreyVerifier.sol ./zecrey-legend-contract/contracts/ZecreyVerifier.sol


echo '5. deploy contracts, register and deposit on BSC Testnet'
cd ~/zecreylegend
cd ./zecrey-legend-contract && npm install
npx hardhat --network testnet run ./scripts/deploy-keccak256/deploy.js
echo 'Recorded latest contract addresses into ~/zecreylegend/zecrey-legend-contract/info/addresses.json'

npx hardhat --network testnet run ./scripts/deploy-keccak256/register.js
npx hardhat --network testnet run ./scripts/deploy-keccak256/deposit.js


echo "6. update contracts in '~/zecreylegend/zecrey-legend/common/model/init/sysconfig.go'"
cd ~/zecreylegend
cd ./zecrey-legend

ZecreyLegendContractAddr=`cat ~/zecreylegend/zecrey-legend-contract/info/addresses.json  | jq -r '.zecreyLegendProxy'`
sed -i "s/ZecreyLegendContractAddr = .*/ZecreyLegendContractAddr = \"${ZecreyLegendContractAddr}\"/" ~/zecreylegend/zecrey-legend/common/model/init/sysconfig.go

GovernanceContractAddr=`cat ~/zecreylegend/zecrey-legend-contract/info/addresses.json  | jq -r '.governance'`
sed -i "s/GovernanceContractAddr   = .*/GovernanceContractAddr   = \"${GovernanceContractAddr}\"/" ~/zecreylegend/zecrey-legend/common/model/init/sysconfig.go

VerifierContractAddr=`cat ~/zecreylegend/zecrey-legend-contract/info/addresses.json  | jq -r '.verifierProxy'`
sed -i "s/VerifierContractAddr     = .*/VerifierContractAddr     = \"${VerifierContractAddr}\"/" ~/zecreylegend/zecrey-legend/common/model/init/sysconfig.go


echo "7. init tables on database"
cd ~/zecreylegend
cd ./zecrey-legend
cd common/model/init/
go run .

echo "8. run governanceMonitor"
pm2 delete all
cd ~/zecreylegend/zecrey-legend
pm2 start --name governanceMonitor "go run service/cronjob/governanceMonitor/governanceMonitor.go"

echo "9. run blockMonitor"
cd ~/zecreylegend/zecrey-legend
pm2 start --name blockMonitor "go run service/cronjob/blockMonitor/blockmonitor.go"

echo "10. run mempoolMonitor"
cd ~/zecreylegend/zecrey-legend
pm2 start --name mempoolMonitor "go run service/cronjob/mempoolMonitor/mempoolmonitor.go"

echo "11. run l2BlockMonitor"
cd ~/zecreylegend/zecrey-legend
pm2 start --name l2BlockMonitor "go run service/cronjob/l2BlockMonitor/l2blockmonitor.go"

echo "12. run globalRPC"
cd ~/zecreylegend/zecrey-legend
pm2 start --name globalRPC "go run service/rpc/globalRPC/globalrpc.go"


