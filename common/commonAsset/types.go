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

package commonAsset

import (
	"encoding/json"
	"math/big"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/common/model/account"
)

type AccountAsset struct {
	AssetId                  int64
	Balance                  *big.Int
	LpAmount                 *big.Int
	OfferCanceledOrFinalized *big.Int
}

func ConstructAccountAsset(assetId int64, balance *big.Int, lpAmount *big.Int, offerCanceledOrFinalized *big.Int) *AccountAsset {
	return &AccountAsset{
		assetId,
		balance,
		lpAmount,
		offerCanceledOrFinalized,
	}
}

func ParseAccountAsset(balance string) (asset *AccountAsset, err error) {
	err = json.Unmarshal([]byte(balance), &asset)
	if err != nil {
		logx.Errorf("[ParseAccountAsset] unable to parse account asset")
		return nil, err
	}
	return asset, nil
}

func (asset *AccountAsset) String() (info string) {
	infoBytes, _ := json.Marshal(asset)
	return string(infoBytes)
}

type AccountInfo struct {
	AccountId       uint
	AccountIndex    int64
	AccountName     string
	PublicKey       string
	AccountNameHash string
	L1Address       string
	Nonce           int64
	CollectionNonce int64
	// map[int64]*AccountAsset
	AssetInfo map[int64]*AccountAsset // key: index, value: balance
	AssetRoot string
	Status    int
}

func FromFormatAccountInfo(formatAccountInfo *AccountInfo) (accountInfo *account.Account, err error) {
	assetInfoBytes, err := json.Marshal(formatAccountInfo.AssetInfo)
	if err != nil {
		return nil, err
	}
	accountInfo = &account.Account{
		Model: gorm.Model{
			ID: formatAccountInfo.AccountId,
		},
		AccountIndex:    formatAccountInfo.AccountIndex,
		AccountName:     formatAccountInfo.AccountName,
		PublicKey:       formatAccountInfo.PublicKey,
		AccountNameHash: formatAccountInfo.AccountNameHash,
		L1Address:       formatAccountInfo.L1Address,
		Nonce:           formatAccountInfo.Nonce,
		CollectionNonce: formatAccountInfo.CollectionNonce,
		AssetInfo:       string(assetInfoBytes),
		AssetRoot:       formatAccountInfo.AssetRoot,
		Status:          formatAccountInfo.Status,
	}
	return accountInfo, nil
}

func ToFormatAccountInfo(accountInfo *account.Account) (formatAccountInfo *AccountInfo, err error) {
	var (
		assetInfo map[int64]*AccountAsset
	)
	err = json.Unmarshal([]byte(accountInfo.AssetInfo), &assetInfo)
	if err != nil {
		return nil, err
	}
	formatAccountInfo = &AccountInfo{
		AccountId:       accountInfo.ID,
		AccountIndex:    accountInfo.AccountIndex,
		AccountName:     accountInfo.AccountName,
		PublicKey:       accountInfo.PublicKey,
		AccountNameHash: accountInfo.AccountNameHash,
		L1Address:       accountInfo.L1Address,
		Nonce:           accountInfo.Nonce,
		CollectionNonce: accountInfo.CollectionNonce,
		AssetInfo:       assetInfo,
		AssetRoot:       accountInfo.AssetRoot,
		Status:          accountInfo.Status,
	}
	return formatAccountInfo, nil
}

type FormatAccountHistoryInfo struct {
	AccountId       uint
	AccountIndex    int64
	Nonce           int64
	CollectionNonce int64
	// map[int64]*AccountAsset
	AssetInfo map[int64]*AccountAsset
	AssetRoot string
	// map[int64]*Liquidity
	L2BlockHeight int64
	Status        int
}
