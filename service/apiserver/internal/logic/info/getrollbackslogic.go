package info

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbnb/types"
)

type GetRollbacksLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetRollbacksLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetRollbacksLogic {
	return &GetRollbacksLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetRollbacksLogic) GetRollbacks(req *types.ReqGetRollbacks) (resp *types.Rollbacks, err error) {
	resp = &types.Rollbacks{
		Rollbacks: make([]*types.Rollback, 0, req.Limit),
		Total:     uint32(0),
	}
	count, err := l.svcCtx.RollbackModel.GetCount(req.FromBlockHeight, int(req.Limit), int64(req.Offset))
	if err != nil && err != types2.DbErrNotFound {
		return nil, types2.AppErrInternal
	}
	if count == 0 {
		return resp, nil
	}
	resp.Total = uint32(count)
	rollbacks, err := l.svcCtx.RollbackModel.Get(req.FromBlockHeight, int(req.Limit), int64(req.Offset))
	if err != nil && err != types2.DbErrNotFound {
		return nil, types2.AppErrInternal
	}
	if rollbacks != nil {
		for _, a := range rollbacks {
			resp.Rollbacks = append(resp.Rollbacks, &types.Rollback{
				FromBlockHeight: a.FromBlockHeight,
				FromTxHash:      a.FromTxHash,
				CreatedAt:       a.CreatedAt.Unix(),
				ID:              a.ID,
			})
		}
	}
	return resp, nil
}
