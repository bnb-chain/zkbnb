module github.com/zecrey-labs/zecrey-legend

go 1.16

require (
	github.com/zeromicro/go-zero v1.3.2
	gorm.io/gorm v1.23.4
)

require (
	github.com/consensys/gnark-crypto v0.7.0
	github.com/google/uuid v1.3.0
	github.com/zecrey-labs/zecrey-crypto v0.0.26
	github.com/zeromicro/go-zero/tools/goctl v1.3.4
)

replace github.com/zecrey-labs/zecrey-crypto => github.com/zecrey-labs/zecrey-crypto v0.0.26
