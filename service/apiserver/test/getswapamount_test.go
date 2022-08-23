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

func (s *AppSuite) TestGetSwapAmount() {
	type args struct {
		pairIndex   int
		assetId     uint32
		assetAmount string
		isFrom      bool
	}
	tests := []struct {
		name     string
		args     args
		httpCode int
	}{
		{"found", args{1, 0, "1", true}, 200},
		{"not found", args{math.MaxUint32, math.MaxUint32, "1", true}, 400},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetSwapAmount(s, tt.args.pairIndex)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				assert.NotNil(t, result.AssetId)
				assert.NotNil(t, result.AssetAmount)
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func GetSwapAmount(s *AppSuite, pairIndex int) (int, *types.SwapAmount) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/swapAmount?pair_index=%d", s.url, pairIndex))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.SwapAmount{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
