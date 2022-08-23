package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/zkbas/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbas/types"
)

func (s *AppSuite) TestSearch() {
	type args struct {
		keyword string
	}
	tests := []struct {
		name     string
		args     args
		httpCode int
		dataType int32
	}{
		{"search block", args{"1"}, 200, types2.TypeBlockHeight},
		{"search account", args{"gas.legend"}, 200, types2.TypeAccountName},
		{"not found", args{"notexist"}, 400, 0},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := Search(s, tt.args.keyword)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				assert.NotNil(t, result.DataType)
				assert.Equal(t, tt.dataType, result.DataType)
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func Search(s *AppSuite, keyword string) (int, *types.Search) {
	resp, err := http.Get(s.url + "/api/v1/search?keyword=" + keyword)
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.Search{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
