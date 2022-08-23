package transaction

import (
	"context"
	"strconv"

	"github.com/bnb-chain/zkbas/service/apiserver/internal/logic/utils"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/types"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/errorcode"
)

const (
	queryByAccountIndex = "account_index"
	queryByAccountName  = "account_name"
	queryByAccountPk    = "account_pk"
)

type GetAccountTxsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAccountTxsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountTxsLogic {
	return &GetAccountTxsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAccountTxsLogic) GetAccountTxs(req *types.ReqGetAccountTxs) (resp *types.Txs, err error) {
	accountIndex := int64(0)
	switch req.By {
	case queryByAccountIndex:
		accountIndex, err = strconv.ParseInt(req.Value, 10, 64)
		if err != nil {
			return nil, errorcode.AppErrInvalidParam.RefineError("invalid value for account_index")
		}
	case queryByAccountName:
		accountIndex, err = l.svcCtx.MemCache.GetAccountIndexByName(req.Value)
	case queryByAccountPk:
		accountIndex, err = l.svcCtx.MemCache.GetAccountIndexByPk(req.Value)
	default:
		return nil, errorcode.AppErrInvalidParam.RefineError("param by should be account_index|account_name|account_pk")
	}

	total, err := l.svcCtx.TxModel.GetTxsCountByAccountIndex(accountIndex)
	if err != nil {
		if err != errorcode.DbErrNotFound {
			return nil, errorcode.AppErrInternal
		}
	}

	resp = &types.Txs{
		Total: uint32(total),
		Txs:   make([]*types.Tx, 0),
	}
	if total == 0 || total <= int64(req.Offset) {
		return resp, nil
	}

	txs, err := l.svcCtx.TxModel.GetTxsListByAccountIndex(accountIndex, int64(req.Limit), int64(req.Offset))
	if err != nil {
		return nil, errorcode.AppErrInternal
	}

	for _, t := range txs {
		tx := utils.DbTx2Tx(t)
		tx.AccountName, _ = l.svcCtx.MemCache.GetAccountNameByIndex(tx.AccountIndex)
		tx.AssetName, _ = l.svcCtx.MemCache.GetAssetNameById(tx.AssetId)
		resp.Txs = append(resp.Txs, tx)
	}
	return resp, nil
}
