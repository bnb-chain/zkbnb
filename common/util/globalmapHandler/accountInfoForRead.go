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

package globalmapHandler

import (
	"encoding/json"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/model/account"
	"github.com/bnb-chain/zkbas/common/util"
)

func GetBasicAccountInfo(
	accountModel AccountModel,
	redisConnection *Redis,
	accountIndex int64,
) (
	accountInfo *AccountInfo,
	err error,
) {
	key := util.GetBasicAccountKey(accountIndex)
	basicAccountInfoStr, err := redisConnection.Get(key)
	if err != nil {
		logx.Errorf("[GetBasicAccountInfo] unable to get account info: %s", err.Error())
		return nil, err
	}
	if basicAccountInfoStr == "" {
		oAccountInfo, err := accountModel.GetAccountByAccountIndex(accountIndex)
		if err != nil {
			logx.Errorf("[GetBasicAccountInfo] unable to get account by account index: %s", err.Error())
			return nil, err
		}
		accountInfo, err = commonAsset.ToFormatAccountInfo(oAccountInfo)
		if err != nil {
			logx.Errorf("[GetBasicAccountInfo] unable to get basic account info: %s", err.Error())
			return nil, err
		}
		// update cache
		oAccountInfoBytes, err := json.Marshal(oAccountInfo)
		if err != nil {
			logx.Errorf("[GetBasicAccountInfo] unable to marshal account info: %s", err.Error())
			return nil, err
		}
		_ = redisConnection.Setex(key, string(oAccountInfoBytes), BasicAccountExpiryTime)
	} else {
		var oAccountInfo *account.Account
		err = json.Unmarshal([]byte(basicAccountInfoStr), &oAccountInfo)
		if err != nil {
			logx.Errorf("[GetBasicAccountInfo] unable to parse account info: %s", err.Error())
			return nil, err
		}
		accountInfo, err = commonAsset.ToFormatAccountInfo(oAccountInfo)
		if err != nil {
			logx.Errorf("[GetBasicAccountInfo] unable to get basic account info: %s", err.Error())
			return nil, err
		}
	}
	return accountInfo, nil
}
