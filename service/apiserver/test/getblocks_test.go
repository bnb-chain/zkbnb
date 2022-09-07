package test

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
)

func (s *ApiServerSuite) TestGetBlocks() {

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
			httpCode, result := GetBlocks(s, tt.args.offset, tt.args.limit)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				if tt.args.offset < int(result.Total) {
					assert.True(t, len(result.Blocks) > 0)
					assert.NotNil(t, result.Blocks[0].Height)
					assert.NotNil(t, result.Blocks[0].Commitment)
					assert.NotNil(t, result.Blocks[0].Status)
					assert.NotNil(t, result.Blocks[0].StateRoot)
					//assert.NotNil(t, result.Blocks[0].Txs)
				}
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func GetBlocks(s *ApiServerSuite, offset, limit int) (int, *types.Blocks) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/blocks?offset=%d&limit=%d", s.url, offset, limit))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.Blocks{}
	//nolint:errcheck
	json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
