package info

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

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
	contractAddressesNames = []string{
		"ZkbasContract",
		"ZnsPriceOracle",
	}
)

func (l *GetLayer2BasicInfoLogic) GetLayer2BasicInfo(_ *types.ReqGetLayer2BasicInfo) (*types.RespGetLayer2BasicInfo, error) {
	resp := &types.RespGetLayer2BasicInfo{
		ContractAddresses: make([]string, 0),
	}
	var err error
	resp.BlockCommitted, err = l.svcCtx.BlockModel.GetCommittedBlocksCount()
	if err != nil {
		logx.Errorf("[GetCommittedBlocksCount] err: %s", err.Error())
		return nil, err
	}
	resp.BlockVerified, err = l.svcCtx.BlockModel.GetVerifiedBlocksCount()
	if err != nil {
		logx.Errorf("[GetVerifiedBlocksCount] err: %s", err.Error())
		return nil, err
	}
	resp.TotalTransactions, err = l.svcCtx.TxModel.GetTxsTotalCount()
	if err != nil {
		logx.Errorf("[GetTxsTotalCount] err: %s", err.Error())
		return nil, err
	}

	now := time.Now()
	today := now.Round(24 * time.Hour).Add(-8 * time.Hour)

	resp.TransactionsCountYesterday, err = l.svcCtx.TxModel.GetTxsTotalCountBetween(today.Add(-24*time.Hour), today)
	if err != nil {
		logx.Errorf("[GetTxCountByTimeRange] err: %s", err.Error())
		return nil, err
	}
	resp.TransactionsCountToday, err = l.svcCtx.TxModel.GetTxsTotalCountBetween(today, now)
	if err != nil {
		logx.Errorf("[GetTxCountByTimeRange] err: %s", err.Error())
		return nil, err
	}
	resp.DauYesterday, err = l.svcCtx.TxDetailModel.GetDauInTxDetailBetween(today.Add(-24*time.Hour), today)
	if err != nil {
		logx.Errorf("[GetDauInTxDetail] err: %s", err.Error())
		return nil, err
	}
	resp.DauToday, err = l.svcCtx.TxDetailModel.GetDauInTxDetailBetween(today, now)
	if err != nil {
		logx.Errorf("[GetDauInTxDetail] err: %s", err.Error())
		return nil, err
	}
	for _, contractAddressesName := range contractAddressesNames {
		contract, err := l.svcCtx.SysConfigModel.GetSysconfigByName(contractAddressesName)
		if err != nil {
			logx.Errorf("[GetSysconfigByName] err: %s", err.Error())
			return nil, err
		}
		resp.ContractAddresses = append(resp.ContractAddresses, contract.Value)
	}
	return resp, nil
}
