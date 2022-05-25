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
	"errors"
	"fmt"
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

func packSendTxResp(
	status int64,
	msg string,
	err string,
	result *globalRPCProto.ResultSendTx,
) (res *globalRPCProto.RespSendTx) {
	res = &globalRPCProto.RespSendTx{
		Status: status,
		Msg:    msg,
		Err:    err,
		Result: result,
	}
	return res
}

func (l *SendTxLogic) SendTx(in *globalRPCProto.ReqSendTx) (resp *globalRPCProto.RespSendTx, err error) {
	var (
		txId       string
		resultResp *globalRPCProto.ResultSendTx
	)
	switch in.TxType {
	case commonTx.TxTypeTransfer:
		txId, err = l.sendTransferTx(in.TxInfo)

		resultResp = &globalRPCProto.ResultSendTx{
			TxId: txId,
		}

		if err != nil {
			errInfo := fmt.Sprintf("[sendtxlogic.sendTransferTx] %s", err.Error())
			logx.Error(errInfo)
			return packSendTxResp(FailStatus, FailMsg, errInfo, resultResp), err
		}
		break
	case commonTx.TxTypeSwap:

		txId, err = l.sendSwapTx(in.TxInfo)

		resultResp = &globalRPCProto.ResultSendTx{
			TxId: txId,
		}

		if err != nil {
			errInfo := fmt.Sprintf("[sendtxlogic.sendTransferTx] %s", err.Error())
			logx.Error(errInfo)
			return packSendTxResp(FailStatus, FailMsg, errInfo, resultResp), err
		}
		break
	case commonTx.TxTypeAddLiquidity:
		txId, err = l.sendAddLiquidityTx(in.TxInfo)

		resultResp = &globalRPCProto.ResultSendTx{
			TxId: txId,
		}

		if err != nil {
			errInfo := fmt.Sprintf("[sendtxlogic.sendTransferTx] %s", err.Error())
			logx.Error(errInfo)
			return packSendTxResp(FailStatus, FailMsg, errInfo, resultResp), err
		}
		break
	case commonTx.TxTypeRemoveLiquidity:
		txId, err = l.sendRemoveLiquidityTx(in.TxInfo)

		resultResp = &globalRPCProto.ResultSendTx{
			TxId: txId,
		}

		if err != nil {
			errInfo := fmt.Sprintf("[sendtxlogic.sendTransferTx] %s", err.Error())
			logx.Error(errInfo)
			return packSendTxResp(FailStatus, FailMsg, errInfo, resultResp), err
		}
		break
	case commonTx.TxTypeWithdraw:
		txId, err = l.sendWithdrawTx(in.TxInfo)

		resultResp = &globalRPCProto.ResultSendTx{
			TxId: txId,
		}

		if err != nil {
			errInfo := fmt.Sprintf("[sendtxlogic.sendTransferTx] %s", err.Error())
			logx.Error(errInfo)
			return packSendTxResp(FailStatus, FailMsg, errInfo, resultResp), err
		}
		break
	case commonTx.TxTypeCreateCollection:
		txId, err = l.sendCreateCollectionTx(in.TxInfo)

		resultResp = &globalRPCProto.ResultSendTx{
			TxId: txId,
		}

		if err != nil {
			errInfo := fmt.Sprintf("[sendtxlogic.sendTransferTx] %s", err.Error())
			logx.Error(errInfo)
			return packSendTxResp(FailStatus, FailMsg, errInfo, resultResp), err
		}
		break
	case commonTx.TxTypeMintNft:
		break
	case commonTx.TxTypeTransferNft:
		break
	case commonTx.TxTypeAtomicMatch:
		break
	case commonTx.TxTypeCancelOffer:
		break
	case commonTx.TxTypeWithdrawNft:
		break
	case commonTx.TxTypeOffer:
		break
	default:
		errInfo := "[sendtxlogic] invalid tx type"
		return packSendTxResp(FailStatus, FailMsg, errInfo, resultResp), errors.New(errInfo)
	}
	return packSendTxResp(SuccessStatus, SuccessMsg, "", resultResp), nil
}
