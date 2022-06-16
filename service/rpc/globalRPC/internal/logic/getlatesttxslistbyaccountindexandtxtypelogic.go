package logic

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/globalRPCProto"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/logic/errcode"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/repo/account"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/repo/block"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/repo/mempool"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/repo/tx"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLatestTxsListByAccountIndexAndTxTypeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	tx      tx.Tx
	mempool mempool.Mempool
	account account.AccountModel
	block   block.Block
}

func NewGetLatestTxsListByAccountIndexAndTxTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLatestTxsListByAccountIndexAndTxTypeLogic {
	return &GetLatestTxsListByAccountIndexAndTxTypeLogic{
		ctx:     ctx,
		svcCtx:  svcCtx,
		Logger:  logx.WithContext(ctx),
		tx:      tx.New(svcCtx.Config),
		mempool: mempool.New(svcCtx),
		account: account.New(svcCtx),
		block:   block.New(svcCtx),
	}
}

func (l *GetLatestTxsListByAccountIndexAndTxTypeLogic) GetLatestTxsListByAccountIndexAndTxType(in *globalRPCProto.ReqGetLatestTxsListByAccountIndexAndTxType) (*globalRPCProto.RespGetLatestTxsListByAccountIndexAndTxType, error) {
	if utils.CheckAccountIndex(in.AccountIndex) {
		logx.Errorf("[CheckAccountIndex] param:%v", in.AccountIndex)
		return nil, errcode.ErrInvalidParam
	}
	if utils.CheckTxType(in.TxType) {
		logx.Errorf("[CheckTxType] param:%v", in.TxType)
		return nil, errcode.ErrInvalidParam
	}
	if utils.CheckTypeLimit(in.Limit) {
		logx.Errorf("[CheckTypeLimit] param:%v", in.Limit)
		return nil, errcode.ErrInvalidParam
	}
	if utils.CheckTypeOffset(in.Offset) {
		logx.Errorf("[CheckTypeOffset] param:%v", in.Offset)
		return nil, errcode.ErrInvalidParam
	}
	txTypeArray, err := GetTxTypeArray(uint(in.TxType))
	if err != nil {
		logx.Errorf("[GetTxTypeArray] err:%v", err)
		return nil, err
	}
	mempoolTxCount, err := l.mempool.GetMempoolTxsTotalCountByAccountIndexAndTxTypeArray(int64(in.AccountIndex), txTypeArray)
	if err != nil {
		logx.Errorf("[GetMempoolTxsTotalCountByAccountIndexAndTxTypeArray] err:%v", err)
		return nil, err
	}
	resp := &globalRPCProto.RespGetLatestTxsListByAccountIndexAndTxType{
		Total:   uint32(mempoolTxCount),
		TxsList: make([]*globalRPCProto.TxInfo, 0),
	}
	var offsetMempool int64
	offsetMempool = int64(in.Offset)
	if mempoolTxCount <= int64(in.Offset) {
		offsetMempool = int64(mempoolTxCount)
	}
	mempoolTxs, err := l.mempool.GetMempoolTxsListByAccountIndexAndTxTypeArray(int64(in.AccountIndex),
		txTypeArray, int64(in.Limit), offsetMempool)
	if err != nil && err != mempool.ErrNotExistInSql {
		logx.Errorf("[GetMempoolTxsListByAccountIndexAndTxTypeArray] err:%v", err)
		return nil, err
	}
	for _, tx := range mempoolTxs {
		var details []*globalRPCProto.TxDetailInfo
		for _, d := range tx.MempoolDetails {
			details = append(details, &globalRPCProto.TxDetailInfo{
				AssetId:      uint32(d.AssetId),
				AssetType:    uint32(d.AssetType),
				AccountIndex: uint32(d.AccountIndex),
				AccountName:  d.AccountName,
				BalanceDelta: d.BalanceDelta,
			})
		}
		resTx := &globalRPCProto.TxInfo{
			TxHash:        tx.TxHash,
			TxType:        uint32(tx.TxType),
			GasFeeAssetId: uint32(tx.GasFeeAssetId),
			GasFee:        tx.GasFee,
			NftIndex:      uint32(tx.NftIndex),
			PairIndex:     uint32(tx.PairIndex),
			AssetId:       uint32(tx.AssetId),
			TxAmount:      tx.TxAmount,
			NativeAddress: tx.NativeAddress,
			TxDetails:     details,
			Memo:          tx.Memo,
			AccountIndex:  uint32(tx.AccountIndex),
			Nonce:         uint64(tx.Nonce),
			CreateAt:      uint64(tx.CreatedAt.Unix()),
			Status:        uint32(tx.Status),
			BlockHeight:   uint64(tx.L2BlockHeight),
		}
		resp.TxsList = append(resp.TxsList, resTx)
	}
	return resp, nil
}
