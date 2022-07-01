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
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/logic/sendrawtx"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/repo/commglobalmap"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type SendTxLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	commglobalmap commglobalmap.Commglobalmap
}

func NewSendTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendTxLogic {
	return &SendTxLogic{
		ctx:           ctx,
		svcCtx:        svcCtx,
		Logger:        logx.WithContext(ctx),
		commglobalmap: commglobalmap.New(svcCtx),
	}
}

func (l *SendTxLogic) SendTx(in *globalRPCProto.ReqSendTx) (resp *globalRPCProto.RespSendTx, err error) {
	resp = &globalRPCProto.RespSendTx{}
	switch in.TxType {
	case commonTx.TxTypeTransfer:
		resp.TxId, err = sendrawtx.SendTransferTx(l.ctx, l.svcCtx, l.commglobalmap, in.TxInfo)
		if err != nil {
			logx.Error("[sendTransferTx] err:%v", err)
			return nil, err
		}
	case commonTx.TxTypeSwap:
		resp.TxId, err = sendrawtx.SendSwapTx(l.ctx, l.svcCtx, l.commglobalmap, in.TxInfo)
		if err != nil {
			logx.Error("[sendSwapTx] err:%v", err)
			return nil, err
		}
	case commonTx.TxTypeAddLiquidity:
		resp.TxId, err = sendrawtx.SendAddLiquidityTx(l.ctx, l.svcCtx, l.commglobalmap, in.TxInfo)
		if err != nil {
			logx.Error("[sendAddLiquidityTx] err:%v", err)
			return nil, err
		}
	case commonTx.TxTypeRemoveLiquidity:
		resp.TxId, err = sendrawtx.SendRemoveLiquidityTx(l.ctx, l.svcCtx, l.commglobalmap, in.TxInfo)
		if err != nil {
			logx.Error("[sendRemoveLiquidityTx] err:%v", err)
			return nil, err
		}
	case commonTx.TxTypeWithdraw:
		resp.TxId, err = sendrawtx.SendWithdrawTx(l.ctx, l.svcCtx, l.commglobalmap, in.TxInfo)
		if err != nil {
			logx.Error("[sendWithdrawTx] err:%v", err)
			return nil, err
		}
	case commonTx.TxTypeTransferNft:
		resp.TxId, err = sendrawtx.SendTransferNftTx(l.ctx, l.svcCtx, l.commglobalmap, in.TxInfo)
		if err != nil {
			logx.Error("[sendWithdrawTx] err:%v", err)
			return nil, err
		}
	case commonTx.TxTypeAtomicMatch:
		resp.TxId, err = sendrawtx.SendAtomicMatchTx(l.ctx, l.svcCtx, l.commglobalmap, in.TxInfo)
		if err != nil {
			logx.Error("[sendWithdrawTx] err:%v", err)
			return nil, err
		}
	case commonTx.TxTypeCancelOffer:
		resp.TxId, err = sendrawtx.SendCancelOfferTx(l.ctx, l.svcCtx, l.commglobalmap, in.TxInfo)
		if err != nil {
			logx.Error("[sendWithdrawTx] err:%v", err)
			return nil, err
		}
	case commonTx.TxTypeWithdrawNft:
		resp.TxId, err = sendrawtx.SendWithdrawNftTx(l.ctx, l.svcCtx, l.commglobalmap, in.TxInfo)
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
