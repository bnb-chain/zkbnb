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

func (l *GetRollbacksLogic) GetRollbacks(req *types.ReqGetRollbacks) (resp []*types.Rollbacks, err error) {
	resp = make([]*types.Rollbacks, 0)
	rollbacks, err := l.svcCtx.RollbackModel.Get(req.FromBlockHeight, int(req.Limit), int64(req.Offset))
	if err != nil && err != types2.DbErrNotFound {
		return nil, types2.AppErrInternal
	}
	if rollbacks != nil {
		for _, a := range rollbacks {
			resp = append(resp, &types.Rollbacks{
				FromBlockHeight: a.FromBlockHeight,
				FromTxHash:      a.FromTxHash,
			})
		}
	}
	return resp, nil
}
