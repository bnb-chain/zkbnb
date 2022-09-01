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

func (s *ApiServerSuite) TestGetLpValue() {
	type args struct {
		pairIndex int
		lpAmount  string
	}

	type testcase struct {
		name     string
		args     args
		httpCode int
	}

	tests := []testcase{
		{"not found", args{math.MaxInt, "2"}, 400},
	}

	statusCode, pairs := GetPairs(s, 0, 100)
	if statusCode == http.StatusOK && len(pairs.Pairs) > 0 {
		for _, pair := range pairs.Pairs {
			if pair.TotalLpAmount != "" && pair.TotalLpAmount != "0" {
				tests = append(tests, []testcase{
					{"found by index", args{int(pair.Index), "9000"}, 200},
				}...)
				break
			}
		}
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetLpValue(s, tt.args.pairIndex, tt.args.lpAmount)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				assert.NotNil(t, result.AssetAId)
				assert.NotNil(t, result.AssetAName)
				assert.NotNil(t, result.AssetAAmount)
				assert.NotNil(t, result.AssetBId)
				assert.NotNil(t, result.AssetBName)
				assert.NotNil(t, result.AssetBAmount)
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func GetLpValue(s *ApiServerSuite, pairIndex int, lpAmount string) (int, *types.LpValue) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/lpValue?pair_index=%d&lp_amount=%s", s.url, pairIndex, lpAmount))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.LpValue{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
