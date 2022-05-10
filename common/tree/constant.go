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

package tree

import (
	"github.com/zecrey-labs/zecrey-core/common/general/model/nft"
	"github.com/zecrey-labs/zecrey-core/common/general/model/sysconfig"
	"github.com/zecrey-labs/zecrey-crypto/accumulators/merkleTree"
	"github.com/zecrey-labs/zecrey-legend/common/model/account"
	"github.com/zecrey-labs/zecrey-legend/common/model/liquidity"
)

type (
	SysconfigModel        = sysconfig.SysconfigModel
	AccountModel          = account.AccountModel
	AccountHistoryModel   = account.AccountHistoryModel
	L2NftHistoryModel     = nft.L2NftHistoryModel
	LiquidityHistoryModel = liquidity.LiquidityHistoryModel

	AccountHistory      = account.AccountHistory
	AccountL2NftHistory = nft.L2NftHistory

	Tree = merkleTree.Tree
	Node = merkleTree.Node
)

const (
	AccountTreeHeight   = 32
	AssetTreeHeight     = 16
	LiquidityTreeHeight = 16
	NftTreeHeight       = 40
)

var (
	NilHash = merkleTree.NilHash
)
