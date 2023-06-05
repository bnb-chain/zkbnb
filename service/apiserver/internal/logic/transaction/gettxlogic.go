package transaction

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/logic/utils"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbnb/types"
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
		resp.Tx = *utils.ConvertTx(tx, l.svcCtx.MemCache)
		block, err := l.svcCtx.MemCache.GetBlockByHeightWithFallback(tx.BlockHeight, func() (interface{}, error) {
			return l.svcCtx.BlockModel.GetBlockByHeight(resp.Tx.BlockHeight)
		})
		if err == nil {
			resp.CommittedAt = block.CommittedAt
			resp.ExecutedAt = block.CreatedAt.Unix()
			resp.VerifiedAt = block.VerifiedAt
		}
	} else {
		if err != types2.DbErrNotFound {
			return nil, types2.AppErrInternal
		}
		poolTx, err := l.svcCtx.TxPoolModel.GetTxByTxHash(req.Hash)
		if err != nil {
			if err == types2.DbErrNotFound {
				return nil, types2.AppErrPoolTxNotFound
			}
			return nil, types2.AppErrInternal
		}
		resp.Tx = *utils.ConvertTx(poolTx, l.svcCtx.MemCache)
	}

	return resp, nil
}
