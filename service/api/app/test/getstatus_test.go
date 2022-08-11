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

func (s *AppSuite) TestGetStatus() {

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

func GetStatus(s *AppSuite) (int, *types.RespGetStatus) {
	resp, err := http.Get(fmt.Sprintf("%s/", s.url))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.RespGetStatus{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}