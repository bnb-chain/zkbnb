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

package globalmapHandler

import (
	"errors"
	"fmt"
	"github.com/zecrey-labs/zecrey-legend/common/model/account"
	"github.com/zeromicro/go-zero/core/logx"
)

func GetAccountInfoFromAccountAndAccountHistory(
	accountModel account.AccountModel,
	accountHistoryModel account.AccountHistoryModel,
	accountIndex int64) (accountInfo *account.Account, err error) {
	// get account info by account index
	// get data from account table
	accountInfo, err = accountModel.GetAccountByAccountIndex(accountIndex)
	if err != nil {
		logx.Errorf("[GetAccountInfoFromAccountAndAccountHistory] unable to get account by account index")
		return nil, err
	}
	latestNonce, err := accountHistoryModel.GetLatestAccountNonceByAccountIndex(accountIndex)
	if err != nil {
		if err != account.ErrNotFound {
			errInfo := fmt.Sprintf("[GetAccountInfoFromAccountAndAccountHistory] %s. invalid accountIndex %v",
				err.Error(), accountIndex)
			logx.Error(errInfo)
			return nil, errors.New(errInfo)
		}
	} else {
		accountInfo.Nonce = latestNonce
	}
	return accountInfo, nil
}
