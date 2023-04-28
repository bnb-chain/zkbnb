package nft

import (
	"context"
	"github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/dao/nft"
	"github.com/bnb-chain/zkbnb/dao/tx"
	types2 "github.com/bnb-chain/zkbnb/types"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetNftByTxHashLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetNftByTxHashLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetNftByTxHashLogic {
	return &GetNftByTxHashLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetNftByTxHashLogic) GetNftByTxHash(req *types.ReqGetNftIndex) (resp *types.NftIndex, err error) {
	poolTx, err := l.svcCtx.TxPoolModel.GetTxUnscopedByTxHash(req.TxHash)
	if err != nil {
		if err == types2.DbErrNotFound {
			return nil, types2.AppErrPoolTxNotFound
		}
		return nil, types2.AppErrInternal
	}
	if poolTx.TxType != types2.TxTypeMintNft {
		return nil, types2.AppErrNftNotFound
	}
	if poolTx.TxStatus == tx.StatusPending {
		return nil, types2.AppErrPoolTxRunning
	} else if poolTx.TxStatus == tx.StatusFailed {
		return nil, types2.AppErrPoolTxFailed
	}
	txInfo, err := types2.ParseMintNftTxInfo(poolTx.TxInfo)
	if err != nil {
		logx.Errorf("parse mint nft tx failed: %s", err.Error())
		return nil, types2.AppErrInvalidTxInfo
	}
	history, err := l.svcCtx.NftMetadataHistoryModel.GetL2NftMetadataHistoryByHash(req.TxHash)
	if err == nil {
		if history.NftIndex == types2.NilNftIndex {
			history.NftIndex = poolTx.NftIndex
			history.Status = nft.NotConfirmed
		}
		l.svcCtx.NftMetadataHistoryModel.UpdateL2NftMetadataHistoryInTransact(history)
	}
	return &types.NftIndex{
		Index:  poolTx.NftIndex,
		IpfsId: common.GenerateCid(txInfo.NftContentHash),
	}, nil
}
