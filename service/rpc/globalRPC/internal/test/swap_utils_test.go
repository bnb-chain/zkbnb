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

package test

import (
	"github.com/stretchr/testify/assert"
	"github.com/zecrey-labs/zecrey-crypto/ffmath"
	"math"
	"math/big"
	"testing"
)

func TestSwapUtils(t *testing.T) {
	real_B_A_DeltaBigInt := ffmath.Div(ffmath.Multiply(big.NewInt(int64(9)), big.NewInt(int64(10000-30))), big.NewInt(int64(10000)))
	real_B_A_Delta := int64(math.Floor(float64(real_B_A_DeltaBigInt.Uint64())))

	assert.Equal(t, int64(8), real_B_A_Delta)
}
