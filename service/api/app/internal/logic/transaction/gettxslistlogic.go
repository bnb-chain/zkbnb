package transaction

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/utils"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetTxsListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTxsListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTxsListLogic {
	return &GetTxsListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTxsListLogic) GetTxsList(req *types.ReqGetTxsList) (resp *types.RespGetTxsList, err error) {
	count, err := l.svcCtx.TxModel.GetTxsTotalCount()
	if err != nil {
		if err != errorcode.DbErrNotFound {
			return nil, errorcode.AppErrInternal
		}
	}
	txs := make([]*types.Tx, 0)
	if count > 0 {
		list, err := l.svcCtx.TxModel.GetTxsList(int64(req.Limit), int64(req.Offset))
		if err != nil {
			return nil, errorcode.AppErrInternal
		}
		for _, t := range list {
			tx := utils.GormTx2Tx(t)
			txs = append(txs, tx)
		}
	}
	resp = &types.RespGetTxsList{
		Total: uint32(count),
		Txs:   txs,
	}
	return resp, nil
}
