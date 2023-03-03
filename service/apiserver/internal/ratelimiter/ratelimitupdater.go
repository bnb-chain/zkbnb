package ratelimiter

import (
	"encoding/json"
	"github.com/apolloconfig/agollo/v4/storage"
	"github.com/zeromicro/go-zero/core/logx"
)

type RateLimitUpdater struct {
}

func (u *RateLimitUpdater) OnChange(event *storage.ChangeEvent) {
	configChange := event.Changes[RateLimitConfigKey]
	newRateLimitConfigObject := configChange.NewValue
	if newRateLimitConfigJson, ok := newRateLimitConfigObject.(string); ok {
		newRateLimitConfig := &RateLimitConfig{}
		err := json.Unmarshal([]byte(newRateLimitConfigJson), newRateLimitConfig)
		if err != nil {
			logx.Errorf("Fail to update RateLimitConfig from the apollo server, Reason:%s", err.Error())
		}
		if err := newRateLimitConfig.validatePathRateLimitConfig(); err != nil {
			logx.Errorf("Fail to validate RateLimitConfig from the apollo server, Reason:%s", err.Error())
		}
		RefreshRateLimitControl(newRateLimitConfig)
		logx.Info("Update RateLimit Control Configuration Successfully!")
	}
}

func (u *RateLimitUpdater) OnNewestChange(event *storage.FullChangeEvent) {
	logx.Infof("Received RateLimit Control Configuration Update!")
}
