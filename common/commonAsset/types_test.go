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
	"fmt"
	"testing"
)

func TestSerialize(t *testing.T) {
	lpInfo := make(map[int64]*Liquidity)
	lpInfo[1] = &Liquidity{
		PairIndex: 5,
		AssetAId:  0,
		AssetA:    "",
		AssetBId:  0,
		AssetB:    "",
		LpAmount:  "",
	}
	lpInfo[2] = &Liquidity{
		PairIndex: 1,
		AssetAId:  1,
		AssetA:    "",
		AssetBId:  0,
		AssetB:    "",
		LpAmount:  "",
	}
	lpInfoBytes, err := json.Marshal(lpInfo)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(lpInfoBytes))
	var nLpInfo map[int64]*Liquidity
	err = json.Unmarshal(lpInfoBytes, &nLpInfo)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(nLpInfo[1].PairIndex)
	fmt.Println(nLpInfo[2].AssetAId)
}

func TestDeserializeAccountInfo(t *testing.T) {
	var nLpInfo map[int64]*Liquidity
	err := json.Unmarshal([]byte("{}"), &nLpInfo)
	if err != nil {
		t.Fatal(err)
	}
}
