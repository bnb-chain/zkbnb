package transaction

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/utils"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetTxsByAccountNameLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTxsByAccountNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTxsByAccountNameLogic {
	return &GetTxsByAccountNameLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTxsByAccountNameLogic) GetTxsByAccountName(req *types.ReqGetTxsByAccountName) (*types.RespGetTxsByAccountName, error) {
	if !utils.ValidateAccountName(req.AccountName) {
		logx.Errorf("invalid AccountName: %s", req.AccountName)
		return nil, errorcode.AppErrInvalidParam.RefineError("invalid AccountName")
	}

	account, err := l.svcCtx.AccountModel.GetAccountByAccountName(req.AccountName)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}
	txIds, err := l.svcCtx.TxDetailModel.GetTxIdsByAccountIndex(account.AccountIndex)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}
	resp := &types.RespGetTxsByAccountName{
		Total: uint32(len(txIds)),
		Txs:   make([]*types.Tx, 0),
	}
	if req.Offset > resp.Total {
		return resp, nil
	}
	end := req.Offset + req.Limit
	if resp.Total < (req.Offset + req.Limit) {
		end = resp.Total
	}
	for _, id := range txIds[req.Offset:end] {
		tx, err := l.svcCtx.TxModel.GetTxByTxId(id)
		if err != nil {
			return nil, err
		}
		resp.Txs = append(resp.Txs, utils.GormTx2Tx(tx))
	}
	return resp, nil
}
