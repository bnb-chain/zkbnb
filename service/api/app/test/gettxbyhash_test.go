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

func (s *AppSuite) TestGetTx() {

	type args struct {
		txHash string
	}
	tests := []struct {
		name     string
		args     args
		httpCode int
	}{
		{"found", args{"Ck9IUvIdohSLz1grisjStIQEsfjoNGebFLU4KO1BQIk="}, 200},
		{"not found", args{"notexists"}, 400},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			httpCode, result := GetTx(s, tt.args.txHash)
			assert.Equal(t, tt.httpCode, httpCode)
			if httpCode == http.StatusOK {
				assert.NotNil(t, result.Tx.BlockHeight)
				assert.NotNil(t, result.Tx.Hash)
				assert.NotNil(t, result.Tx.Type)
				assert.NotNil(t, result.Tx.StateRoot)
				assert.NotNil(t, result.Tx.Info)
				assert.NotNil(t, result.Tx.Status)
				fmt.Printf("result: %+v \n", result)
			}
		})
	}

}

func GetTx(s *AppSuite, txHash string) (int, *types.EnrichedTx) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/tx?tx_hash=%s", s.url, txHash))
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.EnrichedTx{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}
