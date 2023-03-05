package permctrl

import (
	"encoding/json"
	"github.com/apolloconfig/agollo/v4"
	apollo "github.com/apolloconfig/agollo/v4/env/config"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/config"
	"github.com/zeromicro/go-zero/core/logx"
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
	ControlBlockList      []string
}

type PermissionControlConfig struct {
	SwitchPermissionControlConfig      bool
	DefaultPermissionControlConfigItem PermissionControlItem
	TxTypePermissionControlConfigItem  map[uint32]PermissionControlItem
}

func (c *PermissionControlConfig) ValidatePermissionControlConfig() error {
	return nil
}

func (c *PermissionControlConfig) GetPermissionControlConfigItem(txType uint32) PermissionControlItem {
	if item, ok := c.TxTypePermissionControlConfigItem[txType]; ok {
		return item
	} else {
		return c.DefaultPermissionControlConfigItem
	}
}

func LoadApolloPermissionControlConfig(config config.Config) *PermissionControlConfig {

	apolloConfig := &apollo.AppConfig{
		AppID:          config.Apollo.AppID,
		Cluster:        config.Apollo.Cluster,
		IP:             config.Apollo.ApolloIp,
		NamespaceName:  config.Apollo.Namespace,
		IsBackupConfig: config.Apollo.IsBackupConfig,
	}

	client, err := agollo.StartWithConfig(func() (*apollo.AppConfig, error) {
		return apolloConfig, nil
	})
	if err != nil {
		logx.Severef("Fail to start Apollo Client in Permission Control Configuration, Reason:%s", err.Error())
		panic("Fail to start Apollo Client in Permission Control Configuration!")
	}

	apolloClient = client

	permControlUpdater := &PermControlUpdater{}
	apolloClient.AddChangeListener(permControlUpdater)

	apolloCache := apolloClient.GetConfigCache(apolloConfig.NamespaceName)
	permissionControlConfigObject, err := apolloCache.Get(PermissionControlConfigKey)
	if err != nil {
		logx.Severef("Fail to get PermissionControlConfig from the apollo server, Reason:%s", err.Error())
		panic("Fail to get PermissionControlConfig from the apollo server!")
	}
	if permissionControlConfigString, ok := permissionControlConfigObject.(string); ok {
		permissionControlConfig := &PermissionControlConfig{}
		err := json.Unmarshal([]byte(permissionControlConfigString), permissionControlConfig)
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
	} else {
		logx.Severef("Fail to Initiate PermissionControlConfig from the apollo server!")
		panic("Fail to Initiate PermissionControlConfig from the apollo server!")
	}
}
