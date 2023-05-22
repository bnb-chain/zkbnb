package ratelimiter

import (
	"encoding/json"
	"github.com/bnb-chain/zkbnb/common/apollo"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
)

const (
	ApiServerAppId = "zkbnb-apiserver"
	Namespace      = "application"
)

const (
	LimitTypePeriod = "LimitByPeriod"
	LimitTypeToken  = "LimitByToken"
	LimitTypeBoth   = "LimitByBoth"
)

var rateLimitUpdater = &RateLimitUpdater{}

type PeriodRateLimitItem struct {
	GlobalRateSecond int
	GlobalRateQuota  int

	SingleRateSecond int
	SingleRateQuota  int

	UserRateSecond int
	UserRateQuota  int
}

type TokenRateLimitItem struct {
	GlobalRate  int
	GlobalBurst int

	SingleRate  int
	SingleBurst int
}

type RateLimitConfigItem struct {
	RateLimitType       string
	PeriodRateLimitItem PeriodRateLimitItem
	TokenRateLimitItem  TokenRateLimitItem
}

type RateLimitConfig struct {
	RateLimitSwitch  bool
	CacheRedis       cache.CacheConf
	DefaultRateLimit RateLimitConfigItem
	PathRateLimitMap map[string]RateLimitConfigItem
}

func (c *RateLimitConfig) IsPeriodLimitType(requestPath string) bool {
	// If the request path has been set in PathRateLimitMap
	// distinguish whether the limit type is LimitTypePeriod
	if configItem, ok := c.PathRateLimitMap[requestPath]; ok {
		return configItem.RateLimitType == LimitTypePeriod || configItem.RateLimitType == LimitTypeBoth
	}
	//If the request path has not been set in PathRateLimitMap
	//it is limited by default, so return true naturally
	return true
}

func (c *RateLimitConfig) IsTokenLimitType(requestPath string) bool {
	// If the request path has been set in PathRateLimitMap
	// distinguish whether the limit type is LimitTypeToken
	if configItem, ok := c.PathRateLimitMap[requestPath]; ok {
		return configItem.RateLimitType == LimitTypeToken || configItem.RateLimitType == LimitTypeBoth
	}
	//If the request path has not been set in PathRateLimitMap
	//it is limited by default, so return true naturally
	return true
}

func LoadApolloRateLimitConfig() *RateLimitConfig {

	//Add the apollo configuration updater listener for RateLimitConfig
	apollo.AddChangeListener(ApiServerAppId, Namespace, rateLimitUpdater)

	rateLimitConfig := &RateLimitConfig{}
	rateLimitConfigString, err := apollo.LoadApolloConfigFromEnvironment(ApiServerAppId, Namespace, RateLimitConfigKey)
	if err != nil {
		// If fails to initiate rate limit configuration from apollo, directly switch it off
		logx.Errorf("Fail to Initiate RateLimitConfig from the apollo server!")
		rateLimitConfig.RateLimitSwitch = false
	} else {
		err = json.Unmarshal([]byte(rateLimitConfigString), rateLimitConfig)
		if err != nil {
			logx.Errorf("Fail to unmarshal RateLimitConfig from the apollo server, Reason:%s", err.Error())
			panic("Fail to unmarshal RateLimitConfig from the apollo server!")
		}

		if err = rateLimitConfig.ValidateRateLimitConfig(); err != nil {
			logx.Errorf("Fail to validate RateLimitConfig from the apollo server, Reason:%s", err.Error())
			panic("Fail to validate RateLimitConfig from the apollo server!")
		}

		logx.Info("Load RateLimitConfig Successfully!")
	}
	return rateLimitConfig
}
