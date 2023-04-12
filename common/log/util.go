package log

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	AccountIdContext   = "AccountId"
	NftIndexContext    = "AccountId"
	AssetIdContext     = "AssetId"
	BlockHeightContext = "BlockHeight"
)

func UpdateCtxWithKV(ctx context.Context, keyValues ...interface{}) context.Context {
	if len(keyValues)%2 != 0 {
		logx.Error("UpdateCtxWithKV: odd number of arguments, context won't be updated")
		return ctx
	}
	for i := 0; i < len(keyValues); i += 2 {
		ctx = logx.ContextWithFields(ctx, logx.Field(keyValues[i].(string), keyValues[i+1]))
	}
	return ctx
}

func NewCtxWithKV(keyValues ...interface{}) context.Context {
	ctx := context.Background()
	if len(keyValues)%2 != 0 {
		logx.Error("NewCtxWithKV: odd number of arguments, context won't be updated")
		return ctx
	}
	for i := 0; i < len(keyValues); i += 2 {
		ctx = logx.ContextWithFields(ctx, logx.Field(keyValues[i].(string), keyValues[i+1]))
	}
	return ctx
}
