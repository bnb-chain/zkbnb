package apollo

import (
	"encoding/json"
	"errors"
	"github.com/apolloconfig/agollo/v4"
	apollo "github.com/apolloconfig/agollo/v4/env/config"
	"github.com/apolloconfig/agollo/v4/storage"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"os"
)

const (
	CommonAppId = "CommonAppId"
	Cluster     = "ApolloCluster"
	Endpoint    = "ApolloEndpoint"
	Namespace   = "ApolloNamespace"

	CommonConfigKey = "CommonConfig"
)

type Postgres struct {
	MasterDataSource string
	SlaveDataSource  string
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
var apolloConfigMap = make(map[string]*apollo.AppConfig)

func InitCommonConfig() (*CommonConfig, error) {
	if commonConfigString, err := LoadApolloConfigFromEnvironment(CommonAppId, CommonConfigKey); err != nil {
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

func LoadApolloConfigFromEnvironment(appIdKey, configKey string) (string, error) {

	// Initiate the apollo client instance for apollo configuration operation
	apolloClient, err := InitApolloClientInstance(appIdKey)
	if err != nil {
		return "", err
	}

	apolloConfig := apolloConfigMap[appIdKey]
	// Get the latest common configuration from the apollo client
	apolloCache := apolloClient.GetConfigCache(apolloConfig.NamespaceName)
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

func InitApolloClientInstance(appIdKey string) (agollo.Client, error) {
	if client := apolloClientMap[appIdKey]; client == nil {
		// Load and check all the apollo parameters from environment variables
		appId := os.Getenv(appIdKey)
		if len(appId) == 0 {
			return nil, errors.New("appId is not set in environment variables")
		}
		cluster := os.Getenv(Cluster)
		if len(cluster) == 0 {
			return nil, errors.New("apolloCluster is not set in environment variables")
		}
		endpoint := os.Getenv(Endpoint)
		if len(endpoint) == 0 {
			return nil, errors.New("apolloEndpoint is not set in environment variables")
		}
		namespace := os.Getenv(Namespace)
		if len(namespace) == 0 {
			return nil, errors.New("apolloNamespace is not set in environment variables")
		}

		// Construct the apollo config for creating apollo client
		apolloConfigMap[appIdKey] = &apollo.AppConfig{
			AppID:          appId,
			Cluster:        cluster,
			IP:             endpoint,
			NamespaceName:  namespace,
			IsBackupConfig: true,
		}

		// Create the apollo client here for getting the latest configuration
		client, err := agollo.StartWithConfig(func() (*apollo.AppConfig, error) {
			return apolloConfigMap[appIdKey], nil
		})
		if err != nil {
			return nil, err
		}
		apolloClientMap[appIdKey] = client
	}
	return apolloClientMap[appIdKey], nil
}

func AddChangeListener(appIdKey string, listener storage.ChangeListener) {
	apolloClientMap[appIdKey].AddChangeListener(listener)
}
