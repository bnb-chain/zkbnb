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
	"fmt"
	"math/big"
	"testing"
)

func TestComputeDeltaY(t *testing.T) {
	poolA := big.NewInt(100000)
	poolB := big.NewInt(100000)
	deltaY, _, err := ComputeDelta(
		poolA, poolB,
		0, 2, 0, true, big.NewInt(100),
		30,
	)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(deltaY.String())
}
