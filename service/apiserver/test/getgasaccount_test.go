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

func (s *AppSuite) TestGetGasAccount() {
	tests := []struct {
		name     string
		httpCode int
	}{
		{"found", 200},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetGasAccount(s)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				assert.NotNil(t, result.Index)
				assert.NotNil(t, result.Name)
				assert.NotNil(t, result.Status)
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func GetGasAccount(s *AppSuite) (int, *types.GasAccount) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/gasAccount", s.url))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.GasAccount{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
