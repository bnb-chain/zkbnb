package nft

import (
	"context"
	"github.com/bnb-chain/zkbnb/dao/nft"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/signature"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbnb/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateNftByIndexLogic struct {
	logx.Logger
	ctx             context.Context
	svcCtx          *svc.ServiceContext
	verifySignature *signature.VerifySignature
}

func NewUpdateNftByIndexLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateNftByIndexLogic {
	verifySignature := signature.NewVerifySignature(ctx, svcCtx)
	return &UpdateNftByIndexLogic{
		Logger:          logx.WithContext(ctx),
		ctx:             ctx,
		svcCtx:          svcCtx,
		verifySignature: verifySignature,
	}
}

func (l *UpdateNftByIndexLogic) UpdateNftByIndex(req *types.ReqUpdateNft) (resp *types.History, err error) {
	err = l.verifySignature.VerifySignatureInfo(types2.TxTypeEmpty, req.TxInfo, req.TxSignature)
	if err != nil {
		return nil, err
	}
	tx, err := types2.ParseUpdateNftTxInfo(req.TxInfo)
	if err != nil {
		return nil, err
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
