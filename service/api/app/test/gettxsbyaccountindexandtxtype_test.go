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

func (s *AppSuite) TestGetTxsByAccountIndexAndTxType() {

	type args struct {
		accountIndex int
		txType       int
		offset       int
		limit        int
	}
	tests := []struct {
		name     string
		args     args
		httpCode int
	}{
		{"found", args{1, 2, 0, 10}, 200},
		{"not found", args{1, 2, math.MaxInt, 10}, 400},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetTxsByAccountIndexAndTxType(s, tt.args.accountIndex, tt.args.txType, tt.args.offset, tt.args.limit)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				if tt.args.offset < int(result.Total) {
					assert.True(t, len(result.Txs) > 0)
					assert.NotNil(t, result.Txs[0].BlockHeight)
					assert.NotNil(t, result.Txs[0].TxHash)
					assert.NotNil(t, result.Txs[0].TxType)
					assert.NotNil(t, result.Txs[0].StateRoot)
					assert.NotNil(t, result.Txs[0].TxInfo)
					assert.NotNil(t, result.Txs[0].TxStatus)
				}
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func GetTxsByAccountIndexAndTxType(s *AppSuite, accountIndex, txType, offset, limit int) (int, *types.RespGetTxsByAccountIndexAndTxType) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/tx/getTxsByAccountIndexAndTxType?account_index=%d&tx_type=%d&offset=%d&limit=%d", s.url, accountIndex, txType, offset, limit))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.RespGetTxsByAccountIndexAndTxType{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
