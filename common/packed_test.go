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
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToPackedAmount(t *testing.T) {
	a, _ := new(big.Int).SetString("34359738361", 10)
	amount, err := ToPackedAmount(a)
	assert.NoError(t, err)
	assert.Equal(t, amount, int64(1099511627552))
}

func TestToPackedFee(t *testing.T) {
	amount, _ := new(big.Int).SetString("100000000000000", 10)
	fee, err := ToPackedFee(amount)
	assert.NoError(t, err)
	assert.Equal(t, fee, int64(32011))
}
