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
	"github.com/bnb-chain/zkbnb-crypto/util"
	"github.com/zeromicro/go-zero/core/logx"
	"math/big"
	"strconv"
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

func TestUnpackAmount(t *testing.T) {
	a, _ := strconv.ParseInt("1111111111111111111111111111111111111111", 2, 40)

	//a, _ := new(big.Int).SetString("1111111111111111111111111111111111111111", 2)
	logx.Info(a)
	nAmount, _ := util.UnpackAmount(new(big.Int).SetInt64(a))
	logx.Info(nAmount.String())
}

func TestUnpackFee(t *testing.T) {
	a, _ := strconv.ParseInt("1111111111111111", 2, 16)
	logx.Info(a)
	nAmount, _ := util.UnpackFee(new(big.Int).SetInt64(a))
	logx.Info(nAmount.String())

}

func TestCheckPackedAmount(t *testing.T) {
	amount, _ := new(big.Int).SetString("123456789012345678901234567890123456789012", 10)
	logx.Info(amount)
	packedAmount, _ := util.ToPackedAmount(amount)
	logx.Info(packedAmount)

	nAmount, _ := util.UnpackAmount(big.NewInt(packedAmount))
	logx.Info(nAmount.String())
}

//4570789518076018688
//4570789518076018688
//4570789518076018688
