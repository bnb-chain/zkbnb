package config

import (
	"encoding/json"
	"github.com/apolloconfig/agollo/v4/storage"
	"github.com/bnb-chain/zkbnb/common/apollo"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	SenderConfigKey = "SenderConfig"
)

var senderConfig = &SenderConfig{}
var senderUpdater = &SenderUpdater{}

type SenderConfig struct {
	CommitControlSwitch bool
	VerifyControlSwitch bool

	MaxCommitBlockCount uint64
	CommitTxCountLimit  uint64

	MaxCommitTotalGasFee uint64

	MaxVerifyBlockCount uint64
	VerifyTxCountLimit  uint64

	MaxVerifyTotalGasFee uint64

	MaxCommitTxCount uint64
	MaxVerifyTxCount uint64

	MaxCommitBlockInterval uint64
	MaxVerifyBlockInterval uint64

	CommitAvgUnitGasSwitch bool
	VerifyAvgUnitGasSwitch bool

	MaxCommitAvgUnitGas uint64
	MaxVerifyAvgUnitGas uint64
}

type SenderUpdater struct {
}

func InitSenderConfiguration(c Config) {
	//Add the apollo configuration updater listener for SenderConfig
	apollo.AddChangeListener(SenderAppId, Namespace, senderUpdater)

	newSenderConfig := &SenderConfig{}
	newSenderConfigString, err := apollo.LoadApolloConfigFromEnvironment(SenderAppId, Namespace, SenderConfigKey)
	if err != nil {
		// If fails to initiate sender strategy configuration from apollo, directly switch it off
		logx.Severef("Fail to Initiate Sender Configuration from the apollo server!")
		newSenderConfig.CommitControlSwitch = false
		newSenderConfig.VerifyControlSwitch = false
	} else {
		err := json.Unmarshal([]byte(newSenderConfigString), newSenderConfig)
		if err != nil {
			logx.Errorf("Fail to update SenderConfig from the apollo server, Reason:%s", err.Error())
			panic("Fail to update SenderConfig from the apollo server, Reason:" + err.Error())
		}

		// Validate the Sender Configuration from the apollo server side
		if err = newSenderConfig.ValidateSenderConfig(); err != nil {
			logx.Severef("Fail to validate SenderConfig from the apollo server, Reason:%s", err.Error())
			panic("Fail to validate SenderConfig from the apollo server!")
		}
		senderConfig = newSenderConfig

		logx.Info("Initiate and load SenderConfig Successfully!")
		logx.Info("SenderConfig:", newSenderConfigString)
	}
}

func (u *SenderUpdater) OnChange(event *storage.ChangeEvent) {

	logx.Info("Get updates from apollo server")

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
			return
		}

		// Validate the Sender Configuration from the apollo server side
		if err = newSenderConfig.ValidateSenderConfig(); err != nil {
			logx.Severef("Fail to validate SenderConfig from the apollo server, Reason:%s", err.Error())
			return
		}
		senderConfig = newSenderConfig
		logx.Info("Update SenderConfig Successfully:", newSenderConfigObjectJson)
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
