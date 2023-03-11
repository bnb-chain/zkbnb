package chain

import (
	"encoding/json"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/ethereum/go-ethereum/common"
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
		PublicKey:       formatAccountInfo.PublicKey,
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
		PublicKey:       accountInfo.PublicKey,
		L1Address:       accountInfo.L1Address,
		Nonce:           accountInfo.Nonce,
		CollectionNonce: accountInfo.CollectionNonce,
		AssetInfo:       assetInfo,
		AssetRoot:       accountInfo.AssetRoot,
		Status:          accountInfo.Status,
	}
	return formatAccountInfo, nil
}

func EmptyAccount(accountIndex int64, l1Address string, nilAccountAssetRoot []byte) (info *account.Account) {
	return &account.Account{
		AccountIndex:    accountIndex,
		PublicKey:       common2.EmptyPubKey(),
		L1Address:       l1Address,
		Nonce:           types.EmptyNonce,
		CollectionNonce: types.EmptyCollectionNonce,
		AssetInfo:       types.EmptyAccountAssetInfo,
		AssetRoot:       common.Bytes2Hex(nilAccountAssetRoot),
		Status:          account.AccountStatusPending,
	}
}
