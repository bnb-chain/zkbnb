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

func (s *AppSuite) TestGetBlockByCommitment() {

	type args struct {
		blockCommitment string
	}
	tests := []struct {
		name     string
		args     args
		httpCode int
	}{
		{"found", args{"0000000000000000000000000000000000000000000000000000000000000000"}, 200},
		{"not found", args{"notexist"}, 400},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetBlockByCommitment(s, tt.args.blockCommitment)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				assert.NotNil(t, result.Block.BlockHeight)
				assert.NotNil(t, result.Block.BlockCommitment)
				assert.NotNil(t, result.Block.BlockStatus)
				assert.NotNil(t, result.Block.StateRoot)
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func GetBlockByCommitment(s *AppSuite, blockCommitment string) (int, *types.RespGetBlockByCommitment) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/block/getBlockByCommitment?block_commitment=%s", s.url, blockCommitment))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.RespGetBlockByCommitment{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
