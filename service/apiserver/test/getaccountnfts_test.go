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

func (s *AppSuite) TestGetAccountNftList() {
	type args struct {
		by     string
		value  string
		offset int
		limit  int
	}

	type testcase struct {
		name     string
		args     args
		httpCode int
	}

	tests := []testcase{
		{"not found by index", args{"account_index", "9999999999", 0, 10}, 200},
		{"not found by name", args{"account_name", "notexistname", 0, 10}, 200},
		{"not found by pk", args{"account_pk", "notexistpk", 0, 10}, 200},
		{"invalid by", args{"invalidby", "", 0, 10}, 400},
	}

	statusCode, accounts := GetAccounts(s, 2, 100)
	if statusCode == http.StatusOK && len(accounts.Accounts) > 0 {
		tests = append(tests, []testcase{
			{"found by index", args{"account_index", strconv.Itoa(int(accounts.Accounts[0].Index)), 0, 10}, 200},
			{"found by name", args{"account_name", accounts.Accounts[0].Name, 0, 10}, 200},
			{"found by pk", args{"account_pk", accounts.Accounts[0].Pk, 0, 10}, 200},
		}...)
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetAccountNfts(s, tt.args.by, tt.args.value, tt.args.offset, tt.args.limit)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				if tt.args.offset < int(result.Total) {
					assert.True(t, len(result.Nfts) > 0)
					assert.NotNil(t, result.Nfts[0].Index)
					assert.NotNil(t, result.Nfts[0].ContentHash)
					assert.NotNil(t, result.Nfts[0].OwnerAccountIndex)
					assert.NotNil(t, result.Nfts[0].CollectionId)
					assert.NotNil(t, result.Nfts[0].CreatorTreasuryRate)
				}
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func GetAccountNfts(s *AppSuite, by, value string, offset, limit int) (int, *types.Nfts) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/accountNfts?by=%s&value=%s&offset=%d&limit=%d", s.url, by, value, offset, limit))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.Nfts{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
