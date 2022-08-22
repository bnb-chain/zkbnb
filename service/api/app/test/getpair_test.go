package test

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

func (s *AppSuite) TestGetPair() {
	type args struct {
		pairIndex int
	}
	tests := []struct {
		name     string
		args     args
		httpCode int
	}{
		{"found", args{0}, 200},
		{"not found", args{math.MaxInt}, 400},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetPair(s, tt.args.pairIndex)
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

//goland:noinspection GoUnhandledErrorResult
func GetPair(s *AppSuite, pairIndex int) (int, *types.Pair) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/pair?pair_index=%d", s.url, pairIndex))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.Pair{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
