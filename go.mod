module github.com/zecrey-labs/zecrey-legend

go 1.16

require (
	github.com/zeromicro/go-zero v1.3.2
	gorm.io/gorm v1.23.4
)

require (
	github.com/consensys/gnark-crypto v0.7.0
	github.com/ethereum/go-ethereum v1.10.17
	github.com/robfig/cron/v3 v3.0.1
	github.com/zecrey-labs/zecrey-core v0.0.3
	google.golang.org/grpc v1.44.0
	google.golang.org/protobuf v1.27.1
	gorm.io/driver/postgres v1.3.4
)

replace github.com/zecrey-labs/zecrey-crypto => ../zecrey-crypto

replace github.com/zecrey-labs/zecrey-eth-rpc => ../zecrey-eth-rpc

replace github.com/zecrey-labs/zecrey-core => ../zecrey-core
