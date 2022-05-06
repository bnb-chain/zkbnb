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
	"errors"
	"log"
	"math/big"
)

/*
	AccountAssetInfo: for account asset info
*/
type AccountAssetInfo struct {
	// account info
	AccountId       uint
	AccountIndex    int64
	AccountName     string
	AccountNameHash string
	PublicKey       *PublicKey
	Nonce           int64
	// asset info
	AssetId int64
	Balance *big.Int
}

/*
	ConstructAccountAssetInfo: helper function for constructing the data struct for crypto
*/
func ConstructAccountAssetInfo(
	AccountId uint,
	AccountIndex int64,
	AccountName string,
	AccountNameHash string,
	PublicKey string,
	nonce int64,
	AssetId int64,
	Balance string,
) (accountInfo *AccountAssetInfo, err error) {
	// parse pk
	pk, err := ParsePkStr(PublicKey)
	if err != nil {
		return nil, err
	}
	balance, isValid := new(big.Int).SetString(Balance, Base)
	if !isValid {
		log.Println("[ConstructAccountAssetInfo] invalid big int string")
		return nil, errors.New("[ConstructAccountAssetInfo] invalid big int string")
	}
	accountInfo = &AccountAssetInfo{
		AccountId:       AccountId,
		AccountIndex:    AccountIndex,
		AccountName:     AccountName,
		AccountNameHash: AccountNameHash,
		PublicKey:       pk,
		Nonce:           nonce,
		AssetId:         AssetId,
		Balance:         balance,
	}
	return accountInfo, nil
}

type AccountLiquidityInfo struct {
	// account info
	AccountId    uint
	AccountIndex int64
	AccountName  string
	PublicKey    *PublicKey
	Nonce        int64
	// liquidity info
	PairIndex     int64
	AssetAId      int64
	AssetBId      int64
	AssetABalance *big.Int
	AssetBBalance *big.Int
	FeeRate       int64
	TreasuryRate  int64
	// lp info
	LpAmount *big.Int
}

func ConstructAccountLiquidityInfo(
	// account info
	AccountId uint,
	AccountIndex int64,
	AccountName string,
	PublicKey string,
	nonce int64,
	// liquidity info
	PairIndex int64,
	AssetAId int64,
	AssetBId int64,
	AssetABalance string,
	AssetBBalance string,
	FeeRate int64,
	TreasuryRate int64,
	lpAmount string,
) (accountInfo *AccountLiquidityInfo, err error) {
	// parse pk
	pk, err := ParsePkStr(PublicKey)
	if err != nil {
		return nil, err
	}
	// parse big int
	assetABalance, isValid := new(big.Int).SetString(AssetABalance, Base)
	if !isValid {
		return nil, errors.New("[ConstructAccountLiquidityInfo] invalid asset a balance")
	}
	assetBBalance, isValid := new(big.Int).SetString(AssetBBalance, Base)
	if !isValid {
		return nil, errors.New("[ConstructAccountLiquidityInfo] invalid asset b balance")
	}
	// lp amount
	lpBalance, isValid := new(big.Int).SetString(lpAmount, Base)
	if !isValid {
		return nil, errors.New("[ConstructAccountLiquidityInfo] invalid asset b balance")
	}
	accountInfo = &AccountLiquidityInfo{
		AccountId:     AccountId,
		AccountIndex:  AccountIndex,
		AccountName:   AccountName,
		PublicKey:     pk,
		Nonce:         nonce,
		PairIndex:     PairIndex,
		AssetAId:      AssetAId,
		AssetBId:      AssetBId,
		AssetABalance: assetABalance,
		AssetBBalance: assetBBalance,
		FeeRate:       FeeRate,
		TreasuryRate:  TreasuryRate,
		LpAmount:      lpBalance,
	}
	return accountInfo, nil
}

func ConstructNftInfo(
	// nft info
	NftIndex int64,
	CreatorAccountIndex int64,
	OwnerAccountIndex int64,
	AssetId int64,
	AssetAmount string,
	NftContentHash string,
	NftL1TokenId string,
	NftL1Address string,
) (accountInfo *NftInfo, err error) {
	assetAmount, isValid := new(big.Int).SetString(AssetAmount, Base)
	if !isValid {
		log.Println("[ConstructNftInfo] invalid asset amount")
		return nil, errors.New("[ConstructNftInfo] invalid asset amount")
	}
	accountInfo = &NftInfo{
		NftIndex:            NftIndex,
		CreatorAccountIndex: CreatorAccountIndex,
		OwnerAccountIndex:   OwnerAccountIndex,
		AssetId:             AssetId,
		AssetAmount:         assetAmount,
		NftContentHash:      NftContentHash,
		NftL1TokenId:        NftL1TokenId,
		NftL1Address:        NftL1Address,
	}
	return accountInfo, nil
}
