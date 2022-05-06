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

package logic

type AccountSingleAsset struct {
	AccountId    uint
	AssetId      uint32
	AccountIndex uint32
	AccountName  string
	BalanceEnc   string
	PublicKey    string
}

type AccountSingleLockedAsset struct {
	// account info
	AccountId    uint
	AccountIndex uint32
	AccountName  string
	PublicKey    string
	// locked asset info
	ChainId      uint8
	AssetName    string
	LockAssetId  uint32
	AssetId      uint32
	LockedAmount uint64
}
