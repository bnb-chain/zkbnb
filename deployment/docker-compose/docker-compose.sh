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
  ProvingKeyPath: [/server/.zkbas/zkbas1.pk, /server/.zkbas/zkbas10.pk]
  VerifyingKeyPath: [/server/.zkbas/zkbas1.vk, /server/.zkbas/zkbas10.vk]

BlockConfig:
  OptionalBlockSizes: [1, 10]
" > ${CONFIG_PATH}/prover.yaml

echo -e "
Name: witness

Postgres:
  DataSource: host=database user=$DATABASE_USER password=$DATABASE_PASS dbname=$DATABASE_NAME port=5432 sslmode=disable

CacheRedis:
  - Host: redis:6379
    Type: node

TreeDB:
  Driver: memorydb
" > ${CONFIG_PATH}/witness.yaml

echo -e "
Name: monitor

Postgres:
  DataSource: host=database user=$DATABASE_USER password=$DATABASE_PASS dbname=$DATABASE_NAME port=5432 sslmode=disable

CacheRedis:
  - Host: redis:6379
    Type: node

ChainConfig:
  NetworkRPCSysConfigName: \"BscTestNetworkRpc\"
  #NetworkRPCSysConfigName: \"LocalTestNetworkRpc\"
  StartL1BlockHeight: $BLOCK_NUMBER
  ConfirmBlocksCount: 0
  MaxHandledBlocksCount: 5000
" > ${CONFIG_PATH}/monitor.yaml

echo -e "
Name: committer

Postgres:
  DataSource: host=database user=$DATABASE_USER password=$DATABASE_PASS dbname=$DATABASE_NAME port=5432 sslmode=disable

CacheRedis:
  - Host: redis:6379
    Type: node

BlockConfig:
  OptionalBlockSizes: [1, 10]

TreeDB:
  Driver: memorydb
" > ${CONFIG_PATH}/committer.yaml

echo -e "
Name: sender

Postgres:
  DataSource: host=database user=$DATABASE_USER password=$DATABASE_PASS dbname=$DATABASE_NAME port=5432 sslmode=disable

CacheRedis:
  - Host: redis:6379
    Type: node

ChainConfig:
  NetworkRPCSysConfigName: \"BscTestNetworkRpc\"
  #NetworkRPCSysConfigName: \"LocalTestNetworkRpc\"
  ConfirmBlocksCount: 0
  MaxWaitingTime: 120
  MaxBlockCount: 4
  Sk: \"$SK\"
  GasLimit: 5000000

" > ${CONFIG_PATH}/sender.yaml

echo -e "
Name: api-server
Host: 0.0.0.0
Port: 8888

Prometheus:
  Host: 0.0.0.0
  Port: 9091
  Path: /metrics

Postgres:
  DataSource: host=database user=$DATABASE_USER password=$DATABASE_PASS dbname=$DATABASE_NAME port=5432 sslmode=disable

CacheRedis:
  - Host: redis:6379
    Type: node

LogConf:
  ServiceName: api-server
  Mode: console
  Path: ./log/api-server
  StackCooldownMillis: 500
  Level: error

CoinMarketCap:
  Url: $CMC_URL
  Token: $CMC_TOKEN

MemCache:
  AccountExpiration: 200
  AssetExpiration:   600
  BlockExpiration:   400
  TxExpiration:      400
  PriceExpiration:   200

" > ${CONFIG_PATH}/apiserver.yaml

}

function up() {
    cd $WORKDIR
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