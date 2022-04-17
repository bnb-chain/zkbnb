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
 */

package cryptoUtils

import (
	"fmt"
	"testing"
)

var pubArray = []string{}

func TestIsValidPublicKey(t *testing.T) {
	var finalMap = make(map[string]bool)
	for i := range pubArray {
		ok, err := IsValidPublicKey(pubArray[i])
		if err != nil || !ok {
			finalMap[pubArray[i]] = true
		}
	}

	for i := range finalMap {
		fmt.Println(i)
	}
}
