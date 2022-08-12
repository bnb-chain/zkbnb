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

func (s *AppSuite) TestGetCurrencyPrices() {

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
			httpCode, result := GetCurrencyPrices(s)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				assert.NotNil(t, result.CurrencyPrices)
				assert.NotNil(t, result.CurrencyPrices[0].Price)
				assert.NotNil(t, result.CurrencyPrices[0].AssetId)
				assert.NotNil(t, result.CurrencyPrices[0].Pair)
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func GetCurrencyPrices(s *AppSuite) (int, *types.RespGetCurrencyPrices) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/info/getCurrencyPrices", s.url))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.RespGetCurrencyPrices{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
