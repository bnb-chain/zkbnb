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

func (s *AppSuite) TestGetLPValue() {

	type args struct {
		pairIndex int
		lpAmount  string
	}
	tests := []struct {
		name     string
		args     args
		httpCode int
	}{
		{"found", args{0, "1"}, 200},
		{"not found", args{math.MaxInt, "2"}, 400},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetLPValue(s, tt.args.pairIndex, tt.args.lpAmount)
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

func GetLPValue(s *AppSuite, pairIndex int, lpAmount string) (int, *types.RespGetLPValue) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/pair/getLPValue?pair_index=%d&lp_amount=%s", s.url, pairIndex, lpAmount))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.RespGetLPValue{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
