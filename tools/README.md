tree recovery  --service committer --batch 1000 --height 4 --config ./tools/recovery/etc/config.yaml

tree recovery  --service witness --batch 1000 --height 26 --config ./tools/recovery/etc/config.yaml

go run ./cmd/zkbnb/main.go treedb query  --service witness --height 10 --accountIndexList [0,1,2,3,4] --config ./tools/query/etc/config.yaml

go run ./cmd/zkbnb/main.go treedb query  --service committer --height 54 --config ./tools/query/etc/config.yaml

revertblock --config ./tools/revertblock/etc/config.yaml --height 2

estimategas --config ./tools/estimategas/etc/config.yaml --fromHeight 196 --toHeight 238 --maxBlockCount 43  --sendToL1 1


rollback --config ./tools/rollback/etc/config.yaml --height 5

rollbackwitnesssmt --config ./tools/rollbackwitnesssmt/etc/config.yaml --height 2

redis-cli -h 127.0.0.1 -p 6666 flushdb
redis-cli -h 10.23.3.107 -p 6666 flushdb