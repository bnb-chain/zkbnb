package info

import (
	"context"
	"strconv"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbnb/types"
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
		if _, err = l.svcCtx.BlockModel.GetBlockByHeight(blockHeight); err != nil {
			if err == types2.DbErrNotFound {
				return nil, types2.AppErrBlockNotFound
			}
			return nil, types2.AppErrInternal
		}
		resp.DataType = types2.TypeBlockHeight
		return resp, nil
	}

	if strings.Contains(req.Keyword, ".") {
		if _, err = l.svcCtx.MemCache.GetAccountIndexByL1Address(req.Keyword); err != nil {
			if err == types2.DbErrNotFound {
				return nil, types2.AppErrAccountNotFound
			}
			return nil, types2.AppErrInternal
		}
		resp.DataType = types2.TypeAccountName
		return resp, nil
	}

	if _, err = l.svcCtx.TxModel.GetTxByHash(req.Keyword); err == nil {
		resp.DataType = types2.TypeTxType
		return resp, nil
	}

	return resp, types2.AppErrNotFound
}
