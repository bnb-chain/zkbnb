package rpc_client

import (
	"github.com/bnb-chain/zkbnb-eth-rpc/rpc"
	"github.com/zeromicro/go-zero/core/logx"
)

const Online = 1
const Offline = -1

type ClientConfig struct {
	endpoint string
}

type RpcClient struct {
	client   *rpc.ProviderClient
	endpoint string
	status   int
	height   uint64
}

func NewRpcClient(endpoint string) (*RpcClient, error) {
	s := &RpcClient{
		endpoint: endpoint,
	}

	var err error
	s.client, err = rpc.NewClient(endpoint)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (w *RpcClient) healthCheck() {
	latestHeight, err := w.client.GetHeight()
	if err != nil {
		w.status = Offline
		logx.Severef("Node Provider is offline,endpoint=%s, %v", w.endpoint, err)
		return
	}
	w.status = Online
	w.height = latestHeight
	logx.Infof("Health check success,endpoint=%s,latestHeight=%d", w.endpoint, latestHeight)
}

func (w *RpcClient) deepCopy() *RpcClient {
	return &RpcClient{client: w.client, endpoint: w.endpoint, status: w.status, height: w.height}
}
