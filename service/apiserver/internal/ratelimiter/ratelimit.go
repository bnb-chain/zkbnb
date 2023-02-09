package ratelimiter

import (
	"github.com/shirou/gopsutil/host"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"net/http"
)

const AccountId = "accountId"

var periodRateLimiter *PeriodRateLimiter
var tokenRateLimiter *TokenRateLimiter

type RateLimitParam struct {
	RequestPath string
	AccountId   string
}

type RateLimitController func(*RateLimitParam, RateLimitController) error

func RateLimitHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {

		requestPath := request.URL.Path
		accountId := request.Form.Get(AccountId)

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

func InitRateLimitControl(configFilePath string) {

	localhostID := InitLocalhostConfiguration()
	rateLimitConfig := InitRateLimitConfiguration(configFilePath)
	redisInstance := InitRateLimitRedisInstance(rateLimitConfig.RedisConfig.Address)

	periodRateLimiter = InitRateLimitControlByPeriod(localhostID, rateLimitConfig, redisInstance)
	tokenRateLimiter = InitRateLimitControlByToken(localhostID, rateLimitConfig, redisInstance)

	logx.Info("Initiate RateLimit Control Successfully!")
}

func InitRateLimitRedisInstance(redisAddress string) *redis.Redis {
	redisInstance := redis.New(redisAddress)
	logx.Info("Initiate RateLimit Redis Successfully!")
	return redisInstance
}

func InitLocalhostConfiguration() string {
	localHostID, err := host.HostID()
	if err != nil {
		logx.Severe("Initiate LocalHostID Failure!")
		panic("Fail to initiate localhostID!")
	}
	return localHostID
}

func InitRateLimitConfiguration(configFilePath string) *RateLimitConfig {
	rateLimitConfig, err := LoadRateLimitConfig(configFilePath)
	if err != nil {
		logx.Severef("Initiate Rate Limit Configuration Failure, Reason:%s", err.Error())
		panic("Fail to initiate Rate Limit Configuration!")
	}
	logx.Info("Initiate RateLimit Configuration Successfully!")
	return rateLimitConfig
}
