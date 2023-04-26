package transaction

import (
	"context"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/logic/utils"
	types2 "github.com/bnb-chain/zkbnb/types"
	"strconv"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMergedAccountTxsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetMergedAccountTxsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMergedAccountTxsLogic {
	return &GetMergedAccountTxsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetMergedAccountTxsLogic) GetMergedAccountTxs(req *types.ReqGetAccountTxs) (resp *types.Txs, err error) {
	resp = &types.Txs{
		Txs: make([]*types.Tx, 0, req.Limit),
	}

	accountIndex, err := l.fetchAccountIndexFromReq(req)
	if err != nil {
		if err == types2.DbErrNotFound {
			return resp, nil
		}
		return nil, err
	}

	var options []tx.GetTxOptionFunc
	if len(req.Types) > 0 {
		options = append(options, tx.GetTxWithTypes(req.Types))
	}

	totalTxCount, err := l.svcCtx.TxPoolModel.GetTxCountUnscopedByAccountIndex(accountIndex, options...)
	if err != nil {
		if err == types2.DbErrNotFound {
			return resp, nil
		}
		return nil, err
	}

	poolTxs, err := l.svcCtx.TxPoolModel.GetTxsUnscopedByAccountIndex(accountIndex, int64(req.Limit), int64(req.Offset), options...)
	if err != nil {
		if err == types2.DbErrNotFound {
			return resp, nil
		}
		return nil, err
	}

	poolTxIds := l.preparePoolTxIds(poolTxs)
	txs, err := l.svcCtx.TxModel.GetTxsByPoolTxIds(poolTxIds)
	if err != nil && err != types2.DbErrNotFound {
		return nil, err
	}

	resultTxList := l.mergeTxsList(poolTxs, txs)
	resp = &types.Txs{
		Txs:   resultTxList,
		Total: uint32(totalTxCount),
	}
	return resp, nil
}

func (l *GetMergedAccountTxsLogic) fetchAccountIndexFromReq(req *types.ReqGetAccountTxs) (int64, error) {
	switch req.By {
	case queryByAccountIndex:
		accountIndex, err := strconv.ParseInt(req.Value, 10, 64)
		if err != nil || accountIndex < 0 {
			return accountIndex, types2.AppErrInvalidAccountIndex
		}
		return accountIndex, err
	case queryByL1Address:
		accountIndex, err := l.svcCtx.MemCache.GetAccountIndexByL1Address(req.Value)
		return accountIndex, err
	}
	return 0, types2.AppErrInvalidParam.RefineError("param by should be account_index|l1_address")
}

func (l *GetMergedAccountTxsLogic) mergeTxsList(poolTxList []*tx.Tx, txList []*tx.Tx) []*types.Tx {
	resultTxList := make([]*types.Tx, 0, len(poolTxList))
	for _, poolTx := range poolTxList {
		var resultTx *types.Tx
		if tx := l.getTxDataByPoolTxId(txList, poolTx.ID); tx != nil {
			resultTx = utils.ConvertTx(tx)
		} else {
			resultTx = utils.ConvertTx(poolTx)
		}
		if resultTx.AccountIndex >= 0 {
			resultTx.L1Address, _ = l.svcCtx.MemCache.GetL1AddressByIndex(resultTx.AccountIndex)
		}
		resultTx.AssetName, _ = l.svcCtx.MemCache.GetAssetNameById(resultTx.AssetId)
		if resultTx.ToAccountIndex >= 0 {
			resultTx.ToL1Address, _ = l.svcCtx.MemCache.GetL1AddressByIndex(resultTx.ToAccountIndex)
		}
		resultTxList = append(resultTxList, resultTx)
	}
	return resultTxList
}

func (l *GetMergedAccountTxsLogic) getTxDataByPoolTxId(txList []*tx.Tx, poolTxId uint) *tx.Tx {
	if txList != nil && len(txList) > 0 {
		for _, tx := range txList {
			if tx.PoolTxId == poolTxId {
				return tx
			}
		}
	}
	return nil
}

func (l *GetMergedAccountTxsLogic) preparePoolTxIds(txs []*tx.Tx) []int64 {
	poolTxIds := make([]int64, 0, len(txs))
	if len(txs) > 0 {
		for _, tx := range txs {
			poolTxIds = append(poolTxIds, int64(tx.ID))
		}
	}
	return poolTxIds
}
