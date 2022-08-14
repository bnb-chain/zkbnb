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

type GetTxsByAccountIndexLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTxsByAccountIndexLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTxsByAccountIndexLogic {
	return &GetTxsByAccountIndexLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTxsByAccountIndexLogic) GetTxsByAccountIndex(req *types.ReqGetTxsByAccountIndex) (resp *types.RespGetTxsByAccountIndex, err error) {
	resp = &types.RespGetTxsByAccountIndex{
		Txs: make([]*types.Tx, 0),
	}

	total := int64(0)
	txs := make([]*tx.Tx, 0)
	if req.TxType > 0 {
		total, err = l.svcCtx.TxModel.GetTxsCountByAccountIndexTxType(int64(req.AccountIndex), int64(req.TxType))
		if err != nil {
			if err != errorcode.DbErrNotFound {
				return nil, errorcode.AppErrInternal
			}
		}
		resp.Total = uint32(total)
		if total == 0 || total <= int64(req.Offset) {
			return resp, nil
		}

		txs, err = l.svcCtx.TxModel.GetTxsListByAccountIndexTxType(int64(req.AccountIndex), int64(req.TxType), int64(req.Limit), int64(req.Offset))
		if err != nil {
			if err != errorcode.DbErrNotFound {
				return nil, errorcode.AppErrInternal
			}
		}
	} else {
		total, err = l.svcCtx.TxModel.GetTxsCountByAccountIndex(int64(req.AccountIndex))
		if err != nil {
			if err != errorcode.DbErrNotFound {
				return nil, errorcode.AppErrInternal
			}
		}

		resp.Total = uint32(total)
		if total == 0 || total <= int64(req.Offset) {
			return resp, nil
		}

		txs, err = l.svcCtx.TxModel.GetTxsListByAccountIndex(int64(req.AccountIndex), int64(req.Limit), int64(req.Offset))
		if err != nil {
			if err != errorcode.DbErrNotFound {
				return nil, errorcode.AppErrInternal
			}
		}
	}

	for _, t := range txs {
		tx := utils.DbTx2Tx(t)
		tx.AccountName, _ = l.svcCtx.MemCache.GetAccountNameByIndex(tx.AccountIndex)
		resp.Txs = append(resp.Txs, tx)
	}
	return resp, nil
}
