package ratelimiter

import (
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
)

const RateLimitByAccountKeyFormat = "limit:accountId:%s:%s"

// RateLimitControlByUserId Rate Limit Control by user dimension
func (r *PeriodRateLimiter) RateLimitControlByUserId(param *RateLimitParam, controller RateLimitController) error {
	// If the userId is blank or empty, that means the accountId field has not been passed,
	// so we would not do the accountId dimension rate limit control
	if len(param.AccountId) == 0 {
		return nil
	}
	rateLimitKey := fmt.Sprintf(RateLimitByAccountKeyFormat, param.RequestPath, param.AccountId)
	periodLimit := r.GetUserRateLimiter(param.RequestPath)
	if err := r.RateLimitControl(rateLimitKey, periodLimit); err != nil {
		logx.Error("RateLimitControlByUserId hit Period Limit, path:%s, userId:%s!",
			param.RequestPath, param.AccountId)
		return err
	}
	return nil
}
