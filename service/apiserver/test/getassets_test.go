package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
)

func (s *ApiServerSuite) TestGetAssets() {

	type args struct {
		offset int
		limit  int
	}
	tests := []struct {
		name     string
		args     args
		httpCode int
	}{
		{"found", args{0, 10}, 200},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetAssets(s, tt.args.offset, tt.args.limit)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				assert.True(t, len(result.Assets) > 0)
				assert.NotNil(t, result.Assets[0].Name)
				assert.NotNil(t, result.Assets[0].Symbol)
				assert.NotNil(t, result.Assets[0].Address)
				assert.NotNil(t, result.Assets[0].IsGasAsset)
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func GetAssets(s *ApiServerSuite, offset, limit int) (int, *types.Assets) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/assets?offset=%d&limit=%d", s.url, offset, limit))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.Assets{}
	//nolint: errcheck
	json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
