package transaction

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/utils"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetmempoolTxsByAccountNameLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetmempoolTxsByAccountNameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetmempoolTxsByAccountNameLogic {
	return &GetmempoolTxsByAccountNameLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetmempoolTxsByAccountNameLogic) GetmempoolTxsByAccountName(req *types.ReqGetmempoolTxsByAccountName) (*types.RespGetmempoolTxsByAccountName, error) {
	if !utils.ValidateAccountName(req.AccountName) {
		return nil, errorcode.AppErrInvalidParam.RefineError("invalid AccountName")
	}

	accountIndex, err := l.svcCtx.MemCache.GetAccountIndexByName(req.AccountName)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}

	mempoolTxs, err := l.svcCtx.MempoolModel.GetPendingMempoolTxsByAccountIndex(accountIndex)
	if err != nil {
		if err != errorcode.DbErrNotFound {
			return nil, errorcode.AppErrInternal
		}
	}

	resp := &types.RespGetmempoolTxsByAccountName{
		Total:      uint32(len(mempoolTxs)),
		MempoolTxs: make([]*types.Tx, 0),
	}
	for _, t := range mempoolTxs {
		tx := utils.DbMempoolTx2Tx(t)
		tx.AccountName, _ = l.svcCtx.MemCache.GetAccountNameByIndex(tx.AccountIndex)
		resp.MempoolTxs = append(resp.MempoolTxs, tx)
	}
	return resp, nil
}
