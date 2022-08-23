package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/bnb-chain/zkbas/service/apiserver/internal/types"

	"github.com/stretchr/testify/assert"
)

func (s *AppSuite) TestGetGasFeeAssets() {

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
			httpCode, result := GetGasFeeAssets(s)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				assert.NotNil(t, result.Assets)
				assert.NotNil(t, result.Assets[0].Id)
				assert.NotNil(t, result.Assets[0].Symbol)
				assert.NotNil(t, result.Assets[0].Name)
				assert.NotNil(t, result.Assets[0].Address)
				assert.NotNil(t, result.Assets[0].IsGasAsset)
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func GetGasFeeAssets(s *AppSuite) (int, *types.GasFeeAssets) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/gasFeeAssets", s.url))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.GasFeeAssets{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
