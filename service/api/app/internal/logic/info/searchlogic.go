package info

import (
	"context"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type SearchLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSearchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchLogic {
	return &SearchLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SearchLogic) Search(req *types.ReqSearch) (*types.RespSearch, error) {
	resp := &types.RespSearch{}
	blockHeight, err := strconv.ParseInt(req.Info, 10, 64)
	if err == nil {
		if _, err = l.svcCtx.BlockModel.GetBlockByBlockHeight(blockHeight); err != nil {
			logx.Errorf("[GetBlockByBlockHeight] err: %s", err.Error())
			return nil, err
		}
		resp.DataType = util.TypeBlockHeight
		return resp, nil
	}
	// TODO: prevent sql slow query, bloom Filter
	if _, err = l.svcCtx.TxModel.GetTxByTxHash(req.Info); err == nil {
		resp.DataType = util.TypeTxType
		return resp, nil
	}
	if _, err = l.svcCtx.AccountModel.GetAccountByAccountName(req.Info); err != nil {
		logx.Errorf("[GetAccountByAccountName] err: %s", err.Error())
		return nil, err
	}
	resp.DataType = util.TypeAccountName
	return resp, nil
}
