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
	"encoding/binary"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

func ReadUint8(buf []byte, offset int) (newOffset int, res uint8) {
	return offset + 1, buf[offset]
}

func ReadUint16(buf []byte, offset int) (newOffset int, res uint16) {
	res = binary.BigEndian.Uint16(buf[offset : offset+2])
	return offset + 2, res
}

func ReadUint32(buf []byte, offset int) (newOffset int, res uint32) {
	res = binary.BigEndian.Uint32(buf[offset : offset+4])
	return offset + 4, res
}

func ReadUint40(buf []byte, offset int) (newOffset int, res int64) {
	return offset + 5, new(big.Int).SetBytes(buf[offset : offset+5]).Int64()
}

func ReadUint64(buf []byte, offset int) (newOffset int, res uint64) {
	res = binary.BigEndian.Uint64(buf[offset : offset+8])
	return offset + 8, res
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

func ReadAddress(buf []byte, offset int) (newOffset int, res string) {
	res = common.BytesToAddress(buf[offset : offset+20]).Hex()
	return offset + 20, res
}
