#!/usr/bin/env bash
WORKDIR=$(cd `dirname $0`; pwd)
export BASEDIR=$WORKDIR/..
export CONFIG_PATH=$WORKDIR/../configs
export PATH=$PATH:/usr/local/go/bin:/usr/local/go/bin:/root/go/bin:/usr/local/bin

mkdir -p $CONFIG_PATH
set -a
. $WORKDIR/docker.env
set +a


function prepareConfigs() {
if [ -z $1 ] ; then
    echo "invalid block height"
    exit 1
fi

BLOCK_NUMBER=$1

echo -e "
Name: prover
Postgres:
  DataSource: host=database user=$DATABASE_USER password=$DATABASE_PASS dbname=$DATABASE_NAME port=5432 sslmode=disable

CacheRedis:
  - Host: redis:6379
    Type: node

KeyPath:
  ProvingKeyPath: [/server/.zkbnb/zkbnb1.pk]
  VerifyingKeyPath: [/server/.zkbnb/zkbnb1.vk]

LogConf:
  ServiceName: prover
  Mode: console
  Encoding: plain
  StackCooldownMillis: 500
  Level: info

BlockConfig:
  OptionalBlockSizes: [1]
" > ${CONFIG_PATH}/prover.yaml

echo -e "
Name: witness

Postgres:
  DataSource: host=database user=$DATABASE_USER password=$DATABASE_PASS dbname=$DATABASE_NAME port=5432 sslmode=disable

CacheRedis:
  - Host: redis:6379
    Type: node

LogConf:
  ServiceName: witness
  Mode: console
  Encoding: plain
  StackCooldownMillis: 500
  Level: info

TreeDB:
  Driver: memorydb
  AssetTreeCacheSize: 512000
" > ${CONFIG_PATH}/witness.yaml

echo -e "
Name: monitor

Postgres:
  DataSource: host=database user=$DATABASE_USER password=$DATABASE_PASS dbname=$DATABASE_NAME port=5432 sslmode=disable

CacheRedis:
  - Host: redis:6379
    Type: node

LogConf:
  ServiceName: monitor
  Mode: console
  Encoding: plain
  StackCooldownMillis: 500
  Level: info

ChainConfig:
  NetworkRPCSysConfigName: \"BscTestNetworkRpc\"
  #NetworkRPCSysConfigName: \"LocalTestNetworkRpc\"
  StartL1BlockHeight: $BLOCK_NUMBER
  ConfirmBlocksCount: 0
  MaxHandledBlocksCount: 5000
  KeptHistoryBlocksCount: 100000
" > ${CONFIG_PATH}/monitor.yaml

echo -e "
Name: committer

Postgres:
  DataSource: host=database user=$DATABASE_USER password=$DATABASE_PASS dbname=$DATABASE_NAME port=5432 sslmode=disable

CacheRedis:
  - Host: redis:6379
    Type: node

BlockConfig:
  OptionalBlockSizes: [1]

TreeDB:
  Driver: memorydb
  AssetTreeCacheSize: 512000
" > ${CONFIG_PATH}/committer.yaml

echo -e "
Name: sender

Postgres:
  DataSource: host=database user=$DATABASE_USER password=$DATABASE_PASS dbname=$DATABASE_NAME port=5432 sslmode=disable

CacheRedis:
  - Host: redis:6379
    Type: node

LogConf:
  ServiceName: sender
  Mode: console
  Encoding: plain
  StackCooldownMillis: 500
  Level: info

ChainConfig:
  NetworkRPCSysConfigName: \"BscTestNetworkRpc\"
  #NetworkRPCSysConfigName: \"LocalTestNetworkRpc\"
  ConfirmBlocksCount: 0
  MaxWaitingTime: 120
  MaxBlockCount: 4
  Sk: \"$SK\"
  GasLimit: 5000000
  GasPrice: 0

" > ${CONFIG_PATH}/sender.yaml

echo -e "
Name: api-server
Host: 0.0.0.0
Port: 8888

TxPool:
  MaxPendingTxCount: 10000

Postgres:
  DataSource: host=database user=$DATABASE_USER password=$DATABASE_PASS dbname=$DATABASE_NAME port=5432 sslmode=disable
  MaxConn: 100
  MaxIdle: 10

CacheRedis:
  - Host: redis:6379
    Type: node

LogConf:
  ServiceName: api-server
  Mode: console
  Path: ./log/api-server
  Encoding: plain
  StackCooldownMillis: 500
  Level: info

CoinMarketCap:
  Url: $CMC_URL
  Token: $CMC_TOKEN

MemCache:
  AccountExpiration: 200
  AssetExpiration:   600
  BlockExpiration:   400
  TxExpiration:      400
  PriceExpiration:   3600000
  MaxCounterNum:     100000
  MaxKeyNum:         10000

" > ${CONFIG_PATH}/apiserver.yaml

}

function up() {
    cd $WORKDIR
    docker rm -f $(docker ps -aq)
    docker-compose up -d
}

function down() {
    cd $WORKDIR
    docker-compose down
}

CMD=$1
case ${CMD} in
up)
    prepareConfigs $2
    up
    ;;
down)
    down
    ;;
*)
    echo "Usage: docker-compose.sh up \$block_number | down"
    ;;
esac
