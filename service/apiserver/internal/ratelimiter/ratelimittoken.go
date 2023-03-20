package ratelimiter

import (
	"fmt"
	"github.com/zeromicro/go-zero/core/limit"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	DefaultGlobalTokenLimiterKey = "DefaultGlobalTokenLimiter"
	DefaultSingleTokenLimiterKey = "DefaultSingleTokenLimiter:%s"

	PathGlobalTokenLimiterKey = "PathGlobalTokenLimiter:%s"
	PathSingleTokenLimiterKey = "PathSingleTokenLimiter:%s:%s"
)

type TokenRateLimiter struct {
	LocalhostID     string
	RateLimitConfig *RateLimitConfig

	GlobalRateLimitDefault *limit.TokenLimiter
	SingleRateLimitDefault *limit.TokenLimiter

	GlobalRateLimitMap map[string]*limit.TokenLimiter
	SingleRateLimitMap map[string]*limit.TokenLimiter
}

func InitRateLimitControlByToken(localhostID string,
	rateLimitConfig *RateLimitConfig, redisInstance *redis.Redis) *TokenRateLimiter {

	tokenRateLimitItem := rateLimitConfig.DefaultRateLimit.TokenRateLimitItem
	globalRateLimitDefault, singleRateLimitDefault := InitDefaultRateLimitControlByToken(localhostID, tokenRateLimitItem, redisInstance)

	pathRateLimitMap := rateLimitConfig.PathRateLimitMap
	globalPathRateLimitMap, singlePathRateLimitMap := InitPathRateLimitControlByToken(localhostID, pathRateLimitMap, redisInstance)

	tokenRateLimiter := &TokenRateLimiter{
		LocalhostID:            localhostID,
		RateLimitConfig:        rateLimitConfig,
		GlobalRateLimitDefault: globalRateLimitDefault,
		SingleRateLimitDefault: singleRateLimitDefault,
		GlobalRateLimitMap:     globalPathRateLimitMap,
		SingleRateLimitMap:     singlePathRateLimitMap,
	}
	logx.Info("Construct Token RateLimit Facility Successfully!")
	return tokenRateLimiter
}

func InitDefaultRateLimitControlByToken(localhostId string, tokenRateLimitItem TokenRateLimitItem,
	redisInstance *redis.Redis) (*limit.TokenLimiter, *limit.TokenLimiter) {
	globalTokenLimiter := limit.NewTokenLimiter(tokenRateLimitItem.GlobalRate,
		tokenRateLimitItem.GlobalBurst, redisInstance, DefaultGlobalTokenLimiterKey)

	singleTokenLimiter := limit.NewTokenLimiter(tokenRateLimitItem.SingleRate,
		tokenRateLimitItem.SingleBurst, redisInstance, fmt.Sprintf(DefaultSingleTokenLimiterKey, localhostId))

	return globalTokenLimiter, singleTokenLimiter
}

func InitPathRateLimitControlByToken(localhostId string, pathRateLimitMap map[string]RateLimitConfigItem,
	redisInstance *redis.Redis) (map[string]*limit.TokenLimiter, map[string]*limit.TokenLimiter) {
	globalRateLimitMap := make(map[string]*limit.TokenLimiter)
	singleRateLimitMap := make(map[string]*limit.TokenLimiter)

	for path, item := range pathRateLimitMap {
		tokenRateLimitItem := item.TokenRateLimitItem

		// Only if RateLimitType is set to LimitTypeToken,
		// the rateLimitMap could be initiated correctly.
		if item.RateLimitType == LimitTypeToken || item.RateLimitType == LimitTypeBoth {
			globalRateLimitMap[path] = limit.NewTokenLimiter(tokenRateLimitItem.GlobalRate, tokenRateLimitItem.GlobalBurst,
				redisInstance, fmt.Sprintf(PathGlobalTokenLimiterKey, path))

			singleRateLimitMap[path] = limit.NewTokenLimiter(tokenRateLimitItem.SingleRate, tokenRateLimitItem.SingleBurst,
				redisInstance, fmt.Sprintf(PathSingleTokenLimiterKey, path, localhostId))
		}
	}

	return globalRateLimitMap, singleRateLimitMap
}

func (r *TokenRateLimiter) StartRateLimitControl(param *RateLimitParam) error {

	// Only if the request path rate limit has been set for the
	// token rate limit control, we do the limit control process
	if r.RateLimitConfig.IsTokenLimitType(param.RequestPath) {
		err := r.RateLimitControlGlobal(param, r.RateLimitControlSingle)
		return err
	}
	return nil
}

func (r *TokenRateLimiter) GetGlobalRateLimiter(requestPath string) *limit.TokenLimiter {
	if rateLimiter, ok := r.GlobalRateLimitMap[requestPath]; ok {
		return rateLimiter
	}
	return r.GlobalRateLimitDefault
}

func (r *TokenRateLimiter) GetSingleRateLimiter(requestPath string) *limit.TokenLimiter {
	if rateLimiter, ok := r.SingleRateLimitMap[requestPath]; ok {
		return rateLimiter
	}
	return r.SingleRateLimitDefault
}
