package ratelimiter

import (
	"encoding/json"
	"github.com/zeromicro/go-zero/core/logx"
	"os"
)

const (
	LimitTypePeriod = "LimitByPeriod"
	LimitTypeToken  = "LimitByToken"
	LimitTypeBoth   = "LimitByBoth"
)

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

func LoadRateLimitConfig(configFilePath string) (*RateLimitConfig, error) {

	jsonFile, err := os.Open(configFilePath)
	if err != nil {
		logx.Severef("Open Rate Limit Configuration File Raise Error, configFilePath:%s!", configFilePath)
		panic("Open Rate Limit Configuration File Raise Error:" + err.Error())
	}

	rateLimitConfig := &RateLimitConfig{}
	decoder := json.NewDecoder(jsonFile)
	if err = decoder.Decode(rateLimitConfig); err != nil {
		logx.Severef("Decode Rate Limit Configuration Raise Error, configFilePath:%s!", configFilePath)
		panic("Decode Rate Limit Configuration Raise Error:" + err.Error())
	}

	// Do the rate limit config validation after it is loaded
	if err = rateLimitConfig.ValidateRateLimitConfig(); err != nil {
		return nil, err
	}

	return rateLimitConfig, nil
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
