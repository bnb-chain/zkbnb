package permctrl

import (
	"encoding/json"
	"github.com/apolloconfig/agollo/v4"
	"github.com/bnb-chain/zkbnb/common/apollo"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	ApiServerAppId = "ApiServerAppId"
)

const (
	ControlByWhitelist = "ControlByWhitelist"
	ControlByBlacklist = "ControlByBlacklist"

	PermissionControlConfigKey = "PermissionControlConfig"
)

// Apollo client to get the permission control configuration
// so that it could be updated it from the apollo server side
var apolloClient agollo.Client

type PermissionControlItem struct {
	PermissionControlType string
	ControlWhiteList      []string
	ControlBlackList      []string
}

type PermissionControlConfig struct {
	SwitchPermissionControlConfig        bool
	DefaultPermissionControlConfigItem   PermissionControlItem
	TxTypePermissionControlConfigItemMap map[uint32]PermissionControlItem
}

func (c *PermissionControlConfig) ValidatePermissionControlConfig() error {
	return nil
}

func (c *PermissionControlConfig) GetPermissionControlConfigItem(txType uint32) PermissionControlItem {
	if item, ok := c.TxTypePermissionControlConfigItemMap[txType]; ok {
		return item
	} else {
		return c.DefaultPermissionControlConfigItem
	}
}

func LoadApolloPermissionControlConfig() *PermissionControlConfig {
	permControlUpdater := &PermControlUpdater{}
	apollo.AddChangeListener(ApiServerAppId, permControlUpdater)

	permissionControlConfigString, err := apollo.LoadApolloConfigFromEnvironment(ApiServerAppId, PermissionControlConfigKey)
	if err != nil {
		logx.Severef("Fail to Initiate PermissionControlConfig from the apollo server!")
		panic("Fail to Initiate PermissionControlConfig from the apollo server!")
	}

	permissionControlConfig := &PermissionControlConfig{}
	err = json.Unmarshal([]byte(permissionControlConfigString), permissionControlConfig)
	if err != nil {
		logx.Severef("Fail to unmarshal PermissionControlConfig from the apollo server, Reason:%s", err.Error())
		panic("Fail to unmarshal PermissionControlConfig from the apollo server!")
	}

	if err = permissionControlConfig.ValidatePermissionControlConfig(); err != nil {
		logx.Severef("Fail to validate PermissionControlConfig from the apollo server, Reason:%s", err.Error())
		panic("Fail to validate PermissionControlConfig from the apollo server!")
	}

	logx.Info("Load PermissionControlConfig Successfully!")
	return permissionControlConfig

}
