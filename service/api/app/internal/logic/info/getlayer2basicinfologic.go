package info

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/block"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/sysconf"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/tx"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/txdetail"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetLayer2BasicInfoLogic struct {
	logx.Logger
	ctx            context.Context
	svcCtx         *svc.ServiceContext
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
	resp.BlockCommitted, err = l.block.GetCommitedBlocksCount(l.ctx)
	if err != nil {
		logx.Errorf("[GetCommittedBlocksCount] err:%v", err)
		return nil, err
	}
	resp.BlockVerified, err = l.block.GetVerifiedBlocksCount(l.ctx)
	if err != nil {
		logx.Errorf("[GetVerifiedBlocksCount] err:%v", err)
		return nil, err
	}
	resp.TotalTransactions, err = l.tx.GetTxsTotalCount(l.ctx)
	if err != nil {
		logx.Errorf("[GetTxsTotalCount] err:%v", err)
		return nil, err
	}
	resp.TransactionsCountYesterday, err = l.tx.GetTxCountByTimeRange(l.ctx, "yesterday")
	if err != nil {
		logx.Errorf("[GetTxCountByTimeRange] err:%v", err)
		return nil, err
	}
	resp.TransactionsCountToday, err = l.tx.GetTxCountByTimeRange(l.ctx, "today")
	if err != nil {
		logx.Errorf("[GetTxCountByTimeRange] err:%v", err)
		return nil, err
	}
	resp.DauYesterday, err = l.txDetail.GetDauInTxDetail(l.ctx, "yesterday")
	if err != nil {
		logx.Errorf("[GetDauInTxDetail] err:%v", err)
		return nil, err
	}
	resp.DauToday, err = l.txDetail.GetDauInTxDetail(l.ctx, "today")
	if err != nil {
		logx.Errorf("[GetDauInTxDetail] err:%v", err)
		return nil, err
	}
	for _, contractAddressesName := range contractAddressesNames {
		contract, err := l.sysconfigModel.GetSysconfigByName(l.ctx, contractAddressesName)
		if err != nil {
			logx.Errorf("[GetSysconfigByName] err:%v", err)
			return nil, err
		}
		resp.ContractAddresses = append(resp.ContractAddresses, contract.Value)
	}
	return resp, nil
}
