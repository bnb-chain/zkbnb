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

func (s *AppSuite) TestGetBlock() {

	type args struct {
		by    string
		value string
	}
	tests := []struct {
		name     string
		args     args
		httpCode int
	}{
		{"found", args{"height", "1"}, 200},
		{"found", args{"commitment", "0000000000000000000000000000000000000000000000000000000000000000"}, 200},
		{"not found", args{"invalidby", ""}, 400},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetBlock(s, tt.args.by, tt.args.value)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				assert.NotNil(t, result.Height)
				assert.NotNil(t, result.Commitment)
				assert.NotNil(t, result.Status)
				assert.NotNil(t, result.StateRoot)
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

//goland:noinspection ALL
func GetBlock(s *AppSuite, by, value string) (int, *types.Block) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/block?by=%s&value=%s", s.url, by, value))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.Block{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
