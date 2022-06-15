package logic

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/globalRPCProto"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/repo/mempool"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLatestTxsListByAccountIndexLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	mempool mempool.Mempool
}

func NewGetLatestTxsListByAccountIndexLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLatestTxsListByAccountIndexLogic {
	return &GetLatestTxsListByAccountIndexLogic{
		ctx:     ctx,
		svcCtx:  svcCtx,
		Logger:  logx.WithContext(ctx),
		mempool: mempool.New(svcCtx.Config),
	}
}

//  Transaction
func (l *GetLatestTxsListByAccountIndexLogic) GetLatestTxsListByAccountIndex(in *globalRPCProto.ReqGetLatestTxsListByAccountIndex) (*globalRPCProto.RespGetLatestTxsListByAccountIndex, error) {
	count, err := l.mempool.GetMempoolTxsTotalCount()
	if err != nil {
		logx.Errorf("[GetMempoolTxsTotalCount] err:%v", err)
		return nil, err
	}
	resp := &globalRPCProto.RespGetLatestTxsListByAccountIndex{
		Total:   uint32(count),
		TxsList: make([]*globalRPCProto.TxInfo, 0),
	}
	mempoolTxs, err := l.mempool.GetMempoolTxsListByAccountIndex(int64(in.AccountIndex), int64(in.Limit), int64(in.Offset))
	if err != nil {
		logx.Errorf("[GetLatestLiquidityInfoForRead] err:%v", err)
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
