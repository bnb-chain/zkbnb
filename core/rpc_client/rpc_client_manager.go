package rpc_client

import (
	"fmt"
	"github.com/bnb-chain/zkbnb-eth-rpc/rpc"
	"github.com/bnb-chain/zkbnb/dao/sysconfig"
	"github.com/zeromicro/go-zero/core/logx"
	"sort"
	"strings"
)

var rpcClients = make([]*RpcClient, 0)

func InitRpcClients(sysConfigModel sysconfig.SysConfigModel, rpcSysConfigName string) error {
	l1RPCEndpoint, err := sysConfigModel.GetSysConfigByName(rpcSysConfigName)
	if err != nil {
		logx.Infof("fatal error, failed to get network rpc configuration, err:%v, rpcSysConfigName:%s",
			err, rpcSysConfigName)
	}
	endpoints := strings.Split(l1RPCEndpoint.Value, ",")

	if len(endpoints) == 0 {
		return fmt.Errorf("network rpc endpoints are empty")
	}
	for _, endpoint := range endpoints {
		client, err := NewRpcClient(endpoint)
		if err != nil {
			logx.Severef("fatal error, failed to instance rpc.NewClient, err:%v, endpoint:%s",
				err, endpoint)
			continue
		}
		rpcClients = append(rpcClients, client)
	}
	if len(rpcClients) == 0 {
		return fmt.Errorf("rpc clients are empty")
	}
	logx.Info("RpcClients have been initialized")

	HealthCheck()

	return nil
}

func HealthCheck() {
	count := 0
	for _, rpcClient := range rpcClients {
		rpcClient.healthCheck()
		if rpcClient.status == Offline {
			count++
		}
	}
	if len(rpcClients) == count {
		logx.Severef("all the rpc clients are offline")
	}
}

func GetRpcClient() *rpc.ProviderClient {
	var copyRpcClients []*RpcClient
	for _, rpcClient := range rpcClients {
		copyRpcClients = append(copyRpcClients, rpcClient.deepCopy())
	}
	sort.SliceStable(copyRpcClients, func(i, j int) bool {
		return copyRpcClients[i].height > copyRpcClients[j].height
	})
	sort.SliceStable(copyRpcClients, func(i, j int) bool {
		return copyRpcClients[i].status > copyRpcClients[j].status
	})
	return copyRpcClients[0].client
}
