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

package chain

import (
	"errors"

	"github.com/bnb-chain/zkbnb-crypto/ffmath"
	"github.com/bnb-chain/zkbnb/types"
)

func ComputeNewBalance(assetType int64, balance string, balanceDelta string) (newBalance string, err error) {
	switch assetType {
	case types.FungibleAssetType:
		assetInfo, err := types.ParseAccountAsset(balance)
		if err != nil {
			return "", err
		}
		assetDelta, err := types.ParseAccountAsset(balanceDelta)
		if err != nil {
			return "", err
		}
		assetInfo.Balance = ffmath.Add(assetInfo.Balance, assetDelta.Balance)
		if assetDelta.OfferCanceledOrFinalized == nil {
			assetDelta.OfferCanceledOrFinalized = types.ZeroBigInt
		}
		if assetDelta.OfferCanceledOrFinalized.Cmp(types.EmptyOfferCanceledOrFinalized) != 0 {
			assetInfo.OfferCanceledOrFinalized = assetDelta.OfferCanceledOrFinalized
		}
		newBalance = assetInfo.String()
	case types.NftAssetType:
		// just set the old one as the new one
		newBalance = balanceDelta
	default:
		return "", errors.New("[ComputeNewBalance] invalid asset type")
	}
	return newBalance, nil
}
