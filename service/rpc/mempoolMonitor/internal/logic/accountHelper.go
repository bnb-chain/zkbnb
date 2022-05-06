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

package logic

import (
	"github.com/zecrey-labs/zecrey-legend/common/model/account"
	"github.com/zeromicro/go-zero/core/logx"
)

func GetAccountInfoByAccountNameHash(accountNameHash string, accountModel account.AccountModel, accountHistoryModel account.AccountHistoryModel) (accountInfo *account.Account, err error) {
	accountHistoryInfo, err := accountHistoryModel.GetAccountByAccountNameHash(accountNameHash)
	if err != nil {
		if err == ErrNotFound {
			accountInfo, err = accountModel.GetAccountByAccountNameHash(accountNameHash)
			if err != nil {
				logx.Errorf("[MonitorMempool] unable to get account by account name hash: %s", err.Error())
				return nil, err
			}
		} else {
			logx.Errorf("[MonitorMempool] unable to get account history by account name hash: %s", err.Error())
			return nil, err
		}
	} else {
		accountInfo = &account.Account{
			AccountIndex:    accountHistoryInfo.AccountIndex,
			AccountName:     accountHistoryInfo.AccountName,
			PublicKey:       accountHistoryInfo.PublicKey,
			AccountNameHash: accountHistoryInfo.AccountNameHash,
			L1Address:       accountHistoryInfo.L1Address,
			Nonce:           accountHistoryInfo.Nonce,
		}
	}
	return accountInfo, nil
}
