/*
 * Copyright © 2021 Zkbas Protocol
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

package nft

const (
	L2NftTableName                = `l2_nft`
	L2NftHistoryTableName         = `l2_nft_history`
	L2NftCollectionTableName      = `l2_nft_collection`
	L2NftExchangeTableName        = `l2_nft_exchange`
	L2NftWithdrawHistoryTableName = `l2_nft_withdraw_history`

	OfferTableName = `offer`

	OfferFinishedStatus = 1

	CollectionPending = 0 // create collection request received by api
	CollectionCreated = 1 // collection created in l2
)
