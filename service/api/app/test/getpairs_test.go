package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

func (s *AppSuite) TestGetPairs() {

	type args struct {
	}
	tests := []struct {
		name     string
		args     args
		httpCode int
	}{
		{"found", args{}, 200},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetPairs(s)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				assert.NotNil(t, result.Pairs)
				assert.NotNil(t, result.Pairs[0].PairIndex)
				assert.NotNil(t, result.Pairs[0].AssetAName)
				assert.NotNil(t, result.Pairs[0].AssetBName)
				assert.NotNil(t, result.Pairs[0].AssetAId)
				assert.NotNil(t, result.Pairs[0].AssetBId)
				assert.NotNil(t, result.Pairs[0].AssetAAmount)
				assert.NotNil(t, result.Pairs[0].AssetBAmount)
				assert.NotNil(t, result.Pairs[0].FeeRate)
				assert.NotNil(t, result.Pairs[0].TreasuryRate)
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func GetPairs(s *AppSuite) (int, *types.RespGetPairs) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/pair/getAvailablePairs", s.url))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.RespGetPairs{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
