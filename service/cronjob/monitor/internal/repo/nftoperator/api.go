package nftoperator

import (
	table "github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/service/cronjob/monitor/internal/svc"
)

type Model interface {
	CreateNfts(pendingNewNfts []*table.L2Nft) (err error)
}

func New(svcCtx *svc.ServiceContext) Model {
	return &model{
		table: "l2_nft",
		db:    svcCtx.GormPointer,
		cache: svcCtx.Cache,
	}
}
