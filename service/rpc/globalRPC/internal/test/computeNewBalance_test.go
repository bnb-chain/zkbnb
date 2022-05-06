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

package test

import (
	"github.com/stretchr/testify/assert"
	"github.com/zecrey-labs/zecrey/common/commonAsset"
	"github.com/zecrey-labs/zecrey/common/utils"
	"github.com/zeromicro/go-zero/core/logx"
	"testing"
)

func TestComputeNewBalance(t *testing.T) {
	balanceEnc := "EBAbkUxyhNR5OKdbfyDG5gFIen4e3GdF2Ha44xG0VaTf8qg+W7moJFSS/nlTtGSS4tlF6jQAB8F1e7Bq2SEEGg=="
	balanceDelta := "UfTYv2RXj8Sey7oEGpS01D1BiJT8MSXviXQKsqi7aC2pBcsBle6rzh0V/gzbXn8TDtqK2+Dl4U+Sxfu5GHwWEA=="
	res, err := utils.ComputeNewBalance(commonAsset.GeneralAssetType, balanceEnc, balanceDelta)
	assert.Nil(t, err)
	logx.Info(res)
}
