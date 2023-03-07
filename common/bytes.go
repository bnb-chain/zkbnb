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

package common

import (
	"encoding/binary"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	types2 "github.com/bnb-chain/zkbnb-crypto/circuit/types"
	"github.com/bnb-chain/zkbnb/types"
)

func ReadUint8(buf []byte, offset int) (newOffset int, res uint8) {
	return offset + 1, buf[offset]
}

func ReadUint16(buf []byte, offset int) (newOffset int, res uint16) {
	res = binary.BigEndian.Uint16(buf[offset : offset+2])
	return offset + 2, res
}

func ReadUint24(buf []byte, offset int) (newOffset int, res uint64) {
	res = binary.BigEndian.Uint64(buf[offset : offset+3])
	return offset + 3, res
}

func ReadUint32(buf []byte, offset int) (newOffset int, res uint32) {
	res = binary.BigEndian.Uint32(buf[offset : offset+4])
	return offset + 4, res
}

func ReadUint40(buf []byte, offset int) (newOffset int, res int64) {
	return offset + 5, new(big.Int).SetBytes(buf[offset : offset+5]).Int64()
}

func ReadUint128(buf []byte, offset int) (newOffset int, res *big.Int) {
	return offset + 16, new(big.Int).SetBytes(buf[offset : offset+16])
}

func ReadUint256(buf []byte, offset int) (newOffset int, res *big.Int) {
	return offset + 32, new(big.Int).SetBytes(buf[offset : offset+32])
}

func ReadBytes32(buf []byte, offset int) (newOffset int, res []byte) {
	res = make([]byte, 32)
	copy(res[:], buf[offset:offset+32])
	return offset + 32, res
}

func ReadPubKey(buf []byte, offset int) (newOffset int, res []byte) {
	res = make([]byte, 64)
	copy(res[:], buf[offset:offset+64])
	return offset + 64, res
}

func ReadBytes20(buf []byte, offset int) (newOffset int, res []byte) {
	res = make([]byte, 20)
	copy(res[:], buf[offset:offset+20])
	return offset + 20, res
}
func ReadBytes65(buf []byte, offset int) (newOffset int, res []byte) {
	res = make([]byte, 65)
	copy(res[:], buf[offset:offset+65])
	return offset + 65, res
}
func ReadAddress(buf []byte, offset int) (newOffset int, res string) {
	res = common.BytesToAddress(buf[offset : offset+20]).Hex()
	return offset + 20, res
}

func ReadPrefixPaddingBufToChunkSize(buf []byte, offset int) (newOffset int, res []byte) {
	return offset + 32, new(big.Int).SetBytes(buf[offset : offset+32]).Bytes()
}

func ReadAccountNameFromBytes20(buf []byte, offset int) (newOffset int, res []byte) {
	res = make([]byte, 20)
	copy(res[:], buf[offset:offset+20])
	return offset + 20, res
}

func PrefixPaddingBufToChunkSize(buf []byte) []byte {
	return new(big.Int).SetBytes(buf).FillBytes(make([]byte, 32))
}

func SuffixPaddingBufToChunkSize(buf []byte) []byte {
	res := make([]byte, 32)
	copy(res[:], buf[:])
	return res
}

func SuffixPaddingBuToPubdataSize(buf []byte) []byte {
	res := make([]byte, types2.PubDataBitsSizePerTx/8)
	copy(res[:], buf[:])
	return res
}

func AccountNameToBytes20(accountName string) []byte {
	realName := strings.Split(accountName, types.AccountNameSuffix)[0]
	buf := make([]byte, 20)
	copy(buf[:], realName)
	return buf
}

func AddressStrToBytes(addr string) []byte {
	return new(big.Int).SetBytes(common.FromHex(addr)).FillBytes(make([]byte, 20))
}

func PubKeyStrToBytes(pubKey string) []byte {
	return new(big.Int).SetBytes(common.FromHex(pubKey)).FillBytes(make([]byte, 64))
}

func SignatureStrToBytes(signature []byte) []byte {
	return new(big.Int).SetBytes(signature).FillBytes(make([]byte, 65))
}

func Uint16ToBytes(a uint16) []byte {
	return new(big.Int).SetUint64(uint64(a)).FillBytes(make([]byte, 2))
}

func Uint24ToBytes(a int64) []byte {
	return new(big.Int).SetInt64(a).FillBytes(make([]byte, 3))
}

func Uint32ToBytes(a uint32) []byte {
	return new(big.Int).SetUint64(uint64(a)).FillBytes(make([]byte, 4))
}

func Uint40ToBytes(a int64) []byte {
	return new(big.Int).SetInt64(a).FillBytes(make([]byte, 5))
}

func Uint128ToBytes(a *big.Int) []byte {
	return a.FillBytes(make([]byte, 16))
}

func Uint256ToBytes(a *big.Int) []byte {
	return a.FillBytes(make([]byte, 32))
}

func AmountToPackedAmountBytes(a *big.Int) (res []byte, err error) {
	packedAmount, err := ToPackedAmount(a)
	if err != nil {
		return nil, err
	}
	return Uint40ToBytes(packedAmount), nil
}

func FeeToPackedFeeBytes(a *big.Int) (res []byte, err error) {
	packedFee, err := ToPackedFee(a)
	if err != nil {
		return nil, err
	}
	return Uint16ToBytes(uint16(packedFee)), nil
}
