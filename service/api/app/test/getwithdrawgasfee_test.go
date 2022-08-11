package test

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

func (s *AppSuite) TestGetWithdrawGasFee() {

	type args struct {
		assetId int
	}
	tests := []struct {
		name     string
		args     args
		httpCode int
	}{
		{"found", args{0}, 200},
		{"not found", args{math.MaxInt}, 400},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetWithdrawGasFee(s, tt.args.assetId)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				assert.NotNil(t, result.GasFee)
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func GetWithdrawGasFee(s *AppSuite, assetId int) (int, *types.RespGetWithdrawGasFee) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/info/getWithdrawGasFee?asset_id=%d", s.url, assetId))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.RespGetWithdrawGasFee{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
