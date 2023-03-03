package ratelimiter

import (
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/limit"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type PeriodRateLimiter struct {
	LocalHostID     string
	RateLimitConfig *RateLimitConfig

	GlobalRateLimitDefault *limit.PeriodLimit
	SingleRateLimitDefault *limit.PeriodLimit
	UserRateLimitDefault   *limit.PeriodLimit

	GlobalRateLimitMap map[string]*limit.PeriodLimit
	SingleRateLimitMap map[string]*limit.PeriodLimit
	UserRateLimitMap   map[string]*limit.PeriodLimit
}

func InitRateLimitControlByPeriod(localHostID string,
	rateLimitConfig *RateLimitConfig, redisInstance *redis.Redis) *PeriodRateLimiter {

	defaultRateLimitConfig := rateLimitConfig.DefaultRateLimit.PeriodRateLimitItem
	globalRateLimitDefault, singleRateLimitDefault, userRateLimitDefault :=
		InitDefaultRateLimitControlByPeriod(defaultRateLimitConfig, redisInstance)

	pathRateLimitMap := rateLimitConfig.PathRateLimitMap
	globalRateLimitMap, singleRateLimitMap, userRateLimitMap :=
		InitRateLimitControlPathMapByPeriod(pathRateLimitMap, redisInstance)

	periodRateLimiter := &PeriodRateLimiter{
		LocalHostID:            localHostID,
		RateLimitConfig:        rateLimitConfig,
		GlobalRateLimitDefault: globalRateLimitDefault,
		SingleRateLimitDefault: singleRateLimitDefault,
		UserRateLimitDefault:   userRateLimitDefault,
		GlobalRateLimitMap:     globalRateLimitMap,
		SingleRateLimitMap:     singleRateLimitMap,
		UserRateLimitMap:       userRateLimitMap,
	}

	logx.Info("Construct Period RateLimit Facility Successfully!")
	return periodRateLimiter
}

func InitDefaultRateLimitControlByPeriod(defalutRateLimitConfig PeriodRateLimitItem,
	redisInstance *redis.Redis) (*limit.PeriodLimit, *limit.PeriodLimit, *limit.PeriodLimit) {
	globalRateLimitDefault := limit.NewPeriodLimit(defalutRateLimitConfig.GlobalRateSecond,
		defalutRateLimitConfig.GlobalRateQuota, redisInstance, "DefaultGlobalPeriodLimiter")

	singleRateLimitDefault := limit.NewPeriodLimit(defalutRateLimitConfig.SingleRateSecond,
		defalutRateLimitConfig.SingleRateQuota, redisInstance, "DefaultSinglePeriodLimiter")

	userRateLimitDefault := limit.NewPeriodLimit(defalutRateLimitConfig.UserRateSecond,
		defalutRateLimitConfig.UserRateQuota, redisInstance, "DefaultUserPeriodLimiter")

	return globalRateLimitDefault, singleRateLimitDefault, userRateLimitDefault
}

func InitRateLimitControlPathMapByPeriod(pathRateLimitMap map[string]RateLimitConfigItem,
	redisInstance *redis.Redis) (map[string]*limit.PeriodLimit, map[string]*limit.PeriodLimit, map[string]*limit.PeriodLimit) {

	globalPeriodRateLimitMap := make(map[string]*limit.PeriodLimit)
	singlePeriodRateLimitMap := make(map[string]*limit.PeriodLimit)
	userPeriodRateLimitMap := make(map[string]*limit.PeriodLimit)

	for path, item := range pathRateLimitMap {

		// Only if RateLimitType is set to LimitTypePeriod,
		// the rateLimitMap could be initiated correctly.
		if item.RateLimitType == LimitTypePeriod || item.RateLimitType == LimitTypeBoth {
			periodRateLimitItem := item.PeriodRateLimitItem
			globalPeriodRateLimitMap[path] = limit.NewPeriodLimit(periodRateLimitItem.GlobalRateSecond,
				periodRateLimitItem.GlobalRateQuota, redisInstance, "PathGlobalPeriodLimiter")

			singlePeriodRateLimitMap[path] = limit.NewPeriodLimit(periodRateLimitItem.SingleRateSecond,
				periodRateLimitItem.SingleRateQuota, redisInstance, "PathSinglePeriodLimiter")

			userPeriodRateLimitMap[path] = limit.NewPeriodLimit(periodRateLimitItem.UserRateSecond,
				periodRateLimitItem.UserRateQuota, redisInstance, "PathUserPeriodLimiter")
		}
	}

	return globalPeriodRateLimitMap, singlePeriodRateLimitMap, userPeriodRateLimitMap
}

func (r *PeriodRateLimiter) StartRateLimitControl(param *RateLimitParam) error {

	// Only if the request path rate limit has been set for the
	// period rate limit control, we do the limit control process
	if r.RateLimitConfig.IsPeriodLimitType(param.RequestPath) {
		err := r.RateLimitControlGlobal(param, r.RateLimitControlSingle)
		return err
	}
	return nil
}

func (r *PeriodRateLimiter) RateLimitControl(limitKey string, periodLimit *limit.PeriodLimit) error {
	status, err := periodLimit.Take(limitKey)
	if err != nil {
		return err
	}
	switch status {
	case limit.Allowed:
		return nil
	case limit.HitQuota:
		return errors.New("Too Many Request!")
	case limit.OverQuota:
		return errors.New("Too Many Request!")
	}
	return errors.New("Unknown Rate Limit Status Error!")
}

func (r *PeriodRateLimiter) GetGlobalRateLimiter(requestPath string) *limit.PeriodLimit {
	if rateLimiter, ok := r.GlobalRateLimitMap[requestPath]; ok {
		return rateLimiter
	}
	return r.GlobalRateLimitDefault
}

func (r *PeriodRateLimiter) GetSingleRateLimiter(requestPath string) *limit.PeriodLimit {
	if rateLimiter, ok := r.SingleRateLimitMap[requestPath]; ok {
		return rateLimiter
	}
	return r.SingleRateLimitDefault
}

func (r *PeriodRateLimiter) GetUserRateLimiter(requestPath string) *limit.PeriodLimit {
	if rateLimiter, ok := r.UserRateLimitMap[requestPath]; ok {
		return rateLimiter
	}
	return r.UserRateLimitDefault
}
