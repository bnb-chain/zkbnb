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
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/globalRPCProto"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/globalrpc"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/gorm"
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
		m.MempoolModel, m.RedisConnection, accountIndex)
	if err != nil {
		return nil, err
	}
	return accountInfo, nil
}

func (m *globalRPC) GetSwapAmount(pairIndex, assetId uint64, assetAmount string, isFrom bool) (string, uint32, error) {
	resRpc, err := m.globalRPC.GetSwapAmount(m.ctx, &globalrpc.ReqGetSwapAmount{
		PairIndex:   uint32(pairIndex),
		AssetId:     uint32(assetId),
		AssetAmount: assetAmount,
		IsFrom:      isFrom,
	})
	return resRpc.SwapAssetAmount, resRpc.SwapAssetId, err
}

func (m *globalRPC) GetLatestAccountInfoByAccountIndex(accountIndex uint32) ([]*globalrpc.AssetResult, error) {
	res, err := m.globalRPC.GetLatestAssetsListByAccountIndex(m.ctx, &globalrpc.ReqGetLatestAssetsListByAccountIndex{
		AccountIndex: accountIndex,
	})
	return res.ResultAssetsList, err
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
func (m *globalRPC) GetLatestTxsListByAccountIndexAndTxType(accountIndex uint64, txType uint64, limit uint64, offset uint64) ([]*mempool.MempoolTx, error) {
	resRpc, _ := m.globalRPC.GetLatestTxsListByAccountIndexAndTxType(m.ctx, &globalrpc.ReqGetLatestTxsListByAccountIndexAndTxType{
		AccountIndex: uint32(accountIndex),
		TxType:       uint32(txType),
		Offset:       uint32(offset),
		Limit:        uint32(limit),
	})
	res := make([]*mempool.MempoolTx, 0)
	for _, each := range resRpc.GetTxsList() {
		singleTxDetail := make([]*mempool.MempoolTxDetail, 0)
		for _, eachDetail := range each.TxDetails {
			singleTxDetail = append(singleTxDetail, &mempool.MempoolTxDetail{
				AssetId:      int64(eachDetail.AssetId),
				AssetType:    int64(eachDetail.AssetType),
				AccountIndex: int64(eachDetail.AccountIndex),
				AccountName:  eachDetail.AccountName,
				BalanceDelta: eachDetail.BalanceDelta,
			})
		}

		res = append(res, &mempool.MempoolTx{
			Model:          gorm.Model{},
			TxHash:         each.TxHash,
			TxType:         int64(each.TxType),
			GasFeeAssetId:  int64(each.GasFeeAssetId),
			GasFee:         each.GasFee,
			AssetId:        int64(each.AssetId),
			TxAmount:       each.TxAmount,
			NativeAddress:  each.NativeAddress,
			MempoolDetails: singleTxDetail,
			TxInfo:         "",
			ExtraInfo:      "",
			Memo:           each.Memo,
			AccountIndex:   0,
			Nonce:          0,
			ExpiredAt:      0,
			L2BlockHeight:  int64(each.BlockHeight),
			Status:         int(each.Status),
		})
	}
	return res, nil
}

func (m *globalRPC) GetLatestTxsListByAccountIndex(accountIndex uint32, limit uint32) ([]*mempool.MempoolTx, uint32, error) {
	resRpc, err := m.globalRPC.GetLatestTxsListByAccountIndex(m.ctx, &globalrpc.ReqGetLatestTxsListByAccountIndex{
		AccountIndex: accountIndex,
		Limit:        limit,
	})

	txls := make([]*mempool.MempoolTx, 0)
	for _, each := range resRpc.GetTxsList() {
		txls = append(txls, &mempool.MempoolTx{
			Model:          gorm.Model{},
			TxHash:         each.TxHash,
			TxType:         int64(each.TxType),
			GasFeeAssetId:  int64(each.GasFeeAssetId),
			GasFee:         each.GasFee,
			NftIndex:       int64(each.NftIndex),
			PairIndex:      int64(each.PairIndex),
			AssetId:        int64(each.AssetId),
			TxAmount:       each.TxAmount,
			NativeAddress:  each.NativeAddress,
			MempoolDetails: nil,
			TxInfo:         "",
			ExtraInfo:      "",
			Memo:           each.Memo,
			AccountIndex:   int64(each.AccountIndex),
			Nonce:          int64(each.Nonce),
			ExpiredAt:      0,
			L2BlockHeight:  int64(each.BlockHeight),
			Status:         int(each.Status),
		})
	}

	return txls, resRpc.GetTotal(), err
}

func (m *globalRPC) SendTx(txType uint32, txInfo string) (string, error) {
	rpcRsp, err := m.globalRPC.SendTx(m.ctx, &globalrpc.ReqSendTx{
		TxType: txType,
		TxInfo: txInfo,
	})
	return rpcRsp.GetTxId(), err
}
