package svc

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/config"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/account"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/accounthistory"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/block"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/globalrpc"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/l2asset"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/liquidity"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/mempool"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/sysconf"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/tx"
)

type ServiceContext struct {
	Config config.Config

	SysconfigModel sysconf.Sysconf
	Block          block.Block
	Tx             tx.Tx
	L2AssetInfo    l2asset.L2asset

	Account        account.AccountModel
	AccountHistory accounthistory.AccountHistory
	Liquidity      liquidity.Liquidity
	Mempool        mempool.Mempool
	GlobalRPC      globalrpc.GlobalRPC
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:         c,
		SysconfigModel: sysconf.New(c),
		Block:          block.New(c),
		Tx:             tx.New(c),
		L2AssetInfo:    l2asset.New(c),
		Account:        account.New(c),
		AccountHistory: accounthistory.New(c),
		Liquidity:      liquidity.New(c),
		Mempool:        mempool.New(c),
		GlobalRPC:      globalrpc.New(c, context.Background()),
	}
}
