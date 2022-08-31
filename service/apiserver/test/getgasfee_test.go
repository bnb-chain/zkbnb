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

func (s *AppSuite) TestGetGasFee() {
	type testcase struct {
		name     string
		args     int //asset id
		httpCode int
	}

	tests := []testcase{
		{"not found", math.MaxInt, 400},
	}

	statusCode, assets := GetGasFeeAssets(s)
	if statusCode == http.StatusOK && len(assets.Assets) > 0 {
		tests = append(tests, []testcase{
			{"found by index", int(assets.Assets[0].Id), 200},
		}...)
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetGasFee(s, tt.args)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				assert.NotNil(t, result.GasFee)
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func GetGasFee(s *AppSuite, assetId int) (int, *types.GasFee) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/gasFee?asset_id=%d", s.url, assetId))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.GasFee{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
