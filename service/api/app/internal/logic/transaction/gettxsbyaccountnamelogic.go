package transaction

import (
	"context"

	"github.com/bnb-chain/zkbas/common/model/tx"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/errorcode"
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

func (l *GetTxsByAccountNameLogic) GetTxsByAccountName(req *types.ReqGetTxsByAccountName) (resp *types.RespGetTxsByAccountName, err error) {
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

	txs := make([]*tx.Tx, 0)
	total, err := l.svcCtx.TxModel.GetTxsCountByAccountIndex(account.AccountIndex)
	if err != nil {
		if err != errorcode.DbErrNotFound {
			return nil, errorcode.AppErrInternal
		}
	}
	if total > 0 && int64(req.Offset) >= total {
		txs, err = l.svcCtx.TxModel.GetTxsListByAccountIndex(account.AccountIndex, int64(req.Limit), int64(req.Offset))
		if err != nil {
			if err != errorcode.DbErrNotFound {
				return nil, errorcode.AppErrInternal
			}
		}
	}

	resp = &types.RespGetTxsByAccountName{
		Total: uint32(total),
		Txs:   make([]*types.Tx, 0),
	}
	for _, tx := range txs {
		if err != nil {
			return nil, errorcode.AppErrInternal
		}
		resp.Txs = append(resp.Txs, utils.GormTx2Tx(tx))
	}
	return resp, nil
}
