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
  MasterDataSource: host=database user=$DATABASE_USER password=$DATABASE_PASS dbname=$DATABASE_NAME port=5432 sslmode=disable
  SlaveDataSource: host=database user=$DATABASE_USER password=$DATABASE_PASS dbname=$DATABASE_NAME port=5432 sslmode=disable

CacheRedis:
  - Host: redis:6379
    Type: node

KeyPath: [/server/.zkbnb/zkbnb1]

LogConf:
  ServiceName: prover
  Mode: console
  Encoding: plain
  StackCooldownMillis: 500
  Level: info

BlockConfig:
  OptionalBlockSizes: [1]

TreeDB:
  Driver: memorydb
  AssetTreeCacheSize: 512000
" > ${CONFIG_PATH}/prover.yaml

echo -e "
Name: witness

Postgres:
  MasterDataSource: host=database user=$DATABASE_USER password=$DATABASE_PASS dbname=$DATABASE_NAME port=5432 sslmode=disable
  SlaveDataSource: host=database user=$DATABASE_USER password=$DATABASE_PASS dbname=$DATABASE_NAME port=5432 sslmode=disable

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
  MasterDataSource: host=database user=$DATABASE_USER password=$DATABASE_PASS dbname=$DATABASE_NAME port=5432 sslmode=disable
  SlaveDataSource: host=database user=$DATABASE_USER password=$DATABASE_PASS dbname=$DATABASE_NAME port=5432 sslmode=disable

CacheRedis:
  - Host: redis:6379
    Type: node

LogConf:
  ServiceName: monitor
  Mode: console
  Encoding: plain
  StackCooldownMillis: 500
  Level: info

AccountCacheSize: 100000

ChainConfig:
  NetworkRPCSysConfigName: \"BscTestNetworkRpc\"
  StartL1BlockHeight: $BLOCK_NUMBER
  ConfirmBlocksCount: 0
  MaxHandledBlocksCount: 5000
  KeptHistoryBlocksCount: 100000

TreeDB:
  Driver: memorydb
  AssetTreeCacheSize: 512000
" > ${CONFIG_PATH}/monitor.yaml

echo -e "
Name: committer

Postgres:
  MasterDataSource: host=database user=$DATABASE_USER password=$DATABASE_PASS dbname=$DATABASE_NAME port=5432 sslmode=disable
  SlaveDataSource: host=database user=$DATABASE_USER password=$DATABASE_PASS dbname=$DATABASE_NAME port=5432 sslmode=disable

CacheRedis:
  - Host: redis:6379
    Type: node

BlockConfig:
  OptionalBlockSizes: [1]

IpfsUrl:
  10.23.23.40:5001

TreeDB:
  Driver: memorydb
  AssetTreeCacheSize: 512000
" > ${CONFIG_PATH}/committer.yaml

echo -e "
Name: sender

Postgres:
  MasterDataSource: host=database user=$DATABASE_USER password=$DATABASE_PASS dbname=$DATABASE_NAME port=5432 sslmode=disable
  SlaveDataSource: host=database user=$DATABASE_USER password=$DATABASE_PASS dbname=$DATABASE_NAME port=5432 sslmode=disable

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
  ConfirmBlocksCount: 0
  MaxWaitingTime: 120
  MaxBlockCount: 4
  GasLimit: 5000000
  GasPrice: 0

Apollo:
  AppID:             zkbnb-cloud
  Cluster:           testnet
  ApolloIp:          http://internal-tf-cm-test-apollo-config-alb-2119591301.ap-northeast-1.elb.amazonaws.com:9028
  Namespace:         application
  IsBackupConfig:    true

AuthConfig:
  CommitBlockSk: \"${COMMIT_BLOCK_PRIVATE_KEY}\"
  VerifyBlockSk: \"${VERIFY_BLOCK_PRIVATE_KEY}\"

KMSConfig:
  CommitKeyId: 5637f226-2438-43bd-bc35-837d27845b6b
  VerifyKeyId: 0b71f7b5-0d8f-4181-a904-7e29ae7411ee

TreeDB:
  Driver: memorydb
  AssetTreeCacheSize: 512000
" > ${CONFIG_PATH}/sender.yaml

echo -e "
Name: api-server
Host: 0.0.0.0
Port: 8888

TxPool:
  MaxPendingTxCount: 10000

Postgres:
  MasterDataSource: host=database user=$DATABASE_USER password=$DATABASE_PASS dbname=$DATABASE_NAME port=5432 sslmode=disable
  SlaveDataSource: host=database user=$DATABASE_USER password=$DATABASE_PASS dbname=$DATABASE_NAME port=5432 sslmode=disable
  MaxConn: 1000
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

Apollo:
  AppID:             zkbnb-cloud
  Cluster:           testnet
  ApolloIp:          http://internal-tf-cm-test-apollo-config-alb-2119591301.ap-northeast-1.elb.amazonaws.com:9028
  Namespace:         application
  IsBackupConfig:    true

IpfsUrl:
  10.23.23.40:5001

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
