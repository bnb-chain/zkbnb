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
	"errors"
	"fmt"
	"github.com/zecrey-labs/zecrey-legend/common/commonTx"
	"github.com/zeromicro/go-zero/core/logx"
)

func GetTxTypeArray(txType uint) ([]uint8, error) {
	switch txType {
	case L2TransferType:
		return []uint8{commonTx.TxTypeTransfer}, nil
	case LiquidityType:
		return []uint8{commonTx.TxTypeAddLiquidity, commonTx.TxTypeRemoveLiquidity}, nil
	case L2SwapType:
		return []uint8{commonTx.TxTypeSwap}, nil
	case WithdrawAssetsType:
		return []uint8{commonTx.TxTypeWithdraw}, nil
	default:
		errInfo := fmt.Sprintf("[GetTxTypeArray] txType error: %v", txType)
		logx.Error(errInfo)
		return []uint8{}, errors.New(errInfo)
	}
}
