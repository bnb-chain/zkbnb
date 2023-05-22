package info

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbnb/types"
)

type GetLayer2BasicInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetLayer2BasicInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLayer2BasicInfoLogic {
	return &GetLayer2BasicInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

var (
	contractNames = []string{
		"ZkBNBContract",
		"GovernanceContract",
		"DefaultNftFactory",
		"AssetGovernanceContract",
	}
)

func (l *GetLayer2BasicInfoLogic) GetLayer2BasicInfo(fromCache bool) (*types.Layer2BasicInfo, error) {
	resp := &types.Layer2BasicInfo{
		ContractAddresses: make([]types.ContractAddress, 0, len(contractNames)),
	}
	var err error
	resp.BlockCommitted, err = l.svcCtx.MemCache.GetCommittedBlocksCountWithFallback(fromCache, func() (interface{}, error) {
		return l.svcCtx.BlockModel.GetCommittedBlocksCount()
	})

	if err != nil {
		if err != types2.DbErrNotFound {
			return nil, types2.AppErrInternal
		}
	}
	resp.BlockVerified, err = l.svcCtx.MemCache.GetVerifiedBlocksCountWithFallback(fromCache, func() (interface{}, error) {
		return l.svcCtx.BlockModel.GetVerifiedBlocksCount()

	})
	if err != nil {
		if err != types2.DbErrNotFound {
			return nil, types2.AppErrInternal
		}
	}
	resp.TotalTransactionCount, err = l.svcCtx.MemCache.GetTxTotalCountWithFallback(fromCache, func() (interface{}, error) {
		return l.svcCtx.TxModel.GetTxsTotalCount()
	})
	if err != nil {
		if err != types2.DbErrNotFound {
			return nil, types2.AppErrInternal
		}
	}

	now := time.Now()
	today := now.Round(24 * time.Hour).Add(-8 * time.Hour)

	resp.YesterdayTransactionCount, err = l.svcCtx.MemCache.GetTxsTotalCountYesterdayBetweenWithFallback(fromCache, func() (interface{}, error) {
		return l.svcCtx.TxModel.GetTxsTotalCountBetween(today.Add(-24*time.Hour), today)
	})
	if err != nil {
		if err != types2.DbErrNotFound {
			return nil, types2.AppErrInternal
		}
	}
	resp.TodayTransactionCount, err = l.svcCtx.MemCache.GetTxsTotalCountTodayBetweenWithFallback(fromCache, func() (interface{}, error) {
		return l.svcCtx.TxModel.GetTxsTotalCountBetween(today, now)
	})
	if err != nil {
		if err != types2.DbErrNotFound {
			return nil, types2.AppErrInternal
		}
	}
	resp.YesterdayActiveUserCount, err = l.svcCtx.MemCache.GetDistinctAccountsCountYesterdayBetweenWithFallback(fromCache, func() (interface{}, error) {
		return l.svcCtx.TxModel.GetDistinctAccountsCountBetween(today.Add(-24*time.Hour), today)
	})
	if err != nil {
		if err != types2.DbErrNotFound {
			return nil, types2.AppErrInternal
		}
	}
	resp.TodayActiveUserCount, err = l.svcCtx.MemCache.GetDistinctAccountsCountTodayBetweenWithFallback(fromCache, func() (interface{}, error) {
		return l.svcCtx.TxModel.GetDistinctAccountsCountBetween(today, now)
	})

	if err != nil {
		if err != types2.DbErrNotFound {
			return nil, types2.AppErrInternal
		}
	}
	for _, contractName := range contractNames {
		contract, err := l.svcCtx.MemCache.GetSysConfigWithFallback(contractName, fromCache, func() (interface{}, error) {
			return l.svcCtx.SysConfigModel.GetSysConfigByName(contractName)
		})
		if err != nil {
			if err != types2.DbErrNotFound {
				return nil, types2.AppErrInternal
			}
			logx.Errorf("there is no contractName=%s at SysConfig table,pls check", contractName)
			continue
		}
		resp.ContractAddresses = append(resp.ContractAddresses,
			types.ContractAddress{Name: contractName, Address: contract.Value})
	}
	return resp, nil
}
