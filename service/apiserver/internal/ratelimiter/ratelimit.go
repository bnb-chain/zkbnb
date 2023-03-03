package ratelimiter

import (
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/config"
	"github.com/shirou/gopsutil/host"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"net/http"
)

const (
	AccountIndex       = "account_index"
	RateLimitConfigKey = "RateLimitConfig"
)

var localhostID string

var redisInstance *redis.Redis

var periodRateLimiter *PeriodRateLimiter
var tokenRateLimiter *TokenRateLimiter

type RateLimitParam struct {
	RequestPath string
	AccountId   string
}

type RateLimitController func(*RateLimitParam, RateLimitController) error

func RateLimitHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {

		// Parse the form before reading the parameter
		request.ParseForm()
		requestPath := request.URL.Path
		accountId := request.Form.Get(AccountIndex)

		// Construct the rate limit param for
		// the next rate limit control
		rateLimitParam := &RateLimitParam{
			RequestPath: requestPath,
			AccountId:   accountId,
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

func InitRateLimitControl(config config.Config) {

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
