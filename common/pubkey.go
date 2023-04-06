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
	curve "github.com/bnb-chain/zkbnb-crypto/ecc/ztwistededwards/tebn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	"github.com/ethereum/go-ethereum/common"
)

func ParsePubKey(pkStr string) (pk *eddsa.PublicKey, err error) {
	if pkStr == "0000000000000000000000000000000000000000000000000000000000000000" {
		pk := &eddsa.PublicKey{
			A: curve.Point{
				X: fr.NewElement(0),
				Y: fr.NewElement(0),
			},
		}
		return pk, nil
	}
	pkBytes := common.FromHex(pkStr)
	pk = new(eddsa.PublicKey)
	_, err = pk.A.SetBytes(pkBytes)
	if err != nil {
		return nil, err
	}
	return pk, nil
}

func EmptyPubKey() string {
	pk := &eddsa.PublicKey{
		A: curve.Point{
			X: fr.NewElement(0),
			Y: fr.NewElement(0),
		},
	}
	return common.Bytes2Hex(pk.Bytes())
}
