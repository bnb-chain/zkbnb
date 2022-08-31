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

func (s *AppSuite) TestGetTx() {
	type testcase struct {
		name     string
		args     string //tx hash
		httpCode int
	}

	tests := []testcase{
		{"not found", "notexistshash", 400},
	}

	statusCode, txs := GetTxs(s, 0, 100)

	if statusCode == http.StatusOK && len(txs.Txs) > 0 {
		tests = append(tests, []testcase{
			{"found", txs.Txs[0].Hash, 200},
		}...)
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetTx(s, tt.args)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				assert.NotNil(t, result.Tx.BlockHeight)
				assert.NotNil(t, result.Tx.Hash)
				assert.NotNil(t, result.Tx.Type)
				assert.NotNil(t, result.Tx.StateRoot)
				assert.NotNil(t, result.Tx.Info)
				assert.NotNil(t, result.Tx.Status)
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func GetTx(s *AppSuite, hash string) (int, *types.EnrichedTx) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/tx?hash=%s", s.url, hash))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.EnrichedTx{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
