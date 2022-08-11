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

func (s *AppSuite) TestGetGasAccount() {

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
			httpCode, result := GetGasAccount(s)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				assert.NotNil(t, result.AccountIndex)
				assert.NotNil(t, result.AccountName)
				assert.NotNil(t, result.AccountStatus)
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func GetGasAccount(s *AppSuite) (int, *types.RespGetGasAccount) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/info/getGasAccount", s.url))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.RespGetGasAccount{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
