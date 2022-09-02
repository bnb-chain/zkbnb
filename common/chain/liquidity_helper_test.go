/*
 * Copyright © 2021 ZkBAS Protocol
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
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/zkbas/types"
)

func TestComputeDeltaY(t *testing.T) {
	poolA := big.NewInt(1000)
	poolB := big.NewInt(1000)
	deltaY, assetId, err := ComputeDelta(
		poolA, poolB,
		0, 2, 0, false, big.NewInt(500),
		30,
	)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, deltaY.String(), "1004")
	assert.Equal(t, assetId, int64(2))
}

func TestComputeRemoveLiquidityAmount(t *testing.T) {
	liquidityInfo := &types.LiquidityInfo{
		PairIndex:            0,
		AssetAId:             0,
		AssetA:               big.NewInt(99901),
		AssetBId:             2,
		AssetB:               big.NewInt(100100),
		LpAmount:             big.NewInt(100000),
		KLast:                big.NewInt(10000000000),
		FeeRate:              30,
		TreasuryAccountIndex: 0,
		TreasuryRate:         5,
	}
	aAmount, bAmount, _ := ComputeRemoveLiquidityAmount(
		liquidityInfo,
		big.NewInt(100),
	)
	assert.Equal(t, aAmount.Int64(), int64(99))
	assert.Equal(t, bAmount.Int64(), int64(100))
}

func TestComputeInputPrice(t *testing.T) {
	poolA := big.NewInt(1000)
	poolB := big.NewInt(1000)
	deltaY, _ := ComputeInputPrice(
		poolA, poolB,
		big.NewInt(500), 30,
	)
	assert.Equal(t, deltaY.Int64(), int64(332))
}

func TestComputeOutputPrice(t *testing.T) {
	poolA := big.NewInt(1000)
	poolB := big.NewInt(1000)
	deltaY, _ := ComputeOutputPrice(
		poolA, poolB,
		big.NewInt(500), 30,
	)
	assert.Equal(t, deltaY.Int64(), int64(1004))
}
