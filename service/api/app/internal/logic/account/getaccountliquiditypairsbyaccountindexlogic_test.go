package account

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	table "github.com/zecrey-labs/zecrey-legend/common/model/liquidity"
	"github.com/zecrey-labs/zecrey-legend/pkg/zerror"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/liquidity"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"
)

func TestGetAccountLiquidityPairsByAccountIndexLogic_GetAccountLiquidityPairsByAccountIndex(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLiquidity := liquidity.NewMockLiquidityModel(ctrl)

	l := &GetAccountLiquidityPairsByAccountIndexLogic{
		liquidity: mockLiquidity,
	}

	// error case
	mockLiquidity.EXPECT().GetLiquidityByPairIndex(gomock.Any()).Return(nil, zerror.New(-1, "error")).MaxTimes(1)
	req := &types.ReqGetAccountLiquidityPairsByAccountIndex{AccountIndex: 1}
	_, err := l.GetAccountLiquidityPairsByAccountIndex(req)
	assert.NotNil(t, err)

	// normal case
	mockLiquidity.EXPECT().GetLiquidityByPairIndex(gomock.Any()).Return(&table.Liquidity{}, nil).AnyTimes()
	req = &types.ReqGetAccountLiquidityPairsByAccountIndex{AccountIndex: 1}
	_, err = l.GetAccountLiquidityPairsByAccountIndex(req)
	assert.Nil(t, err)
}
