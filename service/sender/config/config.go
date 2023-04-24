package config

import (
	"encoding/json"
	"errors"
	"github.com/bnb-chain/zkbnb/common/apollo"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"os"
)

const (
	KmsCommitKeyId = "KMS_COMMIT_KEY_ID"
	KmsVerifyKeyId = "KMS_VERIFY_KEY_ID"

	AuthCommitBlockSk = "AUTH_COMMIT_BLOCK_SK"
	AuthVerifyBlockSk = "AUTH_COMMIT_BLOCK_SK"
)

const (
	SenderAppId     = "zkbnb-sender"
	SystemConfigKey = "SystemConfig"
	Namespace       = "application"
)

type ChainConfig struct {
	NetworkRPCSysConfigName string
	MaxWaitingTime          int64
	ConfirmBlocksCount      uint64
	GasLimit                uint64
	GasPrice                uint64
	//((MaxGasPrice-GasPrice)/GasPrice)*100
	MaxGasPriceIncreasePercentage uint64 `json:",optional"`
	DisableCommitBlock            bool   `json:",optional"`
	DisableVerifyBlock            bool   `json:",optional"`
}

type AuthConfig struct {
	CommitBlockSk string `json:",optional"`
	VerifyBlockSk string `json:",optional"`
}

type KMSConfig struct {
	CommitKeyId string `json:",optional"`
	VerifyKeyId string `json:",optional"`
}

type Config struct {
	Postgres    apollo.Postgres
	ChainConfig ChainConfig
	AuthConfig  AuthConfig
	KMSConfig   KMSConfig
	LogConf     logx.LogConf
}

func InitSystemConfiguration(config *Config, configFile string) error {
	if err := InitSystemConfigFromEnvironment(config); err != nil {
		logx.Errorf("Init system configuration from environment raise error: %v", err)
	} else {
		logx.Infof("Init system configuration from environment Successfully")
		return nil
	}
	if err := InitSystemConfigFromConfigFile(config, configFile); err != nil {
		logx.Errorf("Init system configuration from config file raise error: %v", err)
		panic("Init system configuration from config file raise error:" + err.Error())
	} else {
		logx.Infof("Init system configuration from config file Successfully")
		return nil
	}
}

func InitSystemConfigFromEnvironment(c *Config) error {
	commonConfig, err := apollo.InitCommonConfig()
	if err != nil {
		return err
	}
	c.Postgres = commonConfig.Postgres

	systemConfigString, err := apollo.LoadApolloConfigFromEnvironment(SenderAppId, Namespace, SystemConfigKey)
	if err != nil {
		return err
	}

	systemConfig := &Config{}
	err = json.Unmarshal([]byte(systemConfigString), systemConfig)
	if err != nil {
		return err
	}

	c.ChainConfig = systemConfig.ChainConfig
	c.LogConf = systemConfig.LogConf

	if err = InitKeyConfigFromEnvironment(c); err != nil {
		return err
	}

	return nil
}

func InitKeyConfigFromEnvironment(c *Config) error {
	// Load and check all the kms key Ids from environment variables
	commitKeyId := os.Getenv(KmsCommitKeyId)
	verifyKeyId := os.Getenv(KmsVerifyKeyId)
	if len(commitKeyId) > 0 && len(verifyKeyId) > 0 {
		c.KMSConfig.CommitKeyId = commitKeyId
		c.KMSConfig.VerifyKeyId = verifyKeyId
		return nil
	}

	// Load and check all the private secrets from environment variables
	commitBlockSk := os.Getenv(AuthCommitBlockSk)
	verifyBlockSk := os.Getenv(AuthVerifyBlockSk)
	if len(commitBlockSk) > 0 && len(verifyBlockSk) > 0 {
		c.AuthConfig.CommitBlockSk = commitBlockSk
		c.AuthConfig.VerifyBlockSk = verifyBlockSk
		return nil
	}

	// If both kms keys and private keys have not been set in the environment, directly return this error
	return errors.New("both kms keys and auth private keys not set in the environment variables")
}

func InitSystemConfigFromConfigFile(c *Config, configFile string) error {
	conf.Load(configFile, c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()
	return nil
}
