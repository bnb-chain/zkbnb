package info

import (
	"context"
	"fmt"
	"strconv"

	"github.com/zecrey-labs/zecrey-legend/common/util"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/explorer/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
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

func (l *SearchLogic) Search(req *types.ReqSearch) (resp *types.RespSearch, err error) {
	// check if it is searching by blockHeight
	blockHeight, e := strconv.ParseInt(req.Info, 10, 64)
	if e == nil {
		_, e = l.svcCtx.Block.GetBlockByBlockHeight(blockHeight)
		resp.DataType = util.TypeBlockHeight
		if e != nil {
			err = fmt.Errorf("[explorer.info.SearchInfo] find block by height %d error: %s", blockHeight, e.Error())
			l.Error(err)
		}
		return
	}
	// check if this is for querying tx by hash
	_, err = l.svcCtx.Tx.GetTxByTxHash(req.Info)
	if err == nil {
		resp.DataType = util.TypeTxType
		return
	}
	// check if this is for querying account by name
	_, err = l.svcCtx.Account.GetAccountByAccountName(req.Info)
	resp.DataType = util.TypeAccountName
	if err != nil {
		err = fmt.Errorf("[explorer.info.SearchInfo] find block by name %s error: %s", req.Info, err.Error())
		l.Error(err)
	}
	return
}
