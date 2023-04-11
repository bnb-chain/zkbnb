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
	"encoding/json"
	"github.com/zeromicro/go-zero/core/logx"
	"testing"
)

func TestParsePubDataForDesert(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			logx.Severef("failed, %v", err)
			panic("failed")
		}
	}()
	pubData := "0x0a000000080000010000e10a0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030303030"
	txInfos, err := ParsePubDataForDesert(pubData)
	if err != nil {
		logx.Error(err)
	}
	txInfosJson, _ := json.Marshal(txInfos)
	logx.Info(string(txInfosJson))
}
