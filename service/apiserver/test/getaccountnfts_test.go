package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/bnb-chain/zkbas/service/apiserver/internal/types"

	"github.com/stretchr/testify/assert"
)

func (s *AppSuite) TestGetAccountNftList() {

	type args struct {
		by     string
		value  string
		offset int
		limit  int
	}
	tests := []struct {
		name     string
		args     args
		httpCode int
	}{
		{"found", args{"account_index", "2", 0, 10}, 200},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetAccountNfts(s, tt.args.by, tt.args.value, tt.args.offset, tt.args.limit)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				if tt.args.offset < int(result.Total) {
					assert.True(t, len(result.Nfts) > 0)
					assert.NotNil(t, result.Nfts[0].Index)
					assert.NotNil(t, result.Nfts[0].ContentHash)
					assert.NotNil(t, result.Nfts[0].OwnerAccountIndex)
					assert.NotNil(t, result.Nfts[0].CollectionId)
					assert.NotNil(t, result.Nfts[0].CreatorTreasuryRate)
				}
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func GetAccountNfts(s *AppSuite, by, value string, offset, limit int) (int, *types.Nfts) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/accountNfts?by=%s&value=%s&offset=%d&limit=%d", s.url, by, value, offset, limit))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.Nfts{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
