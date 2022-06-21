package info

import (
	"context"
	"fmt"

	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/block"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/sysconf"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/repo/tx"
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
	tx             tx.Tx
}

func NewGetLayer2BasicInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLayer2BasicInfoLogic {
	return &GetLayer2BasicInfoLogic{
		Logger:         logx.WithContext(ctx),
		ctx:            ctx,
		svcCtx:         svcCtx,
		sysconfigModel: sysconf.New(svcCtx),
		block:          block.New(svcCtx),
		tx:             tx.New(svcCtx),
	}
}

func (l *GetLayer2BasicInfoLogic) GetLayer2BasicInfo(req *types.ReqGetLayer2BasicInfo) (*types.RespGetLayer2BasicInfo, error) {
	resp := &types.RespGetLayer2BasicInfo{}
	errorHandler := func(e error) bool {
		if e != nil {
			err := fmt.Errorf("[explorer.info.GetLayer2BasicInfo]<=>%s", e.Error())
			l.Error(err)
			return true
		}
		return false
	}

	committedBlocksCount, err := l.block.GetCommitedBlocksCount()
	if errorHandler(err) {
		return nil, err
	}
	resp.BlockCommitted = committedBlocksCount

	executedBlocksCount, err := l.block.GetExecutedBlocksCount()
	if errorHandler(err) {
		return nil, err
	}
	resp.BlockExecuted = executedBlocksCount

	txsCount, err := l.tx.GetTxsTotalCount(l.ctx)
	if errorHandler(err) {
		return nil, err
	}
	resp.TotalTransactionsCount = txsCount

	resp.ContractAddresses = make([]string, 0)
	for _, contractAddressesName := range contractAddressesNames {
		contract, err := l.sysconfigModel.GetSysconfigByName(contractAddressesName)
		if errorHandler(err) {
			return nil, err
		}
		resp.ContractAddresses = append(resp.ContractAddresses, contract.Value)
	}
	return resp, nil
}
