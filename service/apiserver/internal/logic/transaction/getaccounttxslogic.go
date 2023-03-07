package transaction

import (
	"context"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/logic/utils"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbnb/types"
)

const (
	queryByAccountIndex = "account_index"
	queryByL1Address    = "l1_address"
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
	resp = &types.Txs{
		Txs: make([]*types.Tx, 0, req.Limit),
	}

	accountIndex := int64(0)
	switch req.By {
	case queryByAccountIndex:
		accountIndex, err = strconv.ParseInt(req.Value, 10, 64)
		if err != nil || accountIndex < 0 {
			return nil, types2.AppErrInvalidAccountIndex
		}
	case queryByL1Address:
		accountIndex, err = l.svcCtx.MemCache.GetAccountIndexByL1Address(req.Value)
	case queryByAccountPk:
		accountIndex, err = l.svcCtx.MemCache.GetAccountIndexByPk(req.Value)
	default:
		return nil, types2.AppErrInvalidParam.RefineError("param by should be account_index|account_name|account_pk")
	}

	if err != nil {
		if err == types2.DbErrNotFound {
			return resp, nil
		}
		return nil, types2.AppErrInternal
	}

	options := []tx.GetTxOptionFunc{}
	if len(req.Types) > 0 {
		options = append(options, tx.GetTxWithTypes(req.Types))
	}

	total, err := l.svcCtx.TxModel.GetTxsCountByAccountIndex(accountIndex, options...)
	if err != nil {
		return nil, types2.AppErrInternal
	}

	resp.Total = uint32(total)
	if total == 0 || total <= int64(req.Offset) {
		return resp, nil
	}

	txs, err := l.svcCtx.TxModel.GetTxsByAccountIndex(accountIndex, int64(req.Limit), int64(req.Offset), options...)
	if err != nil {
		return nil, types2.AppErrInternal
	}

	for _, dbTx := range txs {
		tx := utils.ConvertTx(dbTx)
		tx.L1Address, _ = l.svcCtx.MemCache.GetL1AddressByIndex(tx.AccountIndex)
		tx.AssetName, _ = l.svcCtx.MemCache.GetAssetNameById(tx.AssetId)
		if tx.ToAccountIndex >= 0 {
			tx.ToL1Address, _ = l.svcCtx.MemCache.GetL1AddressByIndex(tx.ToAccountIndex)
		}
		resp.Txs = append(resp.Txs, tx)
	}
	return resp, nil
}
