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

package txVerification

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/common/commonTx"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"github.com/zecrey-labs/zecrey-legend/common/util"
	"math/big"
)

type (
	TransferTxInfo        = commonTx.TransferTxInfo
	SwapTxInfo            = commonTx.SwapTxInfo
	AddLiquidityTxInfo    = commonTx.AddLiquidityTxInfo
	RemoveLiquidityTxInfo = commonTx.RemoveLiquidityTxInfo
	WithdrawTxInfo        = commonTx.WithdrawTxInfo
	MintNftTxInfo         = commonTx.MintNftTxInfo
	TransferNftTxInfo     = commonTx.TransferNftTxInfo
	SetNftPriceTxInfo     = commonTx.SetNftPriceTxInfo
	BuyNftTxInfo          = commonTx.BuyNftTxInfo
	WithdrawNftTxInfo     = commonTx.WithdrawNftTxInfo

	PublicKey = eddsa.PublicKey

	MempoolTxDetail = mempool.MempoolTxDetail

	PoolInfo = util.PoolInfo
	NftInfo  = util.NftInfo
)

const (
	Base = 10

	TenThousand = 10000

	GeneralAssetType     = commonAsset.GeneralAssetType
	LiquidityAssetType   = commonAsset.LiquidityAssetType
	LiquidityLpAssetType = commonAsset.LiquidityLpAssetType
	NftAssetType         = commonAsset.NftAssetType
)

var (
	ZeroBigInt = big.NewInt(0)
)
