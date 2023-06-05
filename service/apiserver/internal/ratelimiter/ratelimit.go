package ratelimiter

import (
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/cache"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/fetcher/address"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	"github.com/shirou/gopsutil/host"
	"github.com/zeromicro/go-zero/core/logx"
	zeroCache "github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
	"strconv"
)

const (
	RateLimitConfigKey = "RateLimitConfig"
)

const (
	queryByIndex     = "index"
	queryByL1Address = "l1_address"

	queryByAccountIndex = "account_index"
)

var localhostID string

var redisInstance *redis.Redis
var memCache *cache.MemCache
var fetcher *address.Fetcher

var rateLimitConfig *RateLimitConfig

var periodRateLimiter *PeriodRateLimiter
var tokenRateLimiter *TokenRateLimiter

type RateLimitParam struct {
	RequestPath string
	L1Address   string
}

type RateLimitController func(*RateLimitParam, RateLimitController) error

func RateLimitHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {

		// If rateLimitConfig has not been set or RateLimitSwitch has been set to false
		// Do not perform any rate limit control and do the following process directly
		if rateLimitConfig == nil || !rateLimitConfig.RateLimitSwitch {
			next(writer, request)
			return
		}

		// Parse the form before reading the parameter
		request.ParseForm()
		requestPath := request.URL.Path
		l1Address := ParseAccountL1Address(request)

		// Construct the rate limit param for
		// the next rate limit control
		rateLimitParam := &RateLimitParam{
			RequestPath: requestPath,
			L1Address:   l1Address,
		}

		// Perform the Period Rate Limit Process
		if err := periodRateLimiter.StartRateLimitControl(rateLimitParam); err != nil {
			http.Error(writer, err.Error(), http.StatusTooManyRequests)
			return
		}

		// Perform the Token Rate Limit Process
		if err := tokenRateLimiter.StartRateLimitControl(rateLimitParam); err != nil {
			http.Error(writer, err.Error(), http.StatusTooManyRequests)
			return
		}

		// If not rate limited,
		// continue to do the next process
		next(writer, request)
	}
}

func InitRateLimitControl(svcCtx *svc.ServiceContext) {

	memCache = svcCtx.MemCache
	fetcher = address.NewFetcher(svcCtx)
	localhostID = InitLocalhostConfiguration()
	newRateLimitConfig := LoadApolloRateLimitConfig()

	RefreshRateLimitControl(newRateLimitConfig)
}

func RefreshRateLimitControl(newRateLimitConfig *RateLimitConfig) {
	rateLimitConfig = newRateLimitConfig
	if rateLimitConfig.RateLimitSwitch {
		redisInstance = InitRateLimitRedisInstance(rateLimitConfig.CacheRedis)

		periodRateLimiter = InitRateLimitControlByPeriod(localhostID, rateLimitConfig, redisInstance)
		tokenRateLimiter = InitRateLimitControlByToken(localhostID, rateLimitConfig, redisInstance)

		logx.Info("Initiate RateLimit Control Facility Successfully!")
	} else {
		logx.Info("RateLimitSwitch is Off, Do Not Initiate RateLimit Control Facility!")
	}
}

func InitLocalhostConfiguration() string {
	localHostID, err := host.HostID()
	if err != nil {
		logx.Severe("Initiate LocalHostID Failure!")
		panic("Fail to initiate localhostID!")
	}
	return localHostID
}

func InitRateLimitRedisInstance(cacheRedis zeroCache.CacheConf) *redis.Redis {
	redisInstance, err := redis.NewRedis(redis.RedisConf{Host: cacheRedis[0].Host, Pass: cacheRedis[0].Pass, Type: cacheRedis[0].Type})
	if err != nil {
		logx.Severe("Initiate RateLimitRedis Failure!")
		panic("Fail to initiate RateLimitRedis!")
	}
	logx.Info("Construct RateLimit Redis Instance Successfully!")
	return redisInstance
}

func ParseAccountL1Address(r *http.Request) string {
	var req types.ReqAccountParam
	if err := httpx.Parse(r, &req); err != nil {
		return ""
	}

	// For sending transaction interface, we get the l1 address in the below logic
	if len(req.TxInfo) > 0 && req.TxType > 0 {
		l1AddressList, err := fetcher.GetL1AddressByTx(req.TxType, req.TxInfo)
		if err != nil || len(l1AddressList) == 0 {
			return ""
		}
		return l1AddressList[0]
	}

	// If it is not the sending transaction request, we get the l1 address in the below logic again
	if len(req.By) > 0 {
		var accountIndex = int64(0)
		var err error

		if req.By == queryByAccountIndex || req.By == queryByIndex {
			accountIndex, err = strconv.ParseInt(req.Value, 10, 64)
			if err != nil || accountIndex < 0 {
				return ""
			}
		} else if req.By == queryByL1Address {
			return req.Value
		} else {
			return ""
		}

		l1Address, err := memCache.GetL1AddressByIndex(accountIndex)
		if err != nil {
			return ""
		}
		return l1Address
	}

	// If account index is present in the request, we get the l1 address directly
	l1Address, err := memCache.GetL1AddressByIndex(req.AccountIndex)
	if err != nil {
		return ""
	}
	return l1Address
}
