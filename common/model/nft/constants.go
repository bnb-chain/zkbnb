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

package nft

import "github.com/zeromicro/go-zero/core/stores/sqlx"

const (
	TableName                = `l2_nft`
	HistoryTableName         = `l2_nft_history`
	CollectionTableName      = `l2_nft_collection`
	ExchangeTableName        = `l2_nft_exchange`
	WithdrawHistoryTableName = `l2_nft_withdraw_history`

	OfferTableName = `offer`

	StatusAlreadyWithdraw = 1

	OfferFinishedStatus = 1
)

var (
	ErrNotFound = sqlx.ErrNotFound
)
