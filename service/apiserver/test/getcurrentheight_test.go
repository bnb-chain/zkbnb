package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/zkbas/service/apiserver/internal/types"
)

func (s *ApiServerSuite) TestGetCurrentHeight() {
	tests := []struct {
		name     string
		httpCode int
	}{
		{"found", 200},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetCurrentHeight(s)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				assert.NotNil(t, result.Height)
				assert.True(t, result.Height > 0)
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func GetCurrentHeight(s *ApiServerSuite) (int, *types.CurrentHeight) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/currentHeight", s.url))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.CurrentHeight{}
	//nolint: errcheck
	json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
