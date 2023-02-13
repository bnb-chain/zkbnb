package ratelimiter

import (
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

// RateLimitControlSingle Rate Limit Control in single instance dimension
func (r *TokenRateLimiter) RateLimitControlSingle(param *RateLimitParam, controller RateLimitController) error {
	tokenLimiter := r.GetSingleRateLimiter(param.RequestPath)
	if tokenLimiter.Allow() {
		return nil
	} else {
		logx.Errorf("RateLimitControlSingle hit Token Limit, path:%s, hostId:%s!", param.RequestPath, r.LocalHostID)
		return errors.New("Too Many Request!")
	}
}
