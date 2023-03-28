/*
 * Copyright © 2021 ZkBNB Protocol
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
	"encoding/hex"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/poseidon"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	AccountTreeHeight = 32
	AssetTreeHeight   = 16
	NftTreeHeight     = 40
)

const (
	NFTPrefix          = "nft:"
	AccountPrefix      = "account:"
	AccountAssetPrefix = "account_asset:"
)

var (
	NilStateRoot            []byte
	NilAccountRoot          []byte
	NilNftRoot              []byte
	NilAccountAssetRoot     []byte
	NilAccountNodeHash      []byte
	NilNftNodeHash          []byte
	NilAccountAssetNodeHash []byte
)

func init() {
	NilAccountAssetNodeHash = EmptyAccountAssetNodeHash()
	NilAccountAssetRoot = NilAccountAssetNodeHash
	for i := 0; i < AssetTreeHeight; i++ {
		NilAccountAssetRoot = poseidon.PoseidonBytes(NilAccountAssetRoot, NilAccountAssetRoot)
	}
	NilAccountNodeHash = EmptyAccountNodeHash()
	NilAccountRoot = NilAccountNodeHash
	NilNftNodeHash = EmptyNftNodeHash()
	for i := 0; i < AccountTreeHeight; i++ {
		NilAccountRoot = poseidon.PoseidonBytes(NilAccountRoot, NilAccountRoot)
	}
	NilNftRoot = NilNftNodeHash
	for i := 0; i < NftTreeHeight; i++ {
		NilNftRoot = poseidon.PoseidonBytes(NilNftRoot, NilNftRoot)
	}
	// nil state root
	NilStateRoot = poseidon.PoseidonBytes(NilAccountRoot, NilNftRoot)

	logx.Infof("genesis state root: %s asset %s", hex.EncodeToString(NilStateRoot), hex.EncodeToString(NilAccountAssetRoot))
}
