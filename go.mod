module github.com/zecrey-labs/zecrey-legend

go 1.16

require (
	github.com/zeromicro/go-zero v1.3.3
	gorm.io/gorm v1.23.4
)

require (
	github.com/eko/gocache/v2 v2.3.1
	github.com/go-redis/redis/v8 v8.11.5
	github.com/golang-jwt/jwt/v4 v4.3.0 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/zeromicro/go-zero/tools/goctl v1.3.5
	golang.org/x/crypto v0.0.0-20220321153916-2c7772ba3064 // indirect
	gorm.io/driver/postgres v1.3.6
)

replace github.com/zecrey-labs/zecrey-crypto => ../zecrey-crypto

replace github.com/zecrey-labs/zecrey-eth-rpc => ../zecrey-eth-rpc
