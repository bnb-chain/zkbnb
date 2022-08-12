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

func (s *AppSuite) TestGetGasFeeAssetList() {

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
			httpCode, result := GetGasFeeAssetList(s)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				assert.NotNil(t, result.Assets)
				assert.NotNil(t, result.Assets[0].AssetId)
				assert.NotNil(t, result.Assets[0].AssetSymbol)
				assert.NotNil(t, result.Assets[0].AssetName)
				assert.NotNil(t, result.Assets[0].AssetAddress)
				assert.NotNil(t, result.Assets[0].IsGasAsset)
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func GetGasFeeAssetList(s *AppSuite) (int, *types.RespGetGasFeeAssetList) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/info/getGasFeeAssetList", s.url))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.RespGetGasFeeAssetList{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
