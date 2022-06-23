package info

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/block"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/sysconf"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/tx"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/txdetail"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLayer2BasicInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext

	sysconfigModel sysconf.Sysconf
	block          block.Block
	tx             tx.Model
	txDetail       txdetail.Model
}

func NewGetLayer2BasicInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLayer2BasicInfoLogic {
	return &GetLayer2BasicInfoLogic{
		Logger:         logx.WithContext(ctx),
		ctx:            ctx,
		svcCtx:         svcCtx,
		sysconfigModel: sysconf.New(svcCtx),
		block:          block.New(svcCtx),
		tx:             tx.New(svcCtx),
		txDetail:       txdetail.New(svcCtx),
	}
}

func (l *GetLayer2BasicInfoLogic) GetLayer2BasicInfo(req *types.ReqGetLayer2BasicInfo) (*types.RespGetLayer2BasicInfo, error) {
	resp := &types.RespGetLayer2BasicInfo{
		ContractAddresses: make([]string, 0),
	}
	var err error
	resp.BlockCommitted, err = l.block.GetCommitedBlocksCount()
	if err != nil {
		logx.Errorf("[GetCommitedBlocksCount] err:%v", err)
		return nil, err
	}
	resp.BlockVerified, err = l.block.GetVerifiedBlocksCount()
	if err != nil {
		logx.Errorf("[GetVerifiedBlocksCount] err:%v", err)
		return nil, err
	}
	resp.TotalTransactions, err = l.tx.GetTxsTotalCount(l.ctx)
	if err != nil {
		logx.Errorf("[GetTxsTotalCount] err:%v", err)
		return nil, err
	}
	resp.TransactionsCountYesterday, err = l.tx.GetTxCountByTimeRange("yesterday")
	if err != nil {
		logx.Errorf("[GetTxCountByTimeRange] err:%v", err)
		return nil, err
	}
	resp.TransactionsCountToday, err = l.tx.GetTxCountByTimeRange("today")
	if err != nil {
		logx.Errorf("[GetTxCountByTimeRange] err:%v", err)
		return nil, err
	}
	resp.DauYesterday, err = l.txDetail.GetDauInTxDetail(l.ctx, "yesterday")
	if err != nil {
		logx.Errorf("[GetTxCountByTimeRange] err:%v", err)
		return nil, err
	}
	resp.DauToday, err = l.txDetail.GetDauInTxDetail(l.ctx, "today")
	if err != nil {
		logx.Errorf("[GetTxCountByTimeRange] err:%v", err)
		return nil, err
	}
	for _, contractAddressesName := range contractAddressesNames {
		contract, err := l.sysconfigModel.GetSysconfigByName(contractAddressesName)
		if err != nil {
			logx.Errorf("[GetSysconfigByName] err:%v", err)
			return nil, err
		}
		resp.ContractAddresses = append(resp.ContractAddresses, contract.Value)
	}
	return resp, nil
}
