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

func (s *AppSuite) TestGetAccountStatusByAccountPk() {
	type args struct {
		pubKey string
	}
	tests := []struct {
		name     string
		args     args
		httpCode int
	}{
		{"found", args{"fcb8470d33c59a5cbf5e10df426eb97c2773ab890c3364f4162ba782a56ca998"}, 200},
		{"not found", args{"notvalidpubkey"}, 400},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetAccountStatusByAccountPk(s, tt.args.pubKey)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				assert.NotNil(t, result.AccountName)
				assert.NotNil(t, result.AccountIndex)
				assert.NotNil(t, result.AccountStatus)
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func GetAccountStatusByAccountPk(s *AppSuite, accountPk string) (int, *types.RespGetAccountStatusByAccountPk) {
	resp, err := http.Get(s.url + "/api/v1/account/getAccountStatusByAccountPk?account_pk=" + accountPk)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.RespGetAccountStatusByAccountPk{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
