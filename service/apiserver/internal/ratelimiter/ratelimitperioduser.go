package ratelimiter

import (
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
)

const RateLimitByL1AddressFormat = "limit:l1Address:%s:%s"

// RateLimitControlByL1Address Rate Limit Control by user dimension
func (r *PeriodRateLimiter) RateLimitControlByL1Address(param *RateLimitParam, controller RateLimitController) error {
	// If the l1Address is blank or empty, that means the l1Address field has not been passed,
	// so we would not do the accountId dimension rate limit control
	if len(param.L1Address) == 0 {
		return nil
	}
	rateLimitKey := fmt.Sprintf(RateLimitByL1AddressFormat, param.RequestPath, param.L1Address)
	periodLimit := r.GetUserRateLimiter(param.RequestPath)
	if err := r.RateLimitControl(rateLimitKey, periodLimit); err != nil {
		logx.Errorf("RateLimitControlByL1Address hit Period Limit, path:%s, l1Address:%s!",
			param.RequestPath, param.L1Address)
		return err
	}
	return nil
}
