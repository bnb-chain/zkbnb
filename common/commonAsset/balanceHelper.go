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
	"github.com/zeromicro/go-zero/core/logx"
)

/*
	ComputeNewBalance: helper function for computing new balance for different asset types
*/
func ComputeNewBalance(assetType int64, balance string, balanceDelta string) (newBalance string, err error) {
	switch assetType {
	case GeneralAssetType:
		newBalance, err = BigIntStringAdd(balance, balanceDelta)
		if err != nil {
			logx.Errorf("[ComputeNewBalance] unable to compute new balance: %s", err.Error())
			return "", err
		}
		break
	case LiquidityAssetType:
		// balance: PoolInfo
		newBalance, err = AddPoolInfoString(balance, balanceDelta)
		if err != nil {
			logx.Errorf("[ComputeNewBalance] unable to compute new balance: %s", err.Error())
			return "", err
		}
		break
	case LiquidityLpAssetType:
		newBalance, err = BigIntStringAdd(balance, balanceDelta)
		if err != nil {
			logx.Errorf("[ComputeNewBalance] unable to compute new balance: %s", err.Error())
			return "", err
		}
		break
	case NftAssetType:
		// just set the old one as the new one
		newBalance = balanceDelta
		break
	default:
		return "", errors.New("[ComputeNewBalance] invalid asset type")
	}
	return newBalance, nil
}
