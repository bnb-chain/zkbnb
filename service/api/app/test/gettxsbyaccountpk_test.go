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

func (s *AppSuite) TestGetTxsByAccountPk() {

	type args struct {
		accountPk string
		offset    int
		limit     int
	}
	tests := []struct {
		name     string
		args     args
		httpCode int
	}{
		{"found", args{"fcb8470d33c59a5cbf5e10df426eb97c2773ab890c3364f4162ba782a56ca998", 0, 10}, 200},
		{"not found", args{"notexistpk", math.MaxInt, 10}, 400},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetTxsByAccountPk(s, tt.args.accountPk, tt.args.offset, tt.args.limit)
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

func GetTxsByAccountPk(s *AppSuite, accountPk string, offset, limit int) (int, *types.RespGetTxsByAccountPk) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/tx/getTxsByAccountPk?account_pk=%s&offset=%d&limit=%d", s.url, accountPk, offset, limit))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.RespGetTxsByAccountPk{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
