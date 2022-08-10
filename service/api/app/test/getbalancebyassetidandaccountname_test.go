package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/bnb-chain/zkbas/service/api/app/internal/types"

	"github.com/stretchr/testify/assert"
)

func (s *AppSuite) TestGetBalanceByAssetIdAndAccountName() {
	type args struct {
		accountName string
		assetId     int
	}
	tests := []struct {
		name     string
		args     args
		httpCode int
	}{
		{"found", args{"sher.legend", 0}, 200},
		{"not found", args{"notfound.legend", 0}, 400},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetAccountInfoByAccountName(s, tt.args.accountName)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				assert.NotNil(t, result.AccountPk)
				assert.NotNil(t, result.AccountIndex)
				assert.True(t, result.Nonce >= 0)
				assert.NotNil(t, result.Assets)
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func GetBalanceByAssetIdAndAccountName(s *AppSuite, accountName string, assetId int) (int, *types.RespGetBlanceInfoByAssetIdAndAccountName) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/account/getBalanceByAssetIdAndAccountName?account_name=%s&asset_id=%d", s.url, accountName, assetId))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.RespGetBlanceInfoByAssetIdAndAccountName{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
