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

package util

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
)

/*
	Func: GetAccountAssetGlobalKey
	Params: accountIndex uint32, assetId uint32
	Return: string
	Description: Generating the leaf index for accountAssetTree, and fetch account asset info from BalanceDeltaMap
				 Used for BalanceDelta Map.
*/
func GetAccountAssetGlobalKey(accountIndex uint32, assetId uint32) string {
	h := mimc.NewMiMC()
	// H(accountIndex | assetId)
	var buf bytes.Buffer
	accountIndexBytes := make([]byte, 4)
	assetIdBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(accountIndexBytes, accountIndex)
	binary.BigEndian.PutUint32(assetIdBytes, assetId)
	buf.WriteString(AccountAssetPrefix)
	buf.Write(accountIndexBytes)
	buf.Write(assetIdBytes)
	h.Write(buf.Bytes())
	id := h.Sum([]byte{})
	h.Reset()

	return hex.EncodeToString(id)
}

/*
	Func: GetAccountNftGlobalKey
	Params: accountIndex uint32, nftAssetId int64
	Return: string
*/
func GetAccountNftGlobalKey(accountIndex uint32, nftAssetId int64) string {
	h := mimc.NewMiMC()
	// H(accountIndex | assetId)
	var buf bytes.Buffer
	accountIndexBytes := make([]byte, 4)
	nftAssetIdBytes := make([]byte, 8)
	binary.BigEndian.PutUint32(accountIndexBytes, accountIndex)
	binary.BigEndian.PutUint64(nftAssetIdBytes, uint64(nftAssetId))
	buf.WriteString(AccountNftPrefix)
	buf.Write(accountIndexBytes)
	buf.Write(nftAssetIdBytes)
	h.Write(buf.Bytes())
	id := h.Sum([]byte{})
	h.Reset()

	return hex.EncodeToString(id)
}

/*
	Func: GetPoolLiquidityGlobalKey
	Params: accountIndex uint32, pairIndex uint32, assetId uint32
	Return: string
	Description: Generating the leaf index for accountLiquidityTree
				 Used for LiquidityPoolDelta Map.
				 Account Index always equals to GasAccountIndex.
*/
func GetPoolLiquidityGlobalKey(accountIndex uint32, pairIndex uint32) string {
	h := mimc.NewMiMC()
	// H(pairIndex)
	var buf bytes.Buffer
	accountIndexBytes := make([]byte, 4)
	pairIndexBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(accountIndexBytes, accountIndex)
	binary.BigEndian.PutUint32(pairIndexBytes, pairIndex)
	buf.WriteString(PoolLiquidityPrefix)
	buf.Write(pairIndexBytes)
	h.Write(buf.Bytes())
	id := h.Sum([]byte{})
	h.Reset()

	return hex.EncodeToString(id)
}

/*
	Func: GetAccountLPGlobalKey
	Params: accountIndex uint32, pairIndex uint32
	Return: string
	Description: Generating the leaf index for accountLiquidityTree
				 Used for LPDelta Map.
*/
func GetAccountLPGlobalKey(accountIndex uint32, pairIndex uint32) string {
	h := mimc.NewMiMC()
	// H(accountIndex | pairIndex )
	var buf bytes.Buffer
	accountIndexBytes := make([]byte, 4)
	pairIndexBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(accountIndexBytes, accountIndex)
	binary.BigEndian.PutUint32(pairIndexBytes, pairIndex)
	buf.WriteString(LpPrefix)
	buf.Write(accountIndexBytes)
	buf.Write(pairIndexBytes)
	h.Write(buf.Bytes())
	id := h.Sum([]byte{})
	h.Reset()

	return hex.EncodeToString(id)
}
