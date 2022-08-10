package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/bnb-chain/zkbas/service/api/app/internal/types"

	"github.com/stretchr/testify/assert"
)

func (s *AppSuite) TestGetCurrencyPriceBySymbol() {
	type args struct {
		symbol string
	}
	tests := []struct {
		name     string
		args     args
		httpCode int
	}{
		{"found", args{"BNB"}, 200},
		{"not found", args{"notfound.legend"}, 400},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetCurrencyPriceBySymbol(s, tt.args.symbol)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				assert.NotNil(t, result.AssetId)
				assert.NotNil(t, result.Price)
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func GetCurrencyPriceBySymbol(s *AppSuite, symbol string) (int, *types.RespGetCurrencyPriceBySymbol) {
	resp, err := http.Get(s.url + "/api/v1/info/getCurrencyPriceBySymbol?symbol=" + symbol)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.RespGetCurrencyPriceBySymbol{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
