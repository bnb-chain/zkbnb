package transaction

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/utils"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetTxsByAccountIndexAndTxTypeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTxsByAccountIndexAndTxTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTxsByAccountIndexAndTxTypeLogic {
	return &GetTxsByAccountIndexAndTxTypeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTxsByAccountIndexAndTxTypeLogic) GetTxsByAccountIndexAndTxType(req *types.ReqGetTxsByAccountIndexAndTxType) (*types.RespGetTxsByAccountIndexAndTxType, error) {
	txDetails, err := l.svcCtx.TxDetailModel.GetTxDetailByAccountIndex(int64(req.AccountIndex))
	if err != nil {
		if err != errorcode.DbErrNotFound {
			return nil, errorcode.AppErrInternal
		}
	}
	resp := &types.RespGetTxsByAccountIndexAndTxType{
		Txs: make([]*types.Tx, 0),
	}
	for _, txDetail := range txDetails {
		tx, err := l.svcCtx.TxModel.GetTxByTxId(txDetail.TxId)
		if err != nil {
			return nil, errorcode.AppErrInternal
		}
		if tx.TxType == int64(req.TxType) {
			resp.Total = resp.Total + 1
			resp.Txs = append(resp.Txs, utils.GormTx2Tx(tx))
		}
	}
	memPoolTxDetails, err := l.svcCtx.MempoolDetailModel.GetMempoolTxDetailsByAccountIndex(int64(req.AccountIndex))
	if err != nil {
		if err != errorcode.DbErrNotFound {
			return nil, errorcode.AppErrInternal
		}
	}
	for _, txDetail := range memPoolTxDetails {
		tx, err := l.svcCtx.MempoolModel.GetMempoolTxByTxId(txDetail.TxId)
		if err != nil {
			return nil, errorcode.AppErrInternal
		}
		if tx.TxType == int64(req.TxType) {
			resp.Total = resp.Total + 1
			resp.Txs = append(resp.Txs, utils.MempoolTx2Tx(tx))
		}
	}
	return resp, nil
}
