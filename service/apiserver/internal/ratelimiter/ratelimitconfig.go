package ratelimiter

import (
	"encoding/json"
	"github.com/apolloconfig/agollo/v4"
	apollo "github.com/apolloconfig/agollo/v4/env/config"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/config"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	LimitTypePeriod = "LimitByPeriod"
	LimitTypeToken  = "LimitByToken"
	LimitTypeBoth   = "LimitByBoth"
)

// Apollo client to get the rate limit configuration
// and update it from the apollo server side
var apolloClient agollo.Client

type RedisConfig struct {
	Address string
}

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
	RedisConfig      RedisConfig
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

func LoadApolloRateLimitConfig(config config.Config) *RateLimitConfig {

	apolloConfig := &apollo.AppConfig{
		AppID:          config.Apollo.AppID,
		Cluster:        config.Apollo.Cluster,
		IP:             config.Apollo.ApolloIp,
		NamespaceName:  config.Apollo.Namespace,
		IsBackupConfig: config.Apollo.IsBackupConfig,
	}

	client, err := agollo.StartWithConfig(func() (*apollo.AppConfig, error) {
		return apolloConfig, nil
	})
	if err != nil {
		logx.Severef("Fail to start Apollo Client in RateLimit Configuration, Reason:%s", err.Error())
		panic("Fail to start Apollo Client in RateLimit Configuration!")
	}
	apolloClient = client
	rateLimitUpdater := &RateLimitUpdater{}
	apolloClient.AddChangeListener(rateLimitUpdater)

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

		logx.Info("Load RateLimitConfig Successfully!")
		return rateLimitConfig
	} else {
		logx.Severef("Fail to Initiate RateLimitConfig from the apollo server!")
		panic("Fail to Initiate RateLimitConfig from the apollo server!")
	}
}
