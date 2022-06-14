package block

import (
	"context"
	"fmt"

	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBlockByCommitmentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetBlockByCommitmentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBlockByCommitmentLogic {
	return &GetBlockByCommitmentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetBlockByCommitmentLogic) GetBlockByCommitment(req *types.ReqGetBlockByCommitment) (resp *types.RespGetBlockByCommitment, err error) {
	// query basic block info
	block, err := l.svcCtx.Block.GetBlockByCommitment(req.BlockCommitment)
	if err != nil {
		err = fmt.Errorf("[explorer.block.GetBlockByCommitment]<=>%s", err.Error())
		l.Error(err)
		return
	}

	txs := make([]string, 0)
	for _, tx := range block.Txs {
		txs = append(txs, tx.TxHash)
	}
	return
}
