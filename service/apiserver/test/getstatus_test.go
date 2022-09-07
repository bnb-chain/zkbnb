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

func (s *ApiServerSuite) TestGetStatus() {

	type args struct {
	}
	tests := []struct {
		name     string
		args     args
		httpCode int
	}{
		{"found", args{}, 200},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetStatus(s)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				assert.NotNil(t, result.NetworkId)
				assert.NotNil(t, result.Status)
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func GetStatus(s *ApiServerSuite) (int, *types.Status) {
	resp, err := http.Get(fmt.Sprintf("%s/", s.url))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.Status{}
	//nolint: errcheck
	json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
