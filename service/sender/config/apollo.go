package config

import (
	"encoding/json"
	"github.com/apolloconfig/agollo/v4"
	apollo "github.com/apolloconfig/agollo/v4/env/config"
	"github.com/apolloconfig/agollo/v4/storage"
	"github.com/zeromicro/go-zero/core/logx"
)

const SenderConfigKey = "SenderConfig"

var senderConfig *SenderConfig = &SenderConfig{}

// Apollo client to get the sender configuration
// so that it could be updated it from the apollo server side
var apolloClient agollo.Client

type SenderConfig struct {
	MaxCommitBlockCount uint64
	CommitTxCountLimit  uint64

	MaxVerifyBlockCount uint64
	VerifyTxCountLimit  uint64

	MaxCommitTxCount uint64
	MaxVerifyTxCount uint64

	MaxCommitBlockInterval uint64
	MaxVerifyBlockInterval uint64

	MaxCommitAvgUnitGas uint64
	MaxVerifyAvgUnitGas uint64
}

type SenderUpdater struct {
}

func InitApolloConfiguration(c Config) {
	apolloConfig := &apollo.AppConfig{
		AppID:          c.Apollo.AppID,
		Cluster:        c.Apollo.Cluster,
		IP:             c.Apollo.ApolloIp,
		NamespaceName:  c.Apollo.Namespace,
		IsBackupConfig: c.Apollo.IsBackupConfig,
	}

	client, err := agollo.StartWithConfig(func() (*apollo.AppConfig, error) {
		return apolloConfig, nil
	})
	if err != nil {
		logx.Severef("Fail to start Apollo Client in Permission Control Configuration, Reason:%s", err.Error())
		panic("Fail to start Apollo Client in Permission Control Configuration!")
	}

	apolloClient = client
	senderUpdater := &SenderUpdater{}
	apolloClient.AddChangeListener(senderUpdater)

	apolloCache := apolloClient.GetConfigCache(apolloConfig.NamespaceName)
	newSenderConfigObject, err := apolloCache.Get(SenderConfigKey)
	if newSenderConfigObjectJson, ok := newSenderConfigObject.(string); ok {
		newSenderConfig := &SenderConfig{}
		err := json.Unmarshal([]byte(newSenderConfigObjectJson), newSenderConfig)
		if err != nil {
			logx.Errorf("Fail to update SenderConfig from the apollo server, Reason:%s", err.Error())
		}

		// Validate the Sender Configuration from the apollo server side
		if err = newSenderConfig.ValidateSenderConfig(); err != nil {
			logx.Severef("Fail to validate SenderConfig from the apollo server, Reason:%s", err.Error())
			panic("Fail to validate SenderConfig from the apollo server!")
		}
		senderConfig = newSenderConfig
	}
}

func (u *SenderUpdater) OnChange(event *storage.ChangeEvent) {
	configChange := event.Changes[SenderConfigKey]
	if configChange == nil {
		return
	}
	newSenderConfigObject := configChange.NewValue
	if newSenderConfigObjectJson, ok := newSenderConfigObject.(string); ok {
		newSenderConfig := &SenderConfig{}
		err := json.Unmarshal([]byte(newSenderConfigObjectJson), newSenderConfig)
		if err != nil {
			logx.Errorf("Fail to update SenderConfig from the apollo server, Reason:%s", err.Error())
			panic("Fail to update SenderConfig from the apollo server!")
		}

		// Validate the Sender Configuration from the apollo server side
		if err = newSenderConfig.ValidateSenderConfig(); err != nil {
			logx.Severef("Fail to validate SenderConfig from the apollo server, Reason:%s", err.Error())
			panic("Fail to validate SenderConfig from the apollo server!")
		}
		senderConfig = newSenderConfig
		logx.Info("Update SenderConfig Successfully!")
	}
}

func (u *SenderUpdater) OnNewestChange(event *storage.FullChangeEvent) {
	logx.Infof("Received Sender Configuration Update!")
}

func (c *SenderConfig) ValidateSenderConfig() error {
	return nil
}

func GetSenderConfig() *SenderConfig {
	return senderConfig
}
