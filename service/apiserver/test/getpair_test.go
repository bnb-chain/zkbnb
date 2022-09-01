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

func (s *ApiServerSuite) TestGetPair() {
	type testcase struct {
		name     string
		args     int //pair index
		httpCode int
	}

	tests := []testcase{
		{"not found", math.MaxInt, 400},
	}

	statusCode, pairs := GetPairs(s, 0, 100)
	if statusCode == http.StatusOK && len(pairs.Pairs) > 0 {
		tests = append(tests, []testcase{
			{"found by index", int(pairs.Pairs[0].Index), 200},
		}...)
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetPair(s, tt.args)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				assert.NotNil(t, result.AssetAId)
				assert.NotNil(t, result.AssetBId)
				assert.NotNil(t, result.AssetAAmount)
				assert.NotNil(t, result.AssetBAmount)
				assert.NotNil(t, result.TotalLpAmount)
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func GetPair(s *ApiServerSuite, pairIndex int) (int, *types.Pair) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/pair?index=%d", s.url, pairIndex))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.Pair{}
	//nolint: errcheck
	json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
