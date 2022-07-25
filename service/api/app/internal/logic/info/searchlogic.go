package info

import (
	"context"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/account"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/block"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/sysconf"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/tx"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"
)

type SearchLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext

	sysconfigModel sysconf.Sysconf
	block          block.Block
	tx             tx.Model
	account        account.Model
}

func NewSearchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchLogic {
	return &SearchLogic{
		Logger:         logx.WithContext(ctx),
		ctx:            ctx,
		svcCtx:         svcCtx,
		sysconfigModel: sysconf.New(svcCtx),
		block:          block.New(svcCtx),
		tx:             tx.New(svcCtx),
		account:        account.New(svcCtx),
	}
}

func (l *SearchLogic) Search(req *types.ReqSearch) (*types.RespSearch, error) {
	resp := &types.RespSearch{}
	blockHeight, err := strconv.ParseInt(req.Info, 10, 64)
	if err == nil {
		if _, err = l.block.GetBlockByBlockHeight(l.ctx, blockHeight); err != nil {
			logx.Errorf("[GetBlockByBlockHeight] err:%v", err)
			return nil, err
		}
		resp.DataType = util.TypeBlockHeight
		return resp, nil
	}
	// TODO: prevent sql slow query, bloom Filter
	if _, err = l.tx.GetTxByTxHash(l.ctx, req.Info); err == nil {
		resp.DataType = util.TypeTxType
		return resp, nil
	}
	if _, err = l.account.GetAccountByAccountName(l.ctx, req.Info); err != nil {
		logx.Errorf("[GetAccountByAccountName] err:%v", err)
		return nil, err
	}
	resp.DataType = util.TypeAccountName
	return resp, nil
}
