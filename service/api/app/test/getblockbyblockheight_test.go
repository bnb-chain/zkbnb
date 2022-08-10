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

func (s *AppSuite) TestGetBlockByBlockHeight() {

	type args struct {
		blockHeight int
	}
	tests := []struct {
		name     string
		args     args
		httpCode int
	}{
		{"found", args{1}, 200},
		{"not found", args{math.MaxInt}, 400},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetBlockByBlockHeight(s, tt.args.blockHeight)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				assert.NotNil(t, result.Block.BlockHeight)
				assert.NotNil(t, result.Block.BlockCommitment)
				assert.NotNil(t, result.Block.BlockStatus)
				assert.NotNil(t, result.Block.StateRoot)
				assert.NotNil(t, result.Block.Txs)
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func GetBlockByBlockHeight(s *AppSuite, blockHeight int) (int, *types.RespGetBlockByBlockHeight) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/block/getBlockByBlockHeight?block_height=%d", s.url, blockHeight))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.RespGetBlockByBlockHeight{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
