/*
 * Copyright © 2021 Zkbas Protocol
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

package txVerification

import (
	"encoding/hex"
	"errors"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/zeromicro/go-zero/core/logx"
)

/*
	ParsePkStr: parse pk string
*/
func ParsePkStr(pkStr string) (pk *PublicKey, err error) {
	pkBytes, err := hex.DecodeString(pkStr)
	if err != nil {
		logx.Errorf("[ParsePkStr] invalid public key: %s", err.Error())
		return nil, err
	}
	pk = new(PublicKey)
	size, err := pk.SetBytes(pkBytes)
	if err != nil {
		logx.Errorf("[ParsePkStr] invalid public key: %s", err.Error())
		return nil, err
	}
	if size != 32 {
		logx.Error("[ParsePkStr] invalid public key")
		return nil, errors.New("[ParsePkStr] invalid public key")
	}
	return pk, nil
}

func VerifySignature(sig, msg []byte, pubkey string) error {
	hFunc := mimc.NewMiMC()
	pk, err := ParsePkStr(pubkey)
	if err != nil {
		return errors.New("cannot parse public key")
	}
	isValid, err := pk.Verify(sig, msg, hFunc)
	if err != nil {
		logx.Errorf("unable to verify signature: %s", err.Error())
		return errors.New("unable to verify signature")
	}
	if !isValid {
		logx.Errorf("invalid signature")
		return errors.New("invalid signature")
	}
	return nil
}
