package ratelimiter

import (
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/cache"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/config"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	"github.com/shirou/gopsutil/host"
	"github.com/zeromicro/go-zero/core/logx"
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

var periodRateLimiter *PeriodRateLimiter
var tokenRateLimiter *TokenRateLimiter

type RateLimitParam struct {
	RequestPath string
	L1Address   string
}

type RateLimitController func(*RateLimitParam, RateLimitController) error

func RateLimitHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {

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

func InitRateLimitControl(svcCtx *svc.ServiceContext, config config.Config) {

	memCache = svcCtx.MemCache
	localhostID = InitLocalhostConfiguration()
	rateLimitConfig := LoadApolloRateLimitConfig(config)

	RefreshRateLimitControl(rateLimitConfig)

	logx.Info("Initiate RateLimit Control Facility Successfully!")
}

func RefreshRateLimitControl(rateLimitConfig *RateLimitConfig) {
	redisInstance = InitRateLimitRedisInstance(rateLimitConfig.RedisConfig.Address)

	periodRateLimiter = InitRateLimitControlByPeriod(localhostID, rateLimitConfig, redisInstance)
	tokenRateLimiter = InitRateLimitControlByToken(localhostID, rateLimitConfig, redisInstance)
}

func InitLocalhostConfiguration() string {
	localHostID, err := host.HostID()
	if err != nil {
		logx.Severe("Initiate LocalHostID Failure!")
		panic("Fail to initiate localhostID!")
	}
	return localHostID
}

func InitRateLimitRedisInstance(redisAddress string) *redis.Redis {
	redisInstance := redis.New(redisAddress)
	logx.Info("Construct RateLimit Redis Instance Successfully!")
	return redisInstance
}

func ParseAccountL1Address(r *http.Request) string {
	var req types.ReqAccountParam
	if err := httpx.Parse(r, &req); err != nil {
		return ""
	}

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

		l1Address, err := memCache.GetAccountL1AddressByIndex(accountIndex)
		if err != nil {
			return ""
		}
		return l1Address
	}

	l1Address, err := memCache.GetAccountL1AddressByIndex(req.AccountIndex)
	if err != nil {
		return ""
	}
	return l1Address
}
