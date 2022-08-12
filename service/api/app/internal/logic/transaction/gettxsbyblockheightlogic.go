package transaction

import (
	"context"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/utils"

	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTxsByBlockHeightLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTxsByBlockHeightLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTxsByBlockHeightLogic {
	return &GetTxsByBlockHeightLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTxsByBlockHeightLogic) GetTxsByBlockHeight(req *types.ReqGetTxsByBlockHeight) (resp *types.RespGetTxsByBlockHeight, err error) {
	block, err := l.svcCtx.MemCache.GetBlockByHeightWithFallback(int64(req.BlockHeight), func() (interface{}, error) {
		return l.svcCtx.BlockModel.GetBlockByBlockHeight(int64(req.BlockHeight))
	})
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}

	resp = &types.RespGetTxsByBlockHeight{
		Total: uint32(len(block.Txs)),
		Txs:   make([]*types.Tx, 0),
	}
	for _, t := range block.Txs {
		tx := utils.DbTx2Tx(t)
		tx.AccountName, _ = l.svcCtx.MemCache.GetAccountNameByIndex(tx.AccountIndex)
		resp.Txs = append(resp.Txs, tx)
	}
	return resp, nil
}
