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

func (s *AppSuite) TestGetTxs() {

	type args struct {
		offset int
		limit  int
	}
	tests := []struct {
		name     string
		args     args
		httpCode int
	}{
		{"found", args{0, 10}, 200},
		{"not found", args{math.MaxInt, 10}, 400},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetTxs(s, tt.args.offset, tt.args.limit)
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

func GetTxs(s *AppSuite, offset, limit int) (int, *types.RespGetTxs) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/tx/getTxs?offset=%d&limit=%d", s.url, offset, limit))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.RespGetTxs{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}