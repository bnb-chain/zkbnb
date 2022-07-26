package account

import (
	"testing"

	table "github.com/bnb-chain/zkbas/common/model/account"
	"github.com/bnb-chain/zkbas/pkg/zerror"
	"github.com/bnb-chain/zkbas/service/api/app/internal/repo/account"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetAccountStatusByAccountNameLogic_GetAccountStatusByAccountName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockAccount := account.NewMockModel(ctrl)
	l := &GetAccountStatusByAccountNameLogic{
		account: mockAccount,
	}
	// error case
	mockAccount.EXPECT().GetBasicAccountByAccountName(gomock.Any(), gomock.Any()).Return(nil, zerror.New(-1, "error")).MaxTimes(1)
	req := &types.ReqGetAccountStatusByAccountName{AccountName: ""}
	_, err := l.GetAccountStatusByAccountName(req)
	assert.NotNil(t, err)

	// normal case
	mockAccount.EXPECT().GetBasicAccountByAccountName(gomock.Any(), gomock.Any()).Return(&table.Account{}, nil).AnyTimes()
	req = &types.ReqGetAccountStatusByAccountName{AccountName: ""}
	_, err = l.GetAccountStatusByAccountName(req)
	assert.Nil(t, err)
}
