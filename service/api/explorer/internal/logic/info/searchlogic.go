package info

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/repo/account"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/repo/block"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/repo/sysconf"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/repo/tx"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/explorer/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext

	sysconfigModel sysconf.Sysconf
	block          block.Block
	tx             tx.Tx
	account        account.AccountModel
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
	// check if it is searching by blockHeight
	blockHeight, err := strconv.ParseInt(req.Info, 10, 64)
	if err == nil {
		_, err = l.block.GetBlockByBlockHeight(blockHeight)
		resp.DataType = util.TypeBlockHeight
		if err != nil {
			err1 := fmt.Errorf("[explorer.info.SearchInfo] find block by height %d error: %s", blockHeight, err.Error())
			l.Error(err1)
			return nil, err
		}
		return resp, nil
	}
	// check if this is for querying tx by hash
	_, err = l.tx.GetTxByTxHash(req.Info)
	if err == nil {
		resp.DataType = util.TypeTxType
		return resp, nil
	}
	// check if this is for querying account by name
	_, err = l.account.GetAccountByAccountName(l.ctx, req.Info)
	resp.DataType = util.TypeAccountName
	if err != nil {
		err = fmt.Errorf("[explorer.info.SearchInfo] find block by name %s error: %s", req.Info, err.Error())
		l.Error(err)
		return nil, err
	}
	return resp, nil
}
