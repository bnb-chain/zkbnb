package apollo

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/apolloconfig/agollo/v4"
	apollo "github.com/apolloconfig/agollo/v4/env/config"
	"github.com/apolloconfig/agollo/v4/storage"
	"github.com/bnb-chain/zkbnb/common/secret"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"gorm.io/gorm/logger"
	"os"
)

const (
	CommonAppId     = "zkbnb-common"
	Cluster         = "APOLLO_CLUSTER"
	Endpoint        = "APOLLO_ENDPOINT"
	CommonNamespace = "common.configuration"

	CommonConfigKey = "CommonConfig"
)

type Postgres struct {
	MasterDataSource string
	SlaveDataSource  string
	LogLevel         logger.LogLevel `json:",optional"`
	MaxIdle          int
	MaxConn          int
}

type Apollo struct {
	AppID          string
	Cluster        string
	ApolloIp       string
	Namespace      string
	IsBackupConfig bool
}

type CommonConfig struct {
	Postgres   Postgres
	CacheRedis cache.CacheConf
}

var apolloClientMap = make(map[string]agollo.Client)

func InitCommonConfig() (*CommonConfig, error) {
	if commonConfigString, err := LoadApolloConfigFromEnvironment(CommonAppId, CommonNamespace, CommonConfigKey); err != nil {
		return nil, err
	} else {
		// Convert the configuration value to the common config data
		commonConfig := &CommonConfig{}
		err := json.Unmarshal([]byte(commonConfigString), commonConfig)
		if err != nil {
			return nil, err
		}
		return commonConfig, nil
	}
}

func LoadApolloConfigFromEnvironment(appId, namespace, configKey string) (string, error) {

	secretString := secret.LoadSecretString("zkbnb-qa-pgsql")
	logx.Infof("pgsql database connection:%s", secretString)

	// Initiate the apollo client instance for apollo configuration operation
	apolloClient, err := InitApolloClientInstance(appId, namespace)
	if err != nil {
		return "", err
	}

	// Get the latest common configuration from the apollo client
	apolloCache := apolloClient.GetConfigCache(namespace)
	configObject, err := apolloCache.Get(configKey)
	if err != nil {
		return "", err
	}

	// Convert the configuration value to the common config data
	if configString, ok := configObject.(string); ok {
		return configString, nil
	}

	return "", errors.New("configObject is not valid configuration value, configKey:" + configKey)
}

func InitApolloClientInstance(appId, namespace string) (agollo.Client, error) {
	apolloClientKey := fmt.Sprintf("%s:%s", appId, namespace)
	if client := apolloClientMap[apolloClientKey]; client == nil {
		// Load and check all the apollo parameters from environment variables
		cluster := os.Getenv(Cluster)
		if len(cluster) == 0 {
			return nil, errors.New("apolloCluster is not set in environment variables")
		}
		endpoint := os.Getenv(Endpoint)
		if len(endpoint) == 0 {
			return nil, errors.New("apolloEndpoint is not set in environment variables")
		}

		// Construct the apollo config for creating apollo client
		apolloConfig := &apollo.AppConfig{
			AppID:          appId,
			Cluster:        cluster,
			IP:             endpoint,
			NamespaceName:  namespace,
			IsBackupConfig: true,
		}

		// Create the apollo client here for getting the latest configuration
		client, err := agollo.StartWithConfig(func() (*apollo.AppConfig, error) {
			return apolloConfig, nil
		})
		if err != nil {
			return nil, err
		}
		apolloClientMap[apolloClientKey] = client
	}
	return apolloClientMap[apolloClientKey], nil
}

func AddChangeListener(appId, namespace string, listener storage.ChangeListener) {
	apolloClientKey := fmt.Sprintf("%s:%s", appId, namespace)
	if apolloClient, ok := apolloClientMap[apolloClientKey]; ok {
		apolloClient.AddChangeListener(listener)
	}
}
