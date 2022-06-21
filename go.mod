module github.com/bnb-chain/zkbas

go 1.16

require (
	github.com/zeromicro/go-zero v1.3.3
	gorm.io/gorm v1.23.4
)

require (
	github.com/bnb-chain/zkbas-crypto v0.0.0-20220620132000-a56efca5895f
	github.com/bnb-chain/zkbas-eth-rpc v0.0.0-20220620133434-3d27b92fa67d
	github.com/consensys/gnark v0.7.0
	github.com/consensys/gnark-crypto v0.7.0
	github.com/eko/gocache/v2 v2.3.1
	github.com/ethereum/go-ethereum v1.10.17
	github.com/go-redis/redis/v8 v8.11.5
	github.com/google/uuid v1.3.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/robfig/cron/v3 v3.0.1
	github.com/zeromicro/go-zero/tools/goctl v1.3.8
	google.golang.org/grpc v1.46.0
	google.golang.org/protobuf v1.28.0
	gorm.io/driver/postgres v1.3.6
	k8s.io/client-go v0.24.1 // indirect
)

//replace github.com/bnb-chain/zkbas-crypto => ../zkbas-crypto
//
//replace github.com/bnb-chain/zkbas-eth-rpc => ../zkbas-eth-rpc
