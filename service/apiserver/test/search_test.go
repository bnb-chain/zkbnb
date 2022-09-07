package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbnb/types"
)

func (s *ApiServerSuite) TestSearch() {
	type testcase struct {
		name     string
		args     string //keyword
		httpCode int
		dataType int32
	}

	tests := []testcase{
		{"not found by account name", "notexist.legend", 400, 0},
		{"not found by account pk", "notexistnotexist", 400, 0},
		{"not found by block height", "9999999", 400, 0},
		{"not found by tx hash", "notexistnotexist", 400, 0},
	}

	statusCode, accounts := GetAccounts(s, 0, 100)
	if statusCode == http.StatusOK && len(accounts.Accounts) > 0 {
		tests = append(tests, []testcase{
			{"found by account name", accounts.Accounts[0].Name, 200, types2.TypeAccountName},
			{"found by account pk", accounts.Accounts[0].Pk, 200, types2.TypeAccountPk},
		}...)
	}

	statusCode, txs := GetTxs(s, 0, 100)
	if statusCode == http.StatusOK && len(txs.Txs) > 0 {
		tests = append(tests, []testcase{
			{"found by block height", strconv.Itoa(int(txs.Txs[0].BlockHeight)), 200, types2.TypeBlockHeight},
			{"found by tx hash", txs.Txs[0].Hash, 200, types2.TypeTxType},
		}...)
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := Search(s, tt.args)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				assert.NotNil(t, result.DataType)
				assert.Equal(t, tt.dataType, result.DataType)
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func Search(s *ApiServerSuite, keyword string) (int, *types.Search) {
	resp, err := http.Get(s.url + "/api/v1/search?keyword=" + keyword)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.Search{}
	//nolint: errcheck
	json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
