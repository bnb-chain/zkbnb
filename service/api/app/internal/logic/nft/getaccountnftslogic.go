package nft

import (
	"context"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type GetAccountNftsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAccountNftsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAccountNftsLogic {
	return &GetAccountNftsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAccountNftsLogic) GetAccountNfts(req *types.ReqGetAccountNfts) (resp *types.Nfts, err error) {
	accountIndex := int64(0)
	switch req.By {
	case "account_index":
		accountIndex, err = strconv.ParseInt(req.Value, 10, 64)
		if err != nil {
			return nil, errorcode.AppErrInvalidParam.RefineError("invalid value for account_index")
		}
	case "account_name":
		accountIndex, err = l.svcCtx.MemCache.GetAccountIndexByName(req.Value)
	case "account_pk":
		accountIndex, err = l.svcCtx.MemCache.GetAccountIndexByPk(req.Value)
	default:
		return nil, errorcode.AppErrInvalidParam.RefineError("param by should be account_index|account_name|account_pk")
	}

	if err != nil {
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}

	total, err := l.svcCtx.NftModel.GetAccountNftTotalCount(accountIndex)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return nil, errorcode.AppErrNotFound
		}
		return nil, errorcode.AppErrInternal
	}

	resp = &types.Nfts{
		Total: total,
		Nfts:  make([]*types.Nft, 0),
	}
	if total == 0 || total <= int64(req.Offset) {
		return resp, nil
	}

	nftList, err := l.svcCtx.NftModel.GetNftListByAccountIndex(accountIndex, int64(req.Limit), int64(req.Offset))
	if err != nil {
		return nil, errorcode.AppErrInternal
	}

	for _, nftItem := range nftList {
		resp.Nfts = append(resp.Nfts, &types.Nft{
			Index:               nftItem.NftIndex,
			CreatorAccountIndex: nftItem.CreatorAccountIndex,
			OwnerAccountIndex:   nftItem.OwnerAccountIndex,
			ContentHash:         nftItem.NftContentHash,
			L1Address:           nftItem.NftL1Address,
			L1TokenId:           nftItem.NftL1TokenId,
			CreatorTreasuryRate: nftItem.CreatorTreasuryRate,
			CollectionId:        nftItem.CollectionId,
		})
	}
	return resp, nil
}
