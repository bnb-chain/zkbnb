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
	"github.com/stretchr/testify/assert"
	"github.com/zeromicro/go-zero/core/logx"
	"testing"
)

// GetLpValue

func TestGetLpValue(t *testing.T) {
	ctx := svc.NewServiceContext(ConfigProvider())
	srv := server.NewGlobalRPCServer(ctx)
	tmpCtx := new(context.Context)
	res, err := srv.GetLpValue(*tmpCtx, &globalRPCProto.ReqGetLpValue{
		PairIndex: uint32(0),
		LPAmount:  uint64(500),
	})
	assert.Nil(t, err)
	logx.Info(res)
}

func TestGetPairRatio(t *testing.T) {
	ctx := svc.NewServiceContext(ConfigProvider())
	srv := server.NewGlobalRPCServer(ctx)
	tmpCtx := new(context.Context)
	res, err := srv.GetPairRatio(*tmpCtx, &globalRPCProto.ReqGetPairRatio{
		PairIndex: uint32(0),
	})
	assert.Nil(t, err)
	logx.Info(res)
}
