package nftoperator

import (
	table "github.com/zecrey-labs/zecrey-legend/common/model/nft"
	"github.com/zecrey-labs/zecrey-legend/service/cronjob/monitor/internal/svc"
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
