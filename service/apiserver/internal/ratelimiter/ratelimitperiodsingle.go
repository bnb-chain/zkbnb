package ratelimiter

import (
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
)

const RateLimitSingleKeyFormat = "limit:single:%s:%s"

// RateLimitControlSingle Rate Limit Control in single instance dimension
func (r *PeriodRateLimiter) RateLimitControlSingle(param *RateLimitParam, controller RateLimitController) error {
	rateLimitKey := fmt.Sprintf(RateLimitSingleKeyFormat, param.RequestPath, r.LocalHostID)
	periodLimit := r.SingleRateLimitMap[param.RequestPath]
	if err := r.RateLimitControl(rateLimitKey, periodLimit); err != nil {
		logx.Error("RateLimitControlSingle hit Period Limit, path:%s, hostId:%s!", param.RequestPath, r.LocalHostID)
		return err
	}
	return controller(param, nil)
}
