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

func (s *AppSuite) TestGetAccount() {
	type args struct {
		by    string
		value string
	}
	tests := []struct {
		name     string
		args     args
		httpCode int
	}{
		{"found", args{"index", "0"}, 200},
		{"found", args{"name", "gas.legend"}, 200},
		{"found", args{"pk", "fcb8470d33c59a5cbf5e10df426eb97c2773ab890c3364f4162ba782a56ca998"}, 200},
		{"not found", args{"pk", "not exist pk"}, 400},
		{"not found", args{"invalidby", ""}, 400},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetAccount(s, tt.args.by, tt.args.value)
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

func GetAccount(s *AppSuite, by, value string) (int, *types.Account) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/account?by=%s&value=%s", s.url, by, value))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.Account{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
