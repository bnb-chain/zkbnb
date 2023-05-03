package permctrl

import (
	"encoding/json"
	"github.com/apolloconfig/agollo/v4"
	"github.com/bnb-chain/zkbnb/common/apollo"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	ApiServerAppId = "zkbnb-apiserver"
	Namespace      = "application"
)

const (
	ControlByWhitelist = "ControlByWhitelist"
	ControlByBlacklist = "ControlByBlacklist"

	PermissionControlConfigKey = "PermissionControlConfig"
)

// Permission control configuration initialized from the apollo server
// And for permission control both in permcontrol and permcheck logic
var permissionControlConfig *PermissionControlConfig

// Apollo client to get the permission control configuration
// so that it could be updated it from the apollo server side
var apolloClient agollo.Client
var permControlUpdater = &PermControlUpdater{}

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

func (c *PermissionControlConfig) GetPermissionControlConfigItem(txType uint32) PermissionControlItem {
	if item, ok := c.TxTypePermissionControlConfigItemMap[txType]; ok {
		return item
	} else {
		return c.DefaultPermissionControlConfigItem
	}
}

func InitPermissionControl() {
	// Get the permission control configuration from the Apollo server
	permissionControlConfig = LoadApolloPermissionControlConfig()
	logx.Infof("Initiate Permission Control Facility Successfully!")
}

func LoadApolloPermissionControlConfig() *PermissionControlConfig {

	//Add the apollo configuration updater listener for PermissionConfig
	apollo.AddChangeListener(ApiServerAppId, Namespace, permControlUpdater)

	permissionControlConfig := &PermissionControlConfig{}
	permissionControlConfigString, err := apollo.LoadApolloConfigFromEnvironment(ApiServerAppId, Namespace, PermissionControlConfigKey)
	if err != nil {
		// If fails to initiate permission control config from apollo, directly switch this off
		logx.Errorf("Fail to Initiate PermissionControlConfig from the apollo server!")
		permissionControlConfig.SwitchPermissionControlConfig = false
	} else {

		err = json.Unmarshal([]byte(permissionControlConfigString), permissionControlConfig)
		if err != nil {
			logx.Errorf("Fail to unmarshal PermissionControlConfig from the apollo server, Reason:%s", err.Error())
			panic("Fail to unmarshal PermissionControlConfig from the apollo server!")
		}

		if err = permissionControlConfig.ValidatePermissionControlConfig(); err != nil {
			logx.Errorf("Fail to validate PermissionControlConfig from the apollo server, Reason:%s", err.Error())
			panic("Fail to validate PermissionControlConfig from the apollo server!")
		}

		logx.Info("Load PermissionControlConfig Successfully!")
	}
	return permissionControlConfig
}

func (c *PermissionControlConfig) ValidatePermissionControlConfig() error {
	return nil
}
