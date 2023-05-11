go run ./cmd/zkbnb/main.go tree recovery  --service committer --batch 1000 --config ./tools/recovery/etc/config.yaml

go run ./cmd/zkbnb/main.go tree recovery  --service witness --batch 1000  --config ./tools/recovery/etc/config.yaml

go run ./cmd/zkbnb/main.go treedb query  --service witness --height 5 --accountIndexList [1,2,3,4,146] --config ./tools/query/etc/config.yaml

go run ./cmd/zkbnb/main.go treedb query  --service committer --height 54 --config ./tools/query/etc/config.yaml

go run ./cmd/zkbnb/main.go revertblock  --height 5 --config ./tools/revertblock/etc/config.yaml

go run ./cmd/zkbnb/main.go estimategas --config ./tools/estimategas/etc/config.yaml --fromHeight 196 --toHeight 238 --maxBlockCount 43  --sendToL1 1


go run ./cmd/zkbnb/main.go rollback --config ./tools/rollback/etc/config.yaml --height 5

go run ./cmd/zkbnb/main.go rollbackwitnesssmt --height 5 --config ./tools/rollbackwitnesssmt/etc/config.yaml

redis-cli -h 127.0.0.1 -p 6666 flushdb



#ec2
./zkbnb treedb query  --service witness --height 5
./zkbnb revertblock  --height 220

./zkbnb  rollbackwitnesssmt  --height 220

./zkbnb tree recovery  --service committer
./zkbnb tree recovery  --service witness


./zkbnb rollback --height 493