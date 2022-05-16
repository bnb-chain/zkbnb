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
 *
 */

package proverUtil

import (
	"github.com/zecrey-labs/zecrey-crypto/zecrey-legend/circuit/bn254/block"
	"github.com/zecrey-labs/zecrey-crypto/zecrey-legend/circuit/bn254/std"
	"github.com/zecrey-labs/zecrey-legend/common/model/tx"
	"github.com/zecrey-labs/zecrey-legend/common/tree"
)

type (
	Tx   = tx.Tx
	Tree = tree.Tree

	CryptoTx = block.Tx

	CryptoRegisterZnsTx     = std.RegisterZnsTx
	CryptoCreatePairTx      = std.CreatePairTx
	CryptoDepositTx         = std.DepositTx
	CryptoDepositNftTx      = std.DepositNftTx
	CryptoTransferTx        = std.TransferTx
	CryptoSwapTx            = std.SwapTx
	CryptoAddLiquidityTx    = std.AddLiquidityTx
	CryptoRemoveLiquidityTx = std.RemoveLiquidityTx
	CryptoWithdrawTx        = std.WithdrawTx
	CryptoMintNftTx         = std.MintNftTx
	CryptoTransferNftTx     = std.TransferNftTx
	CryptoSetNftPriceTx     = std.SetNftPriceTx
	CryptoBuyNftTx          = std.BuyNftTx
	CryptoWithdrawNftTx     = std.WithdrawNftTx
	CryptoFullExitTx        = std.FullExitTx
	CryptoFullExitNftTx     = std.FullExitNftTx
)
