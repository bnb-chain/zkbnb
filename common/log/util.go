package log

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	AccountIndexCtx     = "accountId"
	NftIndexCtx         = "nftIndex"
	AssetIdCtx          = "assetId"
	BlockHeightContext  = "blockHeight"
	PoolTxIdContext     = "poolTxId"
	PoolTxIdListContext = "poolTxIds"
)

func UpdateCtxWithKV(ctx context.Context, keyValues ...interface{}) context.Context {
	if len(keyValues)%2 != 0 {
		return ctx
	}
	for i := 0; i < len(keyValues); i += 2 {
		key, ok := keyValues[i].(string)
		if ok {
			ctx = logx.ContextWithFields(ctx, logx.Field(key, keyValues[i+1]))
		}
	}
	return ctx
}

func NewCtxWithKV(keyValues ...interface{}) context.Context {
	ctx := context.Background()
	if len(keyValues)%2 != 0 {
		return ctx
	}
	for i := 0; i < len(keyValues); i += 2 {
		key, ok := keyValues[i].(string)
		if ok {
			ctx = logx.ContextWithFields(ctx, logx.Field(key, keyValues[i+1]))
		}
	}
	return ctx
}
