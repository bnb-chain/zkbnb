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

func (s *AppSuite) TestGetLayer2BasicInfo() {

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
			httpCode, result := GetLayer2BasicInfo(s)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				assert.NotNil(t, result.TotalTransactionCount)
				assert.NotNil(t, result.BlockCommitted)
				assert.NotNil(t, result.BlockVerified)
				assert.NotNil(t, result.ContractAddresses[0])
				assert.NotNil(t, result.ContractAddresses[1])
				assert.NotNil(t, result.YesterdayTransactionCount)
				assert.NotNil(t, result.TodayActiveUserCount)
				assert.NotNil(t, result.YesterdayActiveUserCount)
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func GetLayer2BasicInfo(s *AppSuite) (int, *types.Layer2BasicInfo) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/layer2BasicInfo", s.url))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.Layer2BasicInfo{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
