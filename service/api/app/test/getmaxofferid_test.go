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

func (s *AppSuite) TestGetMaxOfferId() {
	type args struct {
		accountIndex int
	}
	tests := []struct {
		name     string
		args     args
		httpCode int
	}{
		{"found", args{2}, 200},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetMaxOfferId(s, tt.args.accountIndex)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				assert.True(t, result.OfferId >= 0)
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

//goland:noinspection GoUnhandledErrorResult
func GetMaxOfferId(s *AppSuite, accountIndex int) (int, *types.MaxOfferId) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/maxOfferId?account_index=%d", s.url, accountIndex))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.MaxOfferId{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
