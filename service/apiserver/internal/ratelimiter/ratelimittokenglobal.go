package ratelimiter

import (
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

// RateLimitControlGlobal Rate Limit Control in global dimension
func (r *TokenRateLimiter) RateLimitControlGlobal(param *RateLimitParam, controller RateLimitController) error {
	tokenLimiter := r.GetGlobalRateLimiter(param.RequestPath)
	if tokenLimiter.Allow() {
		return controller(param, nil)
	} else {
		logx.Infof("LimitControlGlobal hit Token Limit, path:%s!", param.RequestPath)
		return errors.New("Too Many Request!")
	}
}
