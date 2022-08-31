package test

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/zkbas/service/apiserver/internal/types"
)

func (s *ApiServerSuite) TestGetAccounts() {

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
			httpCode, result := GetAccounts(s, tt.args.offset, tt.args.limit)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				if tt.args.offset < int(result.Total) {
					assert.True(t, len(result.Accounts) > 0)
					assert.NotNil(t, result.Accounts[0].Name)
					assert.NotNil(t, result.Accounts[0].Index)
					assert.NotNil(t, result.Accounts[0].Pk)
				}
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func GetAccounts(s *ApiServerSuite, offset, limit int) (int, *types.Accounts) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/accounts?offset=%d&limit=%d", s.url, offset, limit))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.Accounts{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
