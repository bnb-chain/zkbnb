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
	"fmt"
	"github.com/zecrey-labs/zecrey-crypto/ffmath"
	"github.com/zeromicro/go-zero/core/logx"
	"math/big"
	"strconv"
)

var (
	// 2^35 - 1
	PackedAmountMaxMantissa = big.NewInt(34359738367)
	// 2^11 - 1
	PackedFeeMaxMantissa  = big.NewInt(2047)
	PackedAmountMaxAmount = ffmath.Multiply(big.NewInt(34359738367), new(big.Int).Exp(big.NewInt(10), big.NewInt(31), nil))
	PackedFeeMaxAmount    = ffmath.Multiply(big.NewInt(2047), new(big.Int).Exp(big.NewInt(10), big.NewInt(31), nil))
	ZeroBigInt            = big.NewInt(0)
)

/*
	ToPackedAmount: convert big int to 40 bit, 5 bits for 10^x, 35 bits for a * 10^x
*/
func ToPackedAmount(amount *big.Int) (res int64, err error) {
	if amount.Cmp(ZeroBigInt) < 0 || amount.Cmp(PackedAmountMaxAmount) > 0 {
		logx.Errorf("[ToPackedAmount] invalid amount")
		return -1, errors.New("[ToPackedAmount] invalid amount")
	}
	oAmount := new(big.Int).Set(amount)
	exponent := int64(0)
	for oAmount.Cmp(PackedAmountMaxMantissa) > 0 {
		oAmount = ffmath.Div(oAmount, big.NewInt(10))
		exponent++
	}
	exponentBits := strconv.FormatInt(exponent, 2)
	for len(exponentBits) < 5 {
		exponentBits = "0" + exponentBits
	}
	mantissaBits := strconv.FormatInt(oAmount.Int64(), 2)
	packedAmountBits := mantissaBits + exponentBits
	packedAmount, err := strconv.ParseInt(packedAmountBits, 2, 40)
	if err != nil {
		logx.Errorf("[ToPackedAmount] unable to convert to packed amount: %s", err.Error())
		return -1, err
	}
	fmt.Println(exponentBits)
	fmt.Println(mantissaBits)
	fmt.Println(strconv.FormatInt(packedAmount, 2))
	return packedAmount, nil
}

func CleanPackedAmount(amount *big.Int) (nAmount *big.Int, err error) {
	if amount.Cmp(ZeroBigInt) < 0 || amount.Cmp(PackedAmountMaxAmount) > 0 {
		logx.Errorf("[ToPackedAmount] invalid amount")
		return nil, errors.New("[ToPackedAmount] invalid amount")
	}
	oAmount := new(big.Int).Set(amount)
	exponent := int64(0)
	for oAmount.Cmp(PackedAmountMaxMantissa) > 0 {
		oAmount = ffmath.Div(oAmount, big.NewInt(10))
		exponent++
	}
	nAmount = ffmath.Multiply(oAmount, new(big.Int).Exp(big.NewInt(10), big.NewInt(exponent), nil))
	return nAmount, nil
}

/*
	ToPackedFee: convert big int to 16 bit, 5 bits for 10^x, 11 bits for a * 10^x
*/
func ToPackedFee(amount *big.Int) (res uint16, err error) {
	if amount.Cmp(ZeroBigInt) < 0 || amount.Cmp(PackedFeeMaxAmount) > 0 {
		logx.Errorf("[ToPackedFee] invalid amount")
		return 0, errors.New("[ToPackedFee] invalid amount")
	}
	oAmount := new(big.Int).Set(amount)
	exponent := int64(0)
	for oAmount.Cmp(PackedFeeMaxMantissa) > 0 {
		oAmount = ffmath.Div(oAmount, big.NewInt(10))
		exponent++
	}
	exponentBits := strconv.FormatInt(exponent, 2)
	for len(exponentBits) < 5 {
		exponentBits = "0" + exponentBits
	}
	mantissaBits := strconv.FormatInt(oAmount.Int64(), 2)
	packedFeeBits := mantissaBits + exponentBits
	packedFee, err := strconv.ParseInt(packedFeeBits, 2, 16)
	if err != nil {
		logx.Errorf("[ToPackedFee] unable to convert to packed fee: %s", err.Error())
		return 0, err
	}
	return uint16(packedFee), nil
}

func CleanPackedFee(amount *big.Int) (nAmount *big.Int, err error) {
	if amount.Cmp(ZeroBigInt) < 0 || amount.Cmp(PackedFeeMaxAmount) > 0 {
		logx.Errorf("[ToPackedFee] invalid amount")
		return nil, errors.New("[ToPackedFee] invalid amount")
	}
	oAmount := new(big.Int).Set(amount)
	exponent := int64(0)
	for oAmount.Cmp(PackedFeeMaxMantissa) > 0 {
		oAmount = ffmath.Div(oAmount, big.NewInt(10))
		exponent++
	}
	nAmount = ffmath.Multiply(oAmount, new(big.Int).Exp(big.NewInt(10), big.NewInt(exponent), nil))
	return nAmount, nil
}
