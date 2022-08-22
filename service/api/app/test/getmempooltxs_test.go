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

func (s *AppSuite) TestGetMempoolTxs() {

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
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetMempoolTxs(s, tt.args.offset, tt.args.limit)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				if tt.args.offset < int(result.Total) {
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

//goland:noinspection GoUnhandledErrorResult
func GetMempoolTxs(s *AppSuite, offset, limit int) (int, *types.MempoolTxs) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/mempoolTxs?offset=%d&limit=%d", s.url, offset, limit))
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
