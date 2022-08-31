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

func (s *ApiServerSuite) TestGetAccountMempoolTxs() {
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
		{"not found by index", args{"account_index", "9999999"}, 200},
		{"not found by name", args{"account_name", "notexists.legend"}, 200},
		{"not found by pk", args{"account_pk", "notexists"}, 200},
		{"invalidby", args{"invalidby", ""}, 400},
	}

	statusCode, txs := GetMempoolTxs(s, 0, 100)
	if statusCode == http.StatusOK && len(txs.MempoolTxs) > 0 {
		_, account := GetAccount(s, "name", txs.MempoolTxs[len(txs.MempoolTxs)-1].AccountName)
		tests = append(tests, []testcase{
			{"found by index", args{"account_index", strconv.Itoa(int(account.Index))}, 200},
			{"found by name", args{"account_name", account.Name}, 200},
			{"found by pk", args{"account_pk", account.Pk}, 200},
		}...)
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetAccountMempoolTxs(s, tt.args.by, tt.args.value)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				if result.Total > 0 {
					assert.True(t, len(result.MempoolTxs) > 0)
					assert.NotNil(t, result.MempoolTxs[0].BlockHeight)
					assert.NotNil(t, result.MempoolTxs[0].Hash)
					assert.NotNil(t, result.MempoolTxs[0].Type)
					assert.NotNil(t, result.MempoolTxs[0].StateRoot)
					assert.NotNil(t, result.MempoolTxs[0].Info)
					assert.NotNil(t, result.MempoolTxs[0].Status)
				}
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func GetAccountMempoolTxs(s *ApiServerSuite, by, value string) (int, *types.MempoolTxs) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/accountMempoolTxs?by=%s&value=%s", s.url, by, value))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.MempoolTxs{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
