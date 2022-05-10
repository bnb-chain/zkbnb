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

package commonConstant

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/zecrey-labs/zecrey-legend/common/tree"
)

const (
	NilAssetId     = -1
	NilBlockHeight = -1
	NilNonce       = -1
	EmptyAsset     = "{}"
	EmptyLiquidity = "{}"
	NilAccountName   = ""
)

var (
	NilHash           = tree.NilHash
	NilHashStr        = common.Bytes2Hex(tree.NilHash)
	NilAssetAmountStr = "0"
	NilNftContentHash = "0"
	NilL1TokenId      = "-1"
	NilL1Address      = "0"
	NilAccountIndex   = int64(-1)
	NilCollectionId   = int64(-1)
)
