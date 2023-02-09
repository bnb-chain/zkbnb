package ratelimiter

import (
	"github.com/zeromicro/go-zero/core/limit"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type TokenRateLimiter struct {
	LocalHostID     string
	RateLimitConfig *RateLimitConfig

	GlobalRateLimitDefault *limit.TokenLimiter
	SingleRateLimitDefault *limit.TokenLimiter

	GlobalRateLimitMap map[string]*limit.TokenLimiter
	SingleRateLimitMap map[string]*limit.TokenLimiter
}

func InitRateLimitControlByToken(localHostID string,
	rateLimitConfig *RateLimitConfig, redisInstance *redis.Redis) *TokenRateLimiter {

	tokenRateLimitItem := rateLimitConfig.DefaultRateLimit.TokenRateLimitItem
	globalRateLimitDefault, singleRateLimitDefault := InitDefaultRateLimitControlByToken(tokenRateLimitItem, redisInstance)

	pathRateLimitMap := rateLimitConfig.PathRateLimitMap
	globalPathRateLimitMap, singlePathRateLimitMap := InitPathRateLimitControlByToken(pathRateLimitMap, redisInstance)

	tokenRateLimiter := &TokenRateLimiter{
		LocalHostID:            localHostID,
		RateLimitConfig:        rateLimitConfig,
		GlobalRateLimitDefault: globalRateLimitDefault,
		SingleRateLimitDefault: singleRateLimitDefault,
		GlobalRateLimitMap:     globalPathRateLimitMap,
		SingleRateLimitMap:     singlePathRateLimitMap,
	}
	return tokenRateLimiter
}

func InitDefaultRateLimitControlByToken(tokenRateLimitItem TokenRateLimitItem,
	redisInstance *redis.Redis) (*limit.TokenLimiter, *limit.TokenLimiter) {
	globalTokenLimiter := limit.NewTokenLimiter(tokenRateLimitItem.GlobalRate,
		tokenRateLimitItem.GlobalBurst, redisInstance, "DefaultGlobalTokenLimiter")

	singleTokenLimiter := limit.NewTokenLimiter(tokenRateLimitItem.SingleRate,
		tokenRateLimitItem.SingleBurst, redisInstance, "DefaultSingleTokenLimiter")

	return globalTokenLimiter, singleTokenLimiter
}

func InitPathRateLimitControlByToken(pathRateLimitMap map[string]RateLimitConfigItem,
	redisInstance *redis.Redis) (map[string]*limit.TokenLimiter, map[string]*limit.TokenLimiter) {
	globalRateLimitMap := make(map[string]*limit.TokenLimiter)
	singleRateLimitMap := make(map[string]*limit.TokenLimiter)

	for path, item := range pathRateLimitMap {
		tokenRateLimitItem := item.TokenRateLimitItem

		// Only if RateLimitType is set to LimitTypeToken,
		// the rateLimitMap could be initiated correctly.
		if item.RateLimitType == LimitTypeToken || item.RateLimitType == LimitTypeBoth {
			globalRateLimitMap[path] = limit.NewTokenLimiter(tokenRateLimitItem.GlobalRate, tokenRateLimitItem.GlobalBurst,
				redisInstance, "PathGlobalTokenLimiter-"+path)

			singleRateLimitMap[path] = limit.NewTokenLimiter(tokenRateLimitItem.SingleRate, tokenRateLimitItem.SingleBurst,
				redisInstance, "PathSingleTokenLimiter-"+path)
		}
	}

	return globalRateLimitMap, singleRateLimitMap
}

func (r *TokenRateLimiter) StartRateLimitControl(param *RateLimitParam) error {

	// Only if the request path rate limit has been set for the
	// token rate limit control, we do the limit control process
	if r.RateLimitConfig.IsTokenLimitType(param.RequestPath) {
		logx.Infof("Start Token Rate Limit Control for path:%s!", param.RequestPath)
		err := r.RateLimitControlGlobal(param, r.RateLimitControlSingle)
		logx.Infof("End Token Rate Limit Control for path:%s!", param.RequestPath)
		return err
	}
	return nil
}
