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

package util

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/zecrey-labs/zecrey-legend/common/model/basic"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"log"
	"testing"
)

var (
	mempoolModel = mempool.NewMempoolModel(basic.Connection, basic.CacheConf, basic.DB)
)

func TestConvertTxToRegisterZNSPubData(t *testing.T) {
	txInfo, err := mempoolModel.GetMempoolTxByTxId(1)
	if err != nil {
		t.Fatal(err)
	}
	pubData, err := ConvertTxToRegisterZNSPubData(txInfo)
	if err != nil {
		t.Fatal(err)
	}
	log.Println(common.Bytes2Hex(pubData))
}
