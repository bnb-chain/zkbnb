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

package util

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"
)

func ParsePubKey(pkStr string) (pk *eddsa.PublicKey, err error) {
	pkBytes := common.FromHex(pkStr)
	pk = new(eddsa.PublicKey)
	_, err = pk.A.SetBytes(pkBytes)
	if err != nil {
		logx.Errorf("[ParsePubKey] unable to set pk bytes: %s", err.Error())
		return nil, err
	}
	return pk, nil
}
