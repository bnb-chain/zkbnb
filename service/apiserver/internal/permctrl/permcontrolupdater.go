package permctrl

import (
	"encoding/json"
	"github.com/apolloconfig/agollo/v4/storage"
	"github.com/zeromicro/go-zero/core/logx"
)

type PermControlUpdater struct {
}

func (u *PermControlUpdater) OnChange(event *storage.ChangeEvent) {
	configChange := event.Changes[PermissionControlConfigKey]
	newRateLimitConfigObject := configChange.NewValue
	if newRateLimitConfigJson, ok := newRateLimitConfigObject.(string); ok {
		newPermissionControlConfig := &PermissionControlConfig{}
		err := json.Unmarshal([]byte(newRateLimitConfigJson), newPermissionControlConfig)
		if err != nil {
			logx.Errorf("Fail to update PermissionControlConfig from the apollo server, Reason:%s", err.Error())
		}

		// Validate the permission control configuration from the apollo server side
		if err = newPermissionControlConfig.ValidatePermissionControlConfig(); err != nil {
			logx.Severef("Fail to validate PermissionControlConfig from the apollo server, Reason:%s", err.Error())
			panic("Fail to validate PermissionControlConfig from the apollo server!")
		}
		permissionControlConfig = newPermissionControlConfig
		logx.Info("Update PermissionControlConfig Successfully!")
	}
}

func (u *PermControlUpdater) OnNewestChange(event *storage.FullChangeEvent) {
	logx.Infof("Received Permission Control Configuration Update!")
}
