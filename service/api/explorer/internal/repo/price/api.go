package price

import (
	"context"

	"github.com/bnb-chain/zkbas/service/api/explorer/internal/svc"
)

type Price interface {
	GetCurrencyPrice(ctx context.Context, l2Symbol string) (price float64, err error)
}

func New(svcCtx *svc.ServiceContext) Price {
	return &price{
		cache: svcCtx.Cache,
	}
}
