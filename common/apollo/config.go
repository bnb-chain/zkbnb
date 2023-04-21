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

	MasterSecretKey = "MASTER_SECRET_KEY"
	SlaveSecretKey  = "SLAVE_SECRET_KEY"

	DBConnectionFormat = "host=%s user=%s password=%s dbname=%s port=%s sslmode=disable"

	Username = "username"
	Password = "password"
	Engine   = "engine"
	Dbname   = "dbname"
	Host     = "host"
	Port     = "port"
)

type Postgres struct {
	MasterDataSource string          `json:",optional"`
	SlaveDataSource  string          `json:",optional"`
	LogLevel         logger.LogLevel `json:",optional"`
	MaxIdle          int             `json:",optional"`
	MaxConn          int             `json:",optional"`
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

		// Load the postgresql connection parameter from the secret manager
		/*postgresql, err := LoadPostgresqlConfig()
		if err != nil {
			return nil, err
		}
		commonConfig.Postgres.MasterDataSource = postgresql.MasterDataSource
		commonConfig.Postgres.SlaveDataSource = postgresql.SlaveDataSource*/

		return commonConfig, nil
	}
}

func LoadPostgresqlConfig() (*Postgres, error) {
	masterSecretKey := os.Getenv(MasterSecretKey)
	slaveSecretKey := os.Getenv(SlaveSecretKey)

	postgres := &Postgres{}
	masterDataMap, err := secret.LoadSecretData(masterSecretKey)
	if err != nil {
		return nil, err
	}

	masterUsername := masterDataMap[Username]
	masterPassword := masterDataMap[Password]
	masterEngine := masterDataMap[Engine]
	masterDbname := masterDataMap[Dbname]
	masterHost := masterDataMap[Host]
	masterPort := masterDataMap[Port]

	if len(masterUsername) == 0 || len(masterPassword) == 0 || len(masterEngine) == 0 ||
		len(masterDbname) == 0 || len(masterHost) == 0 || len(masterPort) == 0 {
		return nil, errors.New("master datasource configuration is not correct in secret manager, please check it again")
	}
	masterConnectionString := fmt.Sprintf(DBConnectionFormat, masterHost, masterUsername, masterPassword, masterDbname, masterPort)
	postgres.MasterDataSource = masterConnectionString

	slaveDataMap, err := secret.LoadSecretData(slaveSecretKey)
	if err != nil {
		return nil, err
	}

	slaveUsername := slaveDataMap[Username]
	slavePassword := slaveDataMap[Password]
	slaveEngine := slaveDataMap[Engine]
	slaveDbname := slaveDataMap[Dbname]
	slaveHost := slaveDataMap[Host]
	slavePort := slaveDataMap[Port]

	if len(slaveUsername) == 0 || len(slavePassword) == 0 || len(slaveEngine) == 0 ||
		len(slaveDbname) == 0 || len(slaveHost) == 0 || len(slavePort) == 0 {
		return nil, errors.New("slave datasource configuration is not correct in secret manager, please check it again")
	}
	slaveConnectionString := fmt.Sprintf(DBConnectionFormat, slaveHost, slaveUsername, slavePassword, slaveDbname, slavePort)
	postgres.SlaveDataSource = slaveConnectionString

	logx.Info("Load Postgresql database connection configuration from the secret manager successfully")

	return postgres, nil
}

func LoadApolloConfigFromEnvironment(appId, namespace, configKey string) (string, error) {
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
