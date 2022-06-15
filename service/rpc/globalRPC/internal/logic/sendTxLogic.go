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

package logic

import (
	"context"

	"github.com/zecrey-labs/zecrey-legend/common/commonTx"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/globalRPCProto"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type SendTxLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendTxLogic {
	return &SendTxLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SendTxLogic) SendTx(in *globalRPCProto.ReqSendTx) (*globalRPCProto.RespSendTx, error) {
	resp := &globalRPCProto.RespSendTx{}
	var err error
	switch in.TxType {
	case commonTx.TxTypeTransfer:
		resp.TxId, err = l.sendTransferTx(in.TxInfo)
		if err != nil {
			logx.Error("[sendTransferTx] err:%v", err)
			return nil, err
		}
	case commonTx.TxTypeSwap:
		resp.TxId, err = l.sendSwapTx(in.TxInfo)
		if err != nil {
			logx.Error("[sendSwapTx] err:%v", err)
			return nil, err
		}
	case commonTx.TxTypeAddLiquidity:
		resp.TxId, err = l.sendAddLiquidityTx(in.TxInfo)
		if err != nil {
			logx.Error("[sendAddLiquidityTx] err:%v", err)
			return nil, err
		}
	case commonTx.TxTypeRemoveLiquidity:
		resp.TxId, err = l.sendRemoveLiquidityTx(in.TxInfo)
		if err != nil {
			logx.Error("[sendRemoveLiquidityTx] err:%v", err)
			return nil, err
		}
	case commonTx.TxTypeWithdraw:
		resp.TxId, err = l.sendWithdrawTx(in.TxInfo)
		if err != nil {
			logx.Error("[sendWithdrawTx] err:%v", err)
			return nil, err
		}
	case commonTx.TxTypeCreateCollection:
		resp.TxId, err = l.sendCreateCollectionTx(in.TxInfo)
		if err != nil {
			logx.Error("[sendWithdrawTx] err:%v", err)
			return nil, err
		}
	case commonTx.TxTypeMintNft:
		resp.TxId, err = l.sendMintNftTx(in.TxInfo)
		if err != nil {
			logx.Error("[sendWithdrawTx] err:%v", err)
			return nil, err
		}
	case commonTx.TxTypeTransferNft:
		resp.TxId, err = l.sendTransferNftTx(in.TxInfo)
		if err != nil {
			logx.Error("[sendWithdrawTx] err:%v", err)
			return nil, err
		}
	case commonTx.TxTypeAtomicMatch:
		resp.TxId, err = l.sendAtomicMatchTx(in.TxInfo)
		if err != nil {
			logx.Error("[sendWithdrawTx] err:%v", err)
			return nil, err
		}
	case commonTx.TxTypeCancelOffer:
		resp.TxId, err = l.sendCancelOfferTx(in.TxInfo)
		if err != nil {
			logx.Error("[sendWithdrawTx] err:%v", err)
			return nil, err
		}
	case commonTx.TxTypeWithdrawNft:
		resp.TxId, err = l.sendWithdrawNftTx(in.TxInfo)
		if err != nil {
			logx.Error("[sendWithdrawTx] err:%v", err)
			return nil, err
		}
	case commonTx.TxTypeOffer:
		break
	default:
		logx.Error("[sendtxlogic] invalid tx type")
		return nil, err
	}
	return resp, err
}
