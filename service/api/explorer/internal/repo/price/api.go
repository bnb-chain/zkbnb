package price

import (
	"context"
)

type Price interface {
	GetCurrencyPrice(ctx context.Context, l2Symbol string) (price float64, err error)
}
