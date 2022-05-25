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
	"encoding/hex"
	"errors"
	"github.com/zecrey-labs/zecrey-crypto/elgamal/twistededwards/tebn254/twistedElgamal"
	"github.com/zecrey-labs/zecrey-crypto/zecrey/twistededwards/tebn254/zecrey"
	"github.com/zecrey-labs/zecrey-legend/common/commonConstant"
	"github.com/zeromicro/go-zero/core/logx"
	"log"
	"math/big"
)

func SetFixed32Bytes(buf []byte) [32]byte {
	newBuf := new(big.Int).SetBytes(buf).FillBytes(make([]byte, zecrey.PointSize))
	var res [zecrey.PointSize]byte
	copy(res[:], newBuf[:])
	return res
}

func WriteStringBigIntIntoBuf(buf *bytes.Buffer, aStr string) error {
	a, isValid := new(big.Int).SetString(aStr, 10)
	if !isValid {
		logx.Errorf("[WriteStringBigIntIntoBuf] invalid string")
		return errors.New("[WriteStringBigIntIntoBuf] invalid string")
	}
	buf.Write(a.FillBytes(make([]byte, zecrey.PointSize)))
	return nil
}

func WriteAccountNameIntoBuf(buf *bytes.Buffer, accountName string) {
	infoBytes := SetFixed32Bytes([]byte(accountName))
	buf.Write(new(big.Int).SetBytes(infoBytes[:]).FillBytes(make([]byte, zecrey.PointSize)))
}

func WriteAddressIntoBuf(buf *bytes.Buffer, address string) (err error) {
	if address == commonConstant.NilL1Address {
		buf.Write(new(big.Int).FillBytes(make([]byte, 32)))
		return nil
	}
	addrBytes, err := DecodeAddress(address)
	if err != nil {
		log.Println("[WriteAddressIntoBuf] invalid addr:", err)
		return err
	}
	buf.Write(new(big.Int).SetBytes(addrBytes).FillBytes(make([]byte, zecrey.PointSize)))
	return nil
}

func DecodeAddress(addr string) ([]byte, error) {
	if len(addr) != 42 {
		return nil, errors.New("[DecodeAddress] invalid address")
	}
	addrBytes, err := hex.DecodeString(addr[2:])
	if err != nil {
		return nil, err
	}
	if len(addrBytes) != AddressSize {
		log.Println("[DecodeAddress] invalid address")
		return nil, errors.New("[DecodeAddress] invalid address")
	}
	return addrBytes, nil
}

func WriteInt64IntoBuf(buf *bytes.Buffer, a int64) {
	buf.Write(new(big.Int).SetInt64(a).FillBytes(make([]byte, zecrey.PointSize)))
}

func WritePkIntoBuf(buf *bytes.Buffer, pkStr string) (err error) {
	pk, err := ParsePubKey(pkStr)
	if err != nil {
		logx.Errorf("[WriteEncIntoBuf] unable to parse pk: %s", err.Error())
		return err
	}
	writePointIntoBuf(buf, &pk.A)
	return nil
}

func writePointIntoBuf(buf *bytes.Buffer, p *zecrey.Point) {
	buf.Write(p.X.Marshal())
	buf.Write(p.Y.Marshal())
}

func WriteEncIntoBuf(buf *bytes.Buffer, encStr string) error {
	enc, err := twistedElgamal.FromString(encStr)
	if err != nil {
		logx.Errorf("[WriteEncIntoBuf] unable to parse enc str: %s", err.Error())
		return err
	}
	writePointIntoBuf(buf, enc.CL)
	writePointIntoBuf(buf, enc.CR)
	return nil
}
