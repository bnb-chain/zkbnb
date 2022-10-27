package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
)

func (s *ApiServerSuite) TestGetExecutedTxs() {

	type args struct {
		offset   int
		limit    int
		fromHash string
	}
	tests := []struct {
		name     string
		args     args
		httpCode int
	}{
		{"found", args{0, 10, "hash"}, 200},
		{"found without hash", args{0, 10, ""}, 200},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetExecutedTxs(s, tt.args.offset, tt.args.limit, tt.args.fromHash)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				if tt.args.offset < int(result.Total) {
					assert.True(t, len(result.Txs) > 0)
					assert.NotNil(t, result.Txs[0].BlockHeight)
					assert.NotNil(t, result.Txs[0].Hash)
					assert.NotNil(t, result.Txs[0].Type)
					assert.NotNil(t, result.Txs[0].StateRoot)
					assert.NotNil(t, result.Txs[0].Info)
					assert.NotNil(t, result.Txs[0].Status)
				}
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func GetExecutedTxs(s *ApiServerSuite, offset, limit int, from_hash string) (int, *types.Txs) {
	url := fmt.Sprintf("%s/api/v1/executedTxs?offset=%d&limit=%d", s.url, offset, limit)
	if len(from_hash) > 0 {
		url += fmt.Sprintf("&from_hash=%s", from_hash)
	}
	resp, err := http.Get(url)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.Txs{}
	//nolint: errcheck
	json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
