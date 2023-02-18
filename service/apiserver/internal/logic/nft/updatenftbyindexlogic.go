package nft

import (
	"context"
	"github.com/bnb-chain/zkbnb/dao/nft"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/signature"
	types2 "github.com/bnb-chain/zkbnb/types"
	"gorm.io/gorm"
	"strconv"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"

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
	err = l.verifySignature.VerifySignatureInfo(types2.TxTypeEmpty, strconv.FormatInt(req.AccountIndex, 10), req.TxSignature)
	if err != nil {
		return nil, err
	}
	l2Nft, err := l.svcCtx.NftModel.GetNft(req.NftIndex)
	if err != nil {
		if err == types2.DbErrNotFound {
			return nil, types2.AppErrNftNotFound
		}
		return nil, types2.AppErrInternal
	}
	if req.AccountIndex != l2Nft.OwnerAccountIndex {
		return nil, types2.AppErrNotNftOwner
	}
	if l2Nft.IpfsStatus == nft.NotConfirmed {
		return nil, types2.AppErrInvalidNft
	}
	if len(req.MutableAttributes) > 2000 {
		return nil, types2.AppErrInvalidMutableAttributes.RefineError(2000)
	}
	history := &nft.L2NftMetadataHistory{
		NftIndex: req.NftIndex,
		IpnsName: l2Nft.IpnsName,
		IpnsId:   l2Nft.IpnsId,
		Mutable:  req.MutableAttributes,
		Status:   nft.NotConfirmed,
	}
	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		err = l.svcCtx.NftMetadataHistoryModel.DeleteL2NftMetadataHistoryInTransact(tx, req.NftIndex)
		if err != nil {
			return err
		}
		err = l.svcCtx.NftMetadataHistoryModel.CreateL2NftMetadataHistoryInTransact(tx, history)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &types.History{
		IpnsId: l2Nft.IpnsId,
	}, nil
}
