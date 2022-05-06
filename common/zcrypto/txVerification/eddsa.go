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

package txVerification

import (
	"encoding/hex"
	"errors"
	"log"
)

/*
	ParsePkStr: parse pk string
*/
func ParsePkStr(pkStr string) (pk *PublicKey, err error) {
	pkBytes, err := hex.DecodeString(pkStr)
	if err != nil {
		log.Println("[ParsePkStr] invalid public key:", err)
		return nil, err
	}
	pk = new(PublicKey)
	size, err := pk.SetBytes(pkBytes)
	if err != nil {
		log.Println("[ParsePkStr] invalid public key:", err)
		return nil, err
	}
	if size != 32 {
		log.Println("[ParsePkStr] invalid public key")
		return nil, errors.New("[ParsePkStr] invalid public key")
	}
	return pk, nil
}
