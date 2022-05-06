/*
 * Copyright © 2021 Zecrey Protocol
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/core/logx"
)

func TestGetLatestAccountAssetInfo(t *testing.T) {
	ctx := svc.NewServiceContext(ConfigProvider())
	srv := server.NewGlobalRPCServer(ctx)
	tmpCtx := new(context.Context)
	res, err := srv.GetLatestAccountAssetInfo(*tmpCtx, &globalRPCProto.ReqGetLatestAccountAssetInfo{
		AccountIndex: uint64(3),
		AssetId:      uint64(1),
	})
	assert.Nil(t, err)
	logx.Info(res)
}

func TestGetLatestAccountInfoByAccountIndex(t *testing.T) {
	ctx := svc.NewServiceContext(ConfigProvider())
	srv := server.NewGlobalRPCServer(ctx)
	tmpCtx := new(context.Context)
	res, err := srv.GetLatestAccountInfoByAccountIndex(*tmpCtx, &globalRPCProto.ReqGetLatestAccountInfoByAccountIndex{
		AccountIndex: uint64(3),
	})
	assert.Nil(t, err)
	logx.Info(res)
}

func TestGetLatestAccountLockAsset(t *testing.T) {
	ctx := svc.NewServiceContext(ConfigProvider())
	srv := server.NewGlobalRPCServer(ctx)
	tmpCtx := new(context.Context)
	res, err := srv.GetLatestAccountLockAsset(*tmpCtx, &globalRPCProto.ReqGetLatestAccountLockAsset{
		AccountIndex: uint64(3),
		AssetId:      uint64(0),
		ChainId:      uint64(0),
	})
	assert.Nil(t, err)
	logx.Info(res)
}

func TestGetLatestAccountLp(t *testing.T) {
	ctx := svc.NewServiceContext(ConfigProvider())
	srv := server.NewGlobalRPCServer(ctx)
	tmpCtx := new(context.Context)
	res, err := srv.GetLatestAccountLp(*tmpCtx, &globalRPCProto.ReqGetLatestAccountLp{
		AccountIndex: uint64(0),
		PairIndex:    uint64(0),
	})
	assert.Nil(t, err)
	logx.Info(res)
}

func TestGetLatestPoolInfo(t *testing.T) {
	ctx := svc.NewServiceContext(ConfigProvider())
	srv := server.NewGlobalRPCServer(ctx)
	tmpCtx := new(context.Context)
	res, err := srv.GetLatestPoolInfo(*tmpCtx, &globalRPCProto.ReqGetLatestPoolInfo{
		PairIndex: uint64(0),
	})
	assert.Nil(t, err)
	logx.Info(res)
}
