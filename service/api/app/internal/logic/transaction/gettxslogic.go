package transaction

import (
	"context"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/utils"

	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTxsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTxsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTxsLogic {
	return &GetTxsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTxsLogic) GetTxs(req *types.ReqGetTxs) (resp *types.RespGetTxs, err error) {
	total, err := l.svcCtx.TxModel.GetTxsTotalCount()
	if err != nil {
		if err != errorcode.DbErrNotFound {
			return nil, errorcode.AppErrInternal
		}
	}
	txs := make([]*types.Tx, 0)
	if total > 0 && total >= int64(req.Offset) {
		list, err := l.svcCtx.TxModel.GetTxsList(int64(req.Limit), int64(req.Offset))
		if err != nil {
			return nil, errorcode.AppErrInternal
		}
		for _, t := range list {
			tx := utils.GormTx2Tx(t)
			txs = append(txs, tx)
		}
	}
	resp = &types.RespGetTxs{
		Total: uint32(total),
		Txs:   txs,
	}
	return resp, nil
}
