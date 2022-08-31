package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/zkbas/service/apiserver/internal/types"
)

func (s *ApiServerSuite) TestGetAccount() {
	type args struct {
		by    string
		value string
	}

	type testcase struct {
		name     string
		args     args
		httpCode int
	}

	tests := []testcase{
		{"not found by index", args{"index", "9999999999"}, 400},
		{"not found by name", args{"name", "not exist name"}, 400},
		{"not found by pk", args{"pk", "not exist pk"}, 400},
		{"invalid by", args{"invalidby", ""}, 400},
	}

	statusCode, accounts := GetAccounts(s, 0, 100)
	if statusCode == http.StatusOK && len(accounts.Accounts) > 0 {
		tests = append(tests, []testcase{
			{"found by index", args{"index", strconv.Itoa(int(accounts.Accounts[0].Index))}, 200},
			{"found by name", args{"name", accounts.Accounts[0].Name}, 200},
			{"found by pk", args{"pk", accounts.Accounts[0].Pk}, 200},
		}...)
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetAccount(s, tt.args.by, tt.args.value)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				assert.NotNil(t, result.Pk)
				assert.NotNil(t, result.Name)
				assert.True(t, result.Nonce >= 0)
				assert.NotNil(t, result.Assets)
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func GetAccount(s *ApiServerSuite, by, value string) (int, *types.Account) {
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
