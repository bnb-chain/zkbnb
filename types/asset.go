/*
 * Copyright © 2021 ZkBAS Protocol
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

package types

import (
	"math"
	"math/big"
)

const (
	BNBAssetId        = 0
	NilPairIndex      = -1
	NilAssetId        = 0
	NilBlockHeight    = -1
	NilNonce          = 0
	NilAssetInfo      = "{}"
	NilAccountName    = ""
	NilAccountOrder   = -1
	NilExpiredAt      = math.MaxInt64
	NilCollectionId   = int64(0)
	NilAccountIndex   = int64(0)
	NilTxNftIndex     = int64(-1)
	NilTxAccountIndex = int64(-1)
	BNBDecimalsStr    = "1000000000000000000"
)

var (
	NilAssetAmountStr           = "0"
	NilNftContentHash           = "0"
	NilL1TokenId                = "0"
	NilL1Address                = "0"
	NilOfferCanceledOrFinalized = big.NewInt(0)
)
