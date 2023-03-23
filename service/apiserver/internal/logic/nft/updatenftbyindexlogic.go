package nft

import (
	"context"
	"github.com/bnb-chain/zkbnb/dao/nft"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbnb/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateNftByIndexLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateNftByIndexLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateNftByIndexLogic {
	return &UpdateNftByIndexLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateNftByIndexLogic) UpdateNftByIndex(req *types.ReqUpdateNft) (resp *types.History, err error) {
	tx, err := types2.ParseUpdateNftTxInfo(req.TxInfo)
	if err != nil {
		return nil, err
	}
	publicAddress := tx.GetL1AddressBySignature()
	accountInfo, err := l.svcCtx.StateFetcher.GetLatestAccount(tx.AccountIndex)
	if err != nil {
		return nil, err
	}
	originAddress := common.HexToAddress(accountInfo.L1Address)
	//Compare the original address and the public address to verify the identifier
	if publicAddress != originAddress {
		return nil, types2.DbErrFailToL1Signature
	}
	l2Nft, err := l.svcCtx.NftModel.GetNft(tx.NftIndex)
	if err != nil {
		if err == types2.DbErrNotFound {
			return nil, types2.AppErrNftNotFound
		}
		return nil, types2.AppErrInternal
	}
	if tx.AccountIndex != l2Nft.OwnerAccountIndex {
		return nil, types2.AppErrNotNftOwner
	}
	l2NftMetadataHistory, err := l.svcCtx.NftMetadataHistoryModel.GetL2NftMetadataHistory(tx.NftIndex)
	if err != nil {
		if err == types2.DbErrNotFound {
			return nil, types2.AppErrNftNotFound
		}
		return nil, types2.AppErrInternal
	}
	if l2NftMetadataHistory.Nonce+1 != tx.Nonce {
		return nil, types2.AppErrInvalidNftNonce
	}
	if l2NftMetadataHistory.Status != nft.Confirmed {
		return nil, types2.AppErrInvalidNft
	}
	if len(tx.MutableAttributes) > 2000 {
		return nil, types2.AppErrInvalidMutableAttributes.RefineError(2000)
	}
	l2NftMetadataHistory.Nonce = tx.Nonce
	l2NftMetadataHistory.Mutable = tx.MutableAttributes
	l2NftMetadataHistory.Status = nft.NotConfirmed
	l2NftMetadataHistory.IpnsCid = ""
	err = l.svcCtx.NftMetadataHistoryModel.UpdateL2NftMetadataHistoryInTransact(l2NftMetadataHistory)
	if err != nil {
		return nil, err
	}
	return &types.History{
		IpnsId: l2NftMetadataHistory.IpnsId,
	}, nil
}
