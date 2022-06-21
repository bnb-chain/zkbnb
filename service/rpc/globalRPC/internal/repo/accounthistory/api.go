package accounthistory

import (
	table "github.com/bnb-chain/zkbas/common/model/account"
)

type AccountHistory interface {
	GetAccountByAccountName(accountName string) (account *table.AccountHistory, err error)
	GetAccountByAccountIndex(accountIndex int64) (account *table.AccountHistory, err error)
	GetAccountByPk(pk string) (account *table.AccountHistory, err error)
	GetAccountAssetsByIndex(accountIndex int64) (accountAssets []*table.AccountHistory, err error)
}
