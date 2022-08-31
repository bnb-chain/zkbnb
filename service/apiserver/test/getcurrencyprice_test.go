package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/zkbas/service/apiserver/internal/types"
)

func (s *AppSuite) TestGetCurrencyPrice() {
	type args struct {
		by    string
		value string
	}
	tests := []struct {
		name     string
		args     args
		httpCode int
	}{
		{"found", args{"symbol", "BNB"}, 200},
		{"not found", args{"symbol", "notfound"}, 400},
		{"invalidby", args{"invalidby", ""}, 400},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetCurrencyPrice(s, tt.args.by, tt.args.value)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				assert.NotNil(t, result.AssetId)
				assert.NotNil(t, result.Price)
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func GetCurrencyPrice(s *AppSuite, by, value string) (int, *types.CurrencyPrice) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/currencyPrice?by=%s&value=%s", s.url, by, value))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.CurrencyPrice{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
