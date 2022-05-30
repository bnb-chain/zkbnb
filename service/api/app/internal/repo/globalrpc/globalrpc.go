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

package globalrpc

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/common/model/account"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/common/util/globalmapHandler"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/globalrpc"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type globalRPC struct {
	AccountModel        account.AccountModel
	AccountHistoryModel account.AccountHistoryModel
	MempoolModel        mempool.MempoolModel
	MempoolDetailModel  mempool.MempoolTxDetailModel
	RedisConnection     *redis.Redis
	globalRPC           globalrpc.GlobalRPC
	ctx                 context.Context
}

func (m *globalRPC) GetLatestAccountInfo(accountIndex int64) (accountInfo *commonAsset.AccountInfo, err error) {
	accountInfo, err = globalmapHandler.GetLatestAccountInfo(m.AccountModel,
		m.MempoolModel, m.MempoolDetailModel, m.RedisConnection, accountIndex)
	if err != nil {
		return nil, err
	}
	return accountInfo, nil
}

func (m *globalRPC) GetSwapAmount(pairIndex, assetId uint16, assetAmount uint64, isFrom bool) (uint64, uint16, uint16) {
	resRpc, _ := m.globalRPC.GetSwapAmount(m.ctx, &globalrpc.ReqGetSwapAmount{
		PairIndex:   uint32(pairIndex),
		AssetId:     uint32(assetId),
		AssetAmount: assetAmount,
		IsFrom:      isFrom,
	})
	return resRpc.Result.ResAssetAmount, uint16(resRpc.Result.PairIndex), uint16(resRpc.Result.ResAssetId)
}
