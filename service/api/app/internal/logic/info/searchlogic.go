package info

import (
	"context"
	"strconv"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/errorcode"
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

func (l *SearchLogic) Search(req *types.ReqSearch) (*types.Search, error) {
	resp := &types.Search{}
	blockHeight, err := strconv.ParseInt(req.Keyword, 10, 64)
	if err == nil {
		if _, err = l.svcCtx.BlockModel.GetBlockByBlockHeight(blockHeight); err != nil {
			if err == errorcode.DbErrNotFound {
				return nil, errorcode.AppErrNotFound
			}
			return nil, errorcode.AppErrInternal
		}
		resp.DataType = util.TypeBlockHeight
		return resp, nil
	}

	if strings.Contains(req.Keyword, ".") {
		if _, err = l.svcCtx.MemCache.GetAccountIndexByName(req.Keyword); err != nil {
			if err == errorcode.DbErrNotFound {
				return nil, errorcode.AppErrNotFound
			}
			return nil, errorcode.AppErrInternal
		}
		resp.DataType = util.TypeAccountName
		return resp, nil
	}

	if _, err = l.svcCtx.MemCache.GetAccountIndexByPk(req.Keyword); err == nil {
		resp.DataType = util.TypeAccountPk
		return resp, nil
	}

	if _, err = l.svcCtx.TxModel.GetTxByTxHash(req.Keyword); err == nil {
		resp.DataType = util.TypeTxType
		return resp, nil
	}

	return resp, errorcode.AppErrNotFound
}
