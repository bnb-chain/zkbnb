package test

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

func (s *AppSuite) TestGetAccountByIndex() {
	type args struct {
		accountIndex int
	}
	tests := []struct {
		name     string
		args     args
		httpCode int
	}{
		{"found", args{0}, 200},
		{"not found", args{math.MaxInt}, 400},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetAccountByIndex(s, tt.args.accountIndex)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				assert.NotNil(t, result.AccountPk)
				assert.NotNil(t, result.AccountName)
				assert.True(t, result.Nonce >= 0)
				assert.NotNil(t, result.Assets)
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func GetAccountByIndex(s *AppSuite, accountIndex int) (int, *types.RespGetAccountByIndex) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/account/getAccountByIndex?account_index=%d", s.url, accountIndex))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.RespGetAccountByIndex{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
