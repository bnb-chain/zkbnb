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
	"errors"
	"github.com/bnb-chain/zkbas-crypto/ffmath"
	"log"
	"math/big"
)

func BigIntStringAdd(a, b string) (c string, err error) {
	aInt, isValid := new(big.Int).SetString(a, Base)
	if !isValid {
		log.Println("[BigIntStringAdd] invalid big int")
		return "", errors.New("[BigIntStringAdd] invalid big int")
	}
	bInt, isValid := new(big.Int).SetString(b, Base)
	if !isValid {
		log.Println("[BigIntStringAdd] invalid big int")
		return "", errors.New("[BigIntStringAdd] invalid big int")
	}
	c = ffmath.Add(aInt, bInt).String()
	return c, nil
}
