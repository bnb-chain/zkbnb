package ratelimiter

import (
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
)

const RateLimitGlobalKeyFormat = "limit:global:%s"

// RateLimitControlGlobal Rate Limit Control in global dimension
func (r *PeriodRateLimiter) RateLimitControlGlobal(param *RateLimitParam, controller RateLimitController) error {
	rateLimitKey := fmt.Sprintf(RateLimitGlobalKeyFormat, param.RequestPath)
	periodLimit := r.GetGlobalRateLimiter(param.RequestPath)
	if err := r.RateLimitControl(rateLimitKey, periodLimit); err != nil {
		logx.Errorf("RateLimitControlGlobal hit Period Limit, path:%s!", param.RequestPath)
		return err
	}
	return controller(param, r.RateLimitControlByUserId)
}
