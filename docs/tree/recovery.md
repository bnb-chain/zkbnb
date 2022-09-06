## Recovery

Due to the introduction of the persistent SparseMerkleTree structure and support for persisting data to local (leveldb) and remote storage spaces (kvrocks, redis).

When stored tree data is accidentally lost, this tool can help recover tree data.
#### Usage

1. Prepare a config.yaml to set the RDB, Redis, and target tree sources you want to restore.
```yaml
Postgres:
  DataSource: host=127.0.0.1 user=postgres password=ZkBNB@123 dbname=zkbnb port=5432 sslmode=disable

CacheRedis:
  - Host: 127.0.0.1:6379
    # Pass: myredis
    Type: node

TreeDB:
  Driver: leveldb
  LevelDBOption:
    File: /tmp/test
```
2. execute the tool
```sh
recovery -f ${config} -height 300 -service committer
```
