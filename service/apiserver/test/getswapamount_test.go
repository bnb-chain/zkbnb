package test

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/zkbas/service/apiserver/internal/types"
)

func (s *ApiServerSuite) TestGetSwapAmount() {
	type args struct {
		pairIndex   uint32
		assetId     uint32
		assetAmount string
		isFrom      bool
	}

	type testcase struct {
		name     string
		args     args
		httpCode int
	}

	tests := []testcase{
		{"not found", args{math.MaxUint32, math.MaxUint32, "1", true}, 400},
	}

	statusCode, pairs := GetPairs(s, 0, 100)
	if statusCode == http.StatusOK && len(pairs.Pairs) > 0 {
		for _, pair := range pairs.Pairs {
			if pair.TotalLpAmount != "" && pair.TotalLpAmount != "0" {
				tests = append(tests, []testcase{
					{"found by index with from is true", args{pair.Index, pair.AssetAId, "9000", true}, 200},
					{"found by index with from is true", args{pair.Index, pair.AssetBId, "9000", true}, 200},
					{"found by index with from is false", args{pair.Index, pair.AssetAId, "9000", false}, 200},
					{"found by index with from is false", args{pair.Index, pair.AssetBId, "9000", false}, 200},
				}...)
				break
			}
		}
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetSwapAmount(s, tt.args.pairIndex, tt.args.assetId, tt.args.assetAmount, tt.args.isFrom)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				assert.NotNil(t, result.AssetId)
				assert.NotNil(t, result.AssetAmount)
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func GetSwapAmount(s *ApiServerSuite, pairIndex, assetId uint32, assetAmount string, isFrom bool) (int, *types.SwapAmount) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/swapAmount?pair_index=%d&asset_id=%d&asset_amount=%s&is_from=%v", s.url, pairIndex, assetId, assetAmount, isFrom))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.SwapAmount{}
	//nolint: errcheck
	json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
