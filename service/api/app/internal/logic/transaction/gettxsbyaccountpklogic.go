package transaction

import (
	"context"

	"github.com/bnb-chain/zkbas/common/model/tx"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/utils"

	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTxsByAccountPkLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTxsByAccountPkLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTxsByAccountPkLogic {
	return &GetTxsByAccountPkLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTxsByAccountPkLogic) GetTxsByAccountPk(req *types.ReqGetTxsByAccountPk) (resp *types.RespGetTxsByAccountPk, err error) {
	if !utils.ValidateAccountPk(req.AccountPk) {
		logx.Errorf("invalid AccountPk: %s", req.AccountPk)
		return nil, errorcode.AppErrInvalidParam.RefineError("invalid AccountPk")
	}

	account, err := l.svcCtx.AccountModel.GetAccountByPk(req.AccountPk)
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

	resp = &types.RespGetTxsByAccountPk{
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
