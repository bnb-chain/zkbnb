package transaction

import (
	"context"

	"github.com/bnb-chain/zkbas/service/apiserver/internal/logic/utils"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/types"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/errorcode"
)

type GetTxLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTxLogic {
	return &GetTxLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTxLogic) GetTx(req *types.ReqGetTx) (resp *types.EnrichedTx, err error) {
	resp = &types.EnrichedTx{}
	tx, err := l.svcCtx.MemCache.GetTxByHashWithFallback(req.Hash, func() (interface{}, error) {
		return l.svcCtx.TxModel.GetTxByHash(req.Hash)
	})
	if err == nil {
		resp.Tx = *utils.DbTx2Tx(tx)
		resp.Tx.AccountName, _ = l.svcCtx.MemCache.GetAccountNameByIndex(tx.AccountIndex)
		resp.Tx.AssetName, _ = l.svcCtx.MemCache.GetAssetNameById(tx.AssetId)
		block, err := l.svcCtx.MemCache.GetBlockByHeightWithFallback(tx.BlockHeight, func() (interface{}, error) {
			return l.svcCtx.BlockModel.GetBlockByHeight(resp.Tx.BlockHeight)
		})
		if err == nil {
			resp.CommittedAt = block.CommittedAt
			resp.ExecutedAt = block.CreatedAt.Unix()
			resp.VerifiedAt = block.VerifiedAt
		}
	} else {
		if err != errorcode.DbErrNotFound {
			return nil, errorcode.AppErrInternal
		}
		memppolTx, err := l.svcCtx.MempoolModel.GetMempoolTxByTxHash(req.Hash)
		if err != nil {
			if err == errorcode.DbErrNotFound {
				return nil, errorcode.AppErrNotFound
			}
			return nil, errorcode.AppErrInternal
		}
		resp.Tx = *utils.DbMempoolTx2Tx(memppolTx)
		resp.Tx.AccountName, _ = l.svcCtx.MemCache.GetAccountNameByIndex(tx.AccountIndex)
		resp.Tx.AssetName, _ = l.svcCtx.MemCache.GetAssetNameById(tx.AssetId)
	}

	return resp, nil
}
