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

func (s *AppSuite) TestGetAccountMempoolTxs() {

	type args struct {
		by    string
		value string
	}
	tests := []struct {
		name     string
		args     args
		httpCode int
	}{
		{"found", args{"account_name", "sher.legend"}, 200},
		{"not found", args{"account_pk", "notexists.legend"}, 400},
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

func GetAccountMempoolTxs(s *AppSuite, by, value string) (int, *types.MempoolTxs) {
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
