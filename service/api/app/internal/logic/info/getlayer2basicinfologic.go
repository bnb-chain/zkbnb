package info

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
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
		"ZkbasContract",
		"ZnsPriceOracle",
	}
)

func (l *GetLayer2BasicInfoLogic) GetLayer2BasicInfo(_ *types.ReqGetLayer2BasicInfo) (*types.RespGetLayer2BasicInfo, error) {
	resp := &types.RespGetLayer2BasicInfo{
		ContractAddresses: make([]types.ContractAddress, 0),
	}
	var err error
	resp.BlockCommitted, err = l.svcCtx.BlockModel.GetCommittedBlocksCount()
	if err != nil {
		if err != errorcode.DbErrNotFound {
			return nil, errorcode.AppErrInternal
		}
	}
	resp.BlockVerified, err = l.svcCtx.BlockModel.GetVerifiedBlocksCount()
	if err != nil {
		if err != errorcode.DbErrNotFound {
			return nil, errorcode.AppErrInternal
		}
	}
	resp.TotalTransactions, err = l.svcCtx.TxModel.GetTxsTotalCount()
	if err != nil {
		if err != errorcode.DbErrNotFound {
			return nil, errorcode.AppErrInternal
		}
	}

	now := time.Now()
	today := now.Round(24 * time.Hour).Add(-8 * time.Hour)

	resp.TransactionsCountYesterday, err = l.svcCtx.TxModel.GetTxsTotalCountBetween(today.Add(-24*time.Hour), today)
	if err != nil {
		if err != errorcode.DbErrNotFound {
			return nil, errorcode.AppErrInternal
		}
	}
	resp.TransactionsCountToday, err = l.svcCtx.TxModel.GetTxsTotalCountBetween(today, now)
	if err != nil {
		if err != errorcode.DbErrNotFound {
			return nil, errorcode.AppErrInternal
		}
	}
	resp.DauYesterday, err = l.svcCtx.TxModel.GetDistinctAccountCountBetween(today.Add(-24*time.Hour), today)
	if err != nil {
		if err != errorcode.DbErrNotFound {
			return nil, errorcode.AppErrInternal
		}
	}
	resp.DauToday, err = l.svcCtx.TxModel.GetDistinctAccountCountBetween(today, now)
	if err != nil {
		if err != errorcode.DbErrNotFound {
			return nil, errorcode.AppErrInternal
		}
	}
	for _, contractName := range contractNames {
		contract, err := l.svcCtx.SysConfigModel.GetSysConfigByName(contractName)
		if err != nil {
			if err != errorcode.DbErrNotFound {
				return nil, errorcode.AppErrInternal
			}
		}
		resp.ContractAddresses = append(resp.ContractAddresses,
			types.ContractAddress{Name: contractName, Address: contract.Value})
	}
	return resp, nil
}
