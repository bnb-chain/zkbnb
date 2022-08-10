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

func (s *AppSuite) TestGetmempoolTxsByAccountName() {

	type args struct {
		accountName string
	}
	tests := []struct {
		name     string
		args     args
		httpCode int
	}{
		{"found", args{"sher.legend"}, 200},
		{"not found", args{"notexists.legend"}, 400},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetmempoolTxsByAccountName(s, tt.args.accountName)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				if result.Total > 0 {
					assert.True(t, len(result.MempoolTxs) > 0)
					assert.NotNil(t, result.MempoolTxs[0].BlockHeight)
					assert.NotNil(t, result.MempoolTxs[0].TxHash)
					assert.NotNil(t, result.MempoolTxs[0].TxType)
					assert.NotNil(t, result.MempoolTxs[0].StateRoot)
					assert.NotNil(t, result.MempoolTxs[0].TxInfo)
					assert.NotNil(t, result.MempoolTxs[0].TxStatus)
				}
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func GetmempoolTxsByAccountName(s *AppSuite, accountName string) (int, *types.RespGetmempoolTxsByAccountName) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/tx/getmempoolTxsByAccountName?account_name=%s", s.url, accountName))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.RespGetmempoolTxsByAccountName{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
