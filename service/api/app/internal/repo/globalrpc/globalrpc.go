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
	ctx                 context.Context
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

func (m *globalRPC) GetLpValue(pairIndex uint32, lpAmount string) (*globalRPCProto.RespGetLpValue, error) {
	return m.globalRPC.GetLpValue(m.ctx, &globalrpc.ReqGetLpValue{
		PairIndex: pairIndex,
		LPAmount:  lpAmount,
	})
}

func (m *globalRPC) GetPairInfo(pairIndex uint32) (*globalRPCProto.RespGetLatestPairInfo, error) {
	return m.globalRPC.GetLatestPairInfo(m.ctx, &globalrpc.ReqGetLatestPairInfo{
		PairIndex: pairIndex,
	})
}

func (m *globalRPC) GetNextNonce(accountIndex uint32) (uint64, error) {
	rpcRsp, err := m.globalRPC.GetNextNonce(m.ctx, &globalrpc.ReqGetNextNonce{
		AccountIndex: accountIndex,
	})
	return rpcRsp.GetNonce(), err
}

func (m *globalRPC) GetLatestAssetsListByAccountIndex(accountIndex uint32) ([]*globalrpc.AssetResult, error) {
	res, err := m.globalRPC.GetLatestAssetsListByAccountIndex(m.ctx, &globalrpc.ReqGetLatestAssetsListByAccountIndex{
		AccountIndex: accountIndex,
	})
	return res.ResultAssetsList, err
}

func (m *globalRPC) GetLatestAccountInfoByAccountIndex(accountIndex uint32) (*globalrpc.RespGetLatestAccountInfoByAccountIndex, error) {
	res, err := m.globalRPC.GetLatestAccountInfoByAccountIndex(m.ctx, &globalrpc.ReqGetLatestAccountInfoByAccountIndex{
		AccountIndex: accountIndex,
	})
	sort.SliceStable(res.AccountAsset, func(i, j int) bool {
		return res.AccountAsset[i].AssetId < res.AccountAsset[j].AssetId
	})
	return res, err
}

func (m *globalRPC) GetMaxOfferId(accountIndex uint32) (uint64, error) {
	rpcRsp, err := m.globalRPC.GetMaxOfferId(m.ctx, &globalrpc.ReqGetMaxOfferId{
		AccountIndex: accountIndex,
	})
	return rpcRsp.GetOfferId(), err
}

func (m *globalRPC) SendTx(txType uint32, txInfo string) (string, error) {
	rpcRsp, err := m.globalRPC.SendTx(m.ctx, &globalrpc.ReqSendTx{
		TxType: txType,
		TxInfo: txInfo,
	})
	return rpcRsp.GetTxId(), err
}

func (m *globalRPC) SendMintNftTx(txInfo string) (int64, error) {
	rpcRsp, err := m.globalRPC.SendMintNftTx(m.ctx, &globalrpc.ReqSendMintNftTx{
		TxInfo: txInfo,
	})
	return rpcRsp.GetNftIndex(), err
}

func (m *globalRPC) SendCreateCollectionTx(txInfo string) (int64, error) {
	rpcRsp, err := m.globalRPC.SendCreateCollectionTx(m.ctx, &globalrpc.ReqSendCreateCollectionTx{
		TxInfo: txInfo,
	})
	return rpcRsp.GetCollectionId(), err
}

func (m *globalRPC) SendAddLiquidityTx(txInfo string) (string, error) {
	rpcRsp, err := m.globalRPC.SendAddLiquidityTx(m.ctx, &globalrpc.ReqSendTxByRawInfo{
		TxInfo: txInfo,
	})
	return rpcRsp.GetTxId(), err
}

func (m *globalRPC) SendAtomicMatchTx(txInfo string) (string, error) {
	rpcRsp, err := m.globalRPC.SendAtomicMatchTx(m.ctx, &globalrpc.ReqSendTxByRawInfo{
		TxInfo: txInfo,
	})
	return rpcRsp.GetTxId(), err
}

func (m *globalRPC) SendCancelOfferTx(txInfo string) (string, error) {
	rpcRsp, err := m.globalRPC.SendCancelOfferTx(m.ctx, &globalrpc.ReqSendTxByRawInfo{
		TxInfo: txInfo,
	})
	return rpcRsp.GetTxId(), err
}

func (m *globalRPC) SendRemoveLiquidityTx(txInfo string) (string, error) {
	rpcRsp, err := m.globalRPC.SendRemoveLiquidityTx(m.ctx, &globalrpc.ReqSendTxByRawInfo{
		TxInfo: txInfo,
	})
	return rpcRsp.GetTxId(), err
}

func (m *globalRPC) SendSwapTx(txInfo string) (string, error) {
	rpcRsp, err := m.globalRPC.SendSwapTx(m.ctx, &globalrpc.ReqSendTxByRawInfo{
		TxInfo: txInfo,
	})
	return rpcRsp.GetTxId(), err
}

func (m *globalRPC) SendTransferNftTx(txInfo string) (string, error) {
	rpcRsp, err := m.globalRPC.SendTransferNftTx(m.ctx, &globalrpc.ReqSendTxByRawInfo{
		TxInfo: txInfo,
	})
	return rpcRsp.GetTxId(), err
}

func (m *globalRPC) SendTransferTx(txInfo string) (string, error) {
	rpcRsp, err := m.globalRPC.SendTransferTx(m.ctx, &globalrpc.ReqSendTxByRawInfo{
		TxInfo: txInfo,
	})
	return rpcRsp.GetTxId(), err
}

func (m *globalRPC) SendWithdrawNftTx(txInfo string) (string, error) {
	rpcRsp, err := m.globalRPC.SendWithdrawNftTx(m.ctx, &globalrpc.ReqSendTxByRawInfo{
		TxInfo: txInfo,
	})
	return rpcRsp.GetTxId(), err
}

func (m *globalRPC) SendWithdrawTx(txInfo string) (string, error) {
	rpcRsp, err := m.globalRPC.SendWithdrawTx(m.ctx, &globalrpc.ReqSendTxByRawInfo{
		TxInfo: txInfo,
	})
	return rpcRsp.GetTxId(), err
}
