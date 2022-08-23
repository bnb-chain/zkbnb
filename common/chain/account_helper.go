package chain

import (
	"encoding/json"

	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/dao/account"
	"github.com/bnb-chain/zkbas/types"
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

type FormatAccountHistoryInfo struct {
	AccountId       uint
	AccountIndex    int64
	Nonce           int64
	CollectionNonce int64
	// map[int64]*AccountAsset
	AssetInfo map[int64]*types.AccountAsset
	AssetRoot string
	// map[int64]*Liquidity
	L2BlockHeight int64
	Status        int
}
