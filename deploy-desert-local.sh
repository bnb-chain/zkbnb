#!/bin/bash

# Preparation: Install following tools when you first run this script!!!
# GOBIN=/usr/local/bin/ go install  github.com/zeromicro/go-zero/tools/goctl@1.5.1
# yum install jq -y
# npm install pm2 -g
# You should install nodejs above v14
# sh deploy-local.sh new  // append the new parameter to generate pk and vk data when you first run this script.
##
# Attention: Set the following variables to the right one before running!!!
ZkBNB_REPO_PATH=$(cd `dirname $0`; pwd)

export PATH=$PATH:/usr/local/go/bin:/usr/local/go/bin:/root/go/bin
echo 'start install'
ZKBNB_CONTAINERS=$(docker ps -a |grep zkbnb-desert|awk '{print $1}')
[[ -z "${ZKBNB_CONTAINERS}" ]] || docker rm -f ${ZKBNB_CONTAINERS}
docker run -d --name zkbnb-desert-postgres -p 5432:5432 \
  -e PGDATA=/var/lib/postgresql/pgdata  \
  -e POSTGRES_PASSWORD=ZkBNB@123 \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_DB=zkbnb_desert postgres

export PATH=$PATH:/usr/local/go/bin/


echo 'install dependency'

make api-server
cd ${ZkBNB_REPO_PATH} && go mod tidy

sleep 10s


echo "
Name: desertexit

Postgres:
  MasterDataSource: host=127.0.0.1 user=postgres password=ZkBNB@123 dbname=zkbnb_desert port=5432 sslmode=disable
  LogLevel: 4

ChainConfig:
  ConfirmBlocksCount: 0
  MaxWaitingTime: 120
  MaxHandledBlocksCount: 5000
  MaxCancelOutstandingDepositCount: 100
  KeptHistoryBlocksCount: 100000
  GasLimit: 5000000

  StartL1BlockHeight: 1
  BscTestNetRpc: http://127.0.0.1:8545
  ZkBnbContractAddress: 0xF170394283cDf43C5A0900Ef6A3af2886108eFa3
  GovernanceContractAddress: 0xE48fC034056eac15F9063b502d08f5968A90E694

TreeDB:
  Driver: memorydb
  AssetTreeCacheSize: 1000000

CacheConfig:
  AccountCacheSize: 5000000
  NftCacheSize: 5000000
  MemCacheSize: 100000

KeyPath: /Users/user/.zkbnb/zkbnb.desert1

  " > ${ZkBNB_REPO_PATH}/tools/desertexit/etc/config.yaml

echo 'end install'
