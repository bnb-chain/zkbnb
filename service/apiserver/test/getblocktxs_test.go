package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/zkbas/service/apiserver/internal/types"
)

func (s *ApiServerSuite) TestGetBlockTxs() {
	type args struct {
		by    string
		value string
	}

	type testcase struct {
		name     string
		args     args
		httpCode int
	}

	tests := []testcase{
		{"not found by block_height", args{"block_height", "99999999"}, 200},
		{"not found by block_commitment", args{"block_commitment", "fsfsfsfsf100"}, 200},
		{"invalidby", args{"invalidby", ""}, 400},
	}

	statusCode, blocks := GetBlocks(s, 0, 100)
	if statusCode == http.StatusOK && len(blocks.Blocks) > 0 {
		tests = append(tests, []testcase{
			{"found by block_height", args{"block_height", strconv.Itoa(int(blocks.Blocks[len(blocks.Blocks)-1].Height))}, 200},
			{"found by block_commitment", args{"block_commitment", blocks.Blocks[len(blocks.Blocks)-1].Commitment}, 200},
		}...)
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetBlockTxs(s, tt.args.by, tt.args.value)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				if len(result.Txs) > 0 {
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

func GetBlockTxs(s *ApiServerSuite, by, value string) (int, *types.Txs) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/blockTxs?by=%s&value=%s", s.url, by, value))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.Txs{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
