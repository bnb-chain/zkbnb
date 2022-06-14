package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	ProverHubRPC zrpc.RpcClientConf
	KeyPath      struct {
		ProvingKeyPath   string
		VerifyingKeyPath string
	}
}
