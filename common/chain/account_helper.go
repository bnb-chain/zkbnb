package chain

import (
	"encoding/json"

	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/dao/account"
	"github.com/bnb-chain/zkbnb/types"
)

func FromFormatAccountInfo(formatAccountInfo *types.AccountInfo) (accountInfo *account.Account, err error) {
	assetInfoBytes, err := json.Marshal(formatAccountInfo.AssetInfo)
	if err != nil {
		return nil, types.JsonErrMarshal
	}
	accountInfo = &account.Account{
		Model: gorm.Model{
			ID: formatAccountInfo.AccountId,
		},
		AccountIndex:    formatAccountInfo.AccountIndex,
		AccountName:     formatAccountInfo.AccountName,
		PublicKey:       formatAccountInfo.PublicKey,
		AccountNameHash: formatAccountInfo.AccountNameHash,
		L1Address:       formatAccountInfo.L1Address,
		Nonce:           formatAccountInfo.Nonce,
		CollectionNonce: formatAccountInfo.CollectionNonce,
		AssetInfo:       string(assetInfoBytes),
		AssetRoot:       formatAccountInfo.AssetRoot,
		Status:          formatAccountInfo.Status,
	}
	return accountInfo, nil
}

func ToFormatAccountInfo(accountInfo *account.Account) (formatAccountInfo *types.AccountInfo, err error) {
	var assetInfo map[int64]*types.AccountAsset
	err = json.Unmarshal([]byte(accountInfo.AssetInfo), &assetInfo)
	if err != nil {
		return nil, types.JsonErrUnmarshal
	}
	formatAccountInfo = &types.AccountInfo{
		AccountId:       accountInfo.ID,
		AccountIndex:    accountInfo.AccountIndex,
		AccountName:     accountInfo.AccountName,
		PublicKey:       accountInfo.PublicKey,
		AccountNameHash: accountInfo.AccountNameHash,
		L1Address:       accountInfo.L1Address,
		Nonce:           accountInfo.Nonce,
		CollectionNonce: accountInfo.CollectionNonce,
		AssetInfo:       assetInfo,
		AssetRoot:       accountInfo.AssetRoot,
		Status:          accountInfo.Status,
	}
	return formatAccountInfo, nil
}
