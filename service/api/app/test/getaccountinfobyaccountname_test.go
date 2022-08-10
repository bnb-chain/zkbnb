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

func (s *AppSuite) TestGetAccountInfoByAccountName() {
	type args struct {
		accountName string
	}
	tests := []struct {
		name     string
		args     args
		httpCode int
	}{
		{"found", args{"gas.legend"}, 200},
		{"not found", args{"notfound.legend"}, 400},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetAccountInfoByAccountName(s, tt.args.accountName)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				assert.NotNil(t, result.AccountPk)
				assert.NotNil(t, result.AccountIndex)
				assert.True(t, result.Nonce >= 0)
				assert.NotNil(t, result.Assets)
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func GetAccountInfoByAccountName(s *AppSuite, accountName string) (int, *types.RespGetAccountInfoByAccountName) {
	resp, err := http.Get(s.url + "/api/v1/account/getAccountInfoByAccountName?account_name=" + accountName)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.RespGetAccountInfoByAccountName{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
