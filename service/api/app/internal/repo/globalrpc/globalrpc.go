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
	"sort"

	"github.com/zecrey-labs/zecrey-legend/common/model/account"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/globalRPCProto"
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
	cache               multcache.MultCache
}

func (m *globalRPC) GetSwapAmount(ctx context.Context, pairIndex, assetId uint64, assetAmount string, isFrom bool) (string, uint32, error) {
	resRpc, err := m.globalRPC.GetSwapAmount(ctx, &globalrpc.ReqGetSwapAmount{
		PairIndex:   uint32(pairIndex),
		AssetId:     uint32(assetId),
		AssetAmount: assetAmount,
		IsFrom:      isFrom,
	})
	if err != nil {
		return "", 0, err
	}
	return resRpc.SwapAssetAmount, resRpc.SwapAssetId, err
}

func (m *globalRPC) GetLpValue(ctx context.Context, pairIndex uint32, lpAmount string) (*globalRPCProto.RespGetLpValue, error) {
	return m.globalRPC.GetLpValue(ctx, &globalrpc.ReqGetLpValue{
		PairIndex: pairIndex,
		LPAmount:  lpAmount,
	})
}

func (m *globalRPC) GetPairInfo(ctx context.Context, pairIndex uint32) (*globalRPCProto.RespGetLatestPairInfo, error) {
	return m.globalRPC.GetLatestPairInfo(ctx, &globalrpc.ReqGetLatestPairInfo{
		PairIndex: pairIndex,
	})
}

func (m *globalRPC) GetNextNonce(ctx context.Context, accountIndex uint32) (uint64, error) {
	rpcRsp, err := m.globalRPC.GetNextNonce(ctx, &globalrpc.ReqGetNextNonce{
		AccountIndex: accountIndex,
	})
	return rpcRsp.GetNonce(), err
}

func (m *globalRPC) GetLatestAssetsListByAccountIndex(ctx context.Context, accountIndex uint32) ([]*globalrpc.AssetResult, error) {
	res, err := m.globalRPC.GetLatestAssetsListByAccountIndex(ctx, &globalrpc.ReqGetLatestAssetsListByAccountIndex{
		AccountIndex: accountIndex})
	return res.ResultAssetsList, err
}

func (m *globalRPC) GetLatestAccountInfoByAccountIndex(ctx context.Context, accountIndex int64) (*globalrpc.RespGetLatestAccountInfoByAccountIndex, error) {
	f := func() (interface{}, error) {
		res, err := m.globalRPC.GetLatestAccountInfoByAccountIndex(ctx, &globalrpc.ReqGetLatestAccountInfoByAccountIndex{
			AccountIndex: uint32(accountIndex),
		})
		if err != nil {
			return nil, err
		}
		sort.SliceStable(res.AccountAsset, func(i, j int) bool {
			return res.AccountAsset[i].AssetId < res.AccountAsset[j].AssetId
		})
		return res, nil
	}
	account := &globalRPCProto.RespGetLatestAccountInfoByAccountIndex{}
	value, err := m.cache.GetWithSet(ctx, multcache.SpliceCacheKeyAccountByAccountIndex(accountIndex), account, 5, f)
	if err != nil {
		return nil, err
	}
	account, _ = value.(*globalRPCProto.RespGetLatestAccountInfoByAccountIndex)
	return account, err
}

func (m *globalRPC) GetMaxOfferId(ctx context.Context, accountIndex uint32) (uint64, error) {
	rpcRsp, err := m.globalRPC.GetMaxOfferId(ctx, &globalrpc.ReqGetMaxOfferId{
		AccountIndex: accountIndex,
	})
	return rpcRsp.GetOfferId(), err
}

func (m *globalRPC) SendTx(ctx context.Context, txType uint32, txInfo string) (string, error) {
	rpcRsp, err := m.globalRPC.SendTx(ctx, &globalrpc.ReqSendTx{
		TxType: txType,
		TxInfo: txInfo,
	})
	return rpcRsp.GetTxId(), err
}

func (m *globalRPC) SendMintNftTx(ctx context.Context, txInfo string) (int64, error) {
	rpcRsp, err := m.globalRPC.SendMintNftTx(ctx, &globalrpc.ReqSendMintNftTx{
		TxInfo: txInfo,
	})
	return rpcRsp.GetNftIndex(), err
}

func (m *globalRPC) SendCreateCollectionTx(ctx context.Context, txInfo string) (int64, error) {
	rpcRsp, err := m.globalRPC.SendCreateCollectionTx(ctx, &globalrpc.ReqSendCreateCollectionTx{
		TxInfo: txInfo,
	})
	return rpcRsp.GetCollectionId(), err
}

func (m *globalRPC) SendAddLiquidityTx(ctx context.Context, txInfo string) (string, error) {
	rpcRsp, err := m.globalRPC.SendAddLiquidityTx(ctx, &globalrpc.ReqSendTxByRawInfo{
		TxInfo: txInfo,
	})
	return rpcRsp.GetTxId(), err
}

func (m *globalRPC) SendAtomicMatchTx(ctx context.Context, txInfo string) (string, error) {
	rpcRsp, err := m.globalRPC.SendAtomicMatchTx(ctx, &globalrpc.ReqSendTxByRawInfo{
		TxInfo: txInfo,
	})
	return rpcRsp.GetTxId(), err
}

func (m *globalRPC) SendCancelOfferTx(ctx context.Context, txInfo string) (string, error) {
	rpcRsp, err := m.globalRPC.SendCancelOfferTx(ctx, &globalrpc.ReqSendTxByRawInfo{
		TxInfo: txInfo,
	})
	return rpcRsp.GetTxId(), err
}

func (m *globalRPC) SendRemoveLiquidityTx(ctx context.Context, txInfo string) (string, error) {
	rpcRsp, err := m.globalRPC.SendRemoveLiquidityTx(ctx, &globalrpc.ReqSendTxByRawInfo{
		TxInfo: txInfo,
	})
	return rpcRsp.GetTxId(), err
}

func (m *globalRPC) SendSwapTx(ctx context.Context, txInfo string) (string, error) {
	rpcRsp, err := m.globalRPC.SendSwapTx(ctx, &globalrpc.ReqSendTxByRawInfo{
		TxInfo: txInfo,
	})
	return rpcRsp.GetTxId(), err
}

func (m *globalRPC) SendTransferNftTx(ctx context.Context, txInfo string) (string, error) {
	rpcRsp, err := m.globalRPC.SendTransferNftTx(ctx, &globalrpc.ReqSendTxByRawInfo{
		TxInfo: txInfo,
	})
	return rpcRsp.GetTxId(), err
}

func (m *globalRPC) SendTransferTx(ctx context.Context, txInfo string) (string, error) {
	rpcRsp, err := m.globalRPC.SendTransferTx(ctx, &globalrpc.ReqSendTxByRawInfo{
		TxInfo: txInfo,
	})
	return rpcRsp.GetTxId(), err
}

func (m *globalRPC) SendWithdrawNftTx(ctx context.Context, txInfo string) (string, error) {
	rpcRsp, err := m.globalRPC.SendWithdrawNftTx(ctx, &globalrpc.ReqSendTxByRawInfo{
		TxInfo: txInfo,
	})
	return rpcRsp.GetTxId(), err
}

func (m *globalRPC) SendWithdrawTx(ctx context.Context, txInfo string) (string, error) {
	rpcRsp, err := m.globalRPC.SendWithdrawTx(ctx, &globalrpc.ReqSendTxByRawInfo{
		TxInfo: txInfo,
	})
	return rpcRsp.GetTxId(), err
}
