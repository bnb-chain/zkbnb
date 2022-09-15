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

func (s *ApiServerSuite) TestGetAsset() {
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
		{"not found by id", args{"id", "9999999999"}, 400},
		{"not found by symbol", args{"symbol", "notexist"}, 400},
		{"invalid by", args{"invalidby", ""}, 400},
		{"found by id", args{"id", "0"}, 200},
		{"found by symbol", args{"symbol", "BNB"}, 200},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetAsset(s, tt.args.by, tt.args.value)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				assert.NotNil(t, result.Id)
				assert.NotNil(t, result.Name)
				assert.NotNil(t, result.Symbol)
				assert.NotNil(t, result.Price)
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func GetAsset(s *ApiServerSuite, by, value string) (int, *types.Asset) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/asset?by=%s&value=%s", s.url, by, value))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	fmt.Println(body)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.Asset{}
	//nolint: errcheck
	json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
