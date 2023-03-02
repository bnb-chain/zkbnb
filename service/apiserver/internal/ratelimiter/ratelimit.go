package ratelimiter

import (
	"encoding/json"
	"github.com/apolloconfig/agollo/v4"
	apollo "github.com/apolloconfig/agollo/v4/env/config"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/config"
	"github.com/shirou/gopsutil/host"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"net/http"
)

const (
	AccountId          = "accountId"
	RateLimitConfigKey = "RateLimitConfig"
)

var apolloClient *agollo.Client

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

func InitRateLimitControl(config config.Config) {

	localhostID := InitLocalhostConfiguration()
	rateLimitConfig := InitRateLimitConfiguration(config)
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

func InitRateLimitConfiguration(config config.Config) *RateLimitConfig {

	apolloConfig := &apollo.AppConfig{
		AppID:          config.Apollo.AppID,
		Cluster:        config.Apollo.Cluster,
		IP:             config.Apollo.ApolloIp,
		NamespaceName:  config.Apollo.Namespace,
		IsBackupConfig: config.Apollo.IsBackupConfig,
	}

	apolloClient, err := agollo.StartWithConfig(func() (*apollo.AppConfig, error) {
		return apolloConfig, nil
	})
	if err != nil {
		logx.Severef("Start Apollo Client In RateLimit Configuration Failure, Reason:%s", err.Error())
		panic("Fail to start Apollo Client in RateLimit Configuration!")
	}

	apolloCache := apolloClient.GetConfigCache(apolloConfig.NamespaceName)
	rateLimitConfigObject, err := apolloCache.Get(RateLimitConfigKey)
	if err != nil {
		logx.Severef("Fail to get RateLimitConfig from the apollo server, Reason:%s", err.Error())
		panic("Fail to get RateLimitConfig from the apollo server!")
	}

	if rateLimitConfigString, ok := rateLimitConfigObject.(string); ok {
		rateLimitConfig := &RateLimitConfig{}
		err := json.Unmarshal([]byte(rateLimitConfigString), rateLimitConfig)
		if err != nil {
			logx.Severef("Fail to unmarshal RateLimitConfig from the apollo server, Reason:%s", err.Error())
			panic("Fail to unmarshal RateLimitConfig from the apollo server!")
		}

		if err = rateLimitConfig.ValidateRateLimitConfig(); err != nil {
			logx.Severef("Fail to validate RateLimitConfig from the apollo server, Reason:%s", err.Error())
			panic("Fail to validate RateLimitConfig from the apollo server!")
		}

		logx.Info("Initiate RateLimit Configuration Successfully!")
		return rateLimitConfig

	} else {
		logx.Severef("Fail to Initiate RateLimitConfig from the apollo server!")
		panic("Fail to Initiate RateLimitConfig from the apollo server!")
	}
}
