package ratelimiter

import (
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
)

const RateLimitByUserKeyFormat = "limit:userId:%s:%s"

// RateLimitControlByUserId Rate Limit Control by user dimension
func (r *PeriodRateLimiter) RateLimitControlByUserId(param *RateLimitParam, controller RateLimitController) error {
	rateLimitKey := fmt.Sprintf(RateLimitByUserKeyFormat, param.RequestPath, param.UserIdentifier)
	periodLimit := r.UserRateLimitMap[param.RequestPath]
	if err := r.RateLimitControl(rateLimitKey, periodLimit); err != nil {
		logx.Error("RateLimitControlByUserId hit Period Limit, path:%s, userId:%s!",
			param.RequestPath, param.UserIdentifier)
		return err
	}
	return nil
}
