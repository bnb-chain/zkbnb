/*
 * Copyright © 2021 Zkbas Protocol
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package mempool

import (
	"fmt"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/model/account"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"
)

var (
	cacheZkbasMempoolIdPrefix     = "cache:zkbas:mempool:id:"
	cacheZkbasMempoolTxHashPrefix = "cache:zkbas:mempool:txHash:"
)

type (
	MempoolModel interface {
		CreateMempoolTxTable() error
		DropMempoolTxTable() error
		GetMempoolTxByTxId(id uint) (mempoolTx *MempoolTx, err error)
		GetAllMempoolTxsList() (mempoolTxs []*MempoolTx, err error)
		GetMempoolTxsListForCommitter() (mempoolTxs []*MempoolTx, err error)
		GetMempoolTxsList(limit int64, offset int64) (mempoolTxs []*MempoolTx, err error)
		GetMempoolTxsListByAccountIndex(accountIndex int64, limit int64, offset int64) (mempoolTxs []*MempoolTx, err error)
		GetMempoolTxsListByAccountIndexAndTxType(accountIndex int64, txType uint8, limit int64, offset int64) (mempoolTxs []*MempoolTx, err error)
		GetMempoolTxsListByAccountIndexAndTxTypeArray(accountIndex int64, txTypeArray []uint8, limit int64, offset int64) (mempoolTxs []*MempoolTx, err error)
		GetMempoolTxsTotalCount() (count int64, err error)
		GetMempoolTxsTotalCountByAccountIndex(accountIndex int64) (count int64, err error)
		GetMempoolTxsTotalCountByAccountIndexAndTxType(accountIndex int64, txType uint8) (count int64, err error)
		GetMempoolTxsTotalCountByAccountIndexAndTxTypeArray(accountIndex int64, txTypeArray []uint8) (count int64, err error)
		GetMempoolTxsTotalCountByPublicKey(pk string) (count int64, err error)
		GetMempoolTxByTxHash(hash string) (mempoolTxs *MempoolTx, err error)
		GetMempoolTxsByBlockHeight(l2BlockHeight int64) (rowsAffected int64, mempoolTxs []*MempoolTx, err error)
		GetPendingLiquidityTxs() (mempoolTxs []*MempoolTx, err error)
		GetPendingNftTxs() (mempoolTxs []*MempoolTx, err error)
		CreateBatchedMempoolTxs(mempoolTxs []*MempoolTx) error
		CreateMempoolTxAndL2CollectionAndNonce(mempoolTx *MempoolTx, nftInfo *nft.L2NftCollection) error
		CreateMempoolTxAndL2Nft(mempoolTx *MempoolTx, nftInfo *nft.L2Nft) error
		CreateMempoolTxAndL2NftExchange(mempoolTx *MempoolTx, offers []*nft.Offer, nftExchange *nft.L2NftExchange) error
		CreateMempoolTxAndUpdateOffer(mempoolTx *MempoolTx, offer *nft.Offer, isUpdate bool) error
		DeleteMempoolTxs(txIds []*int64) error

		GetPendingMempoolTxsByAccountIndex(accountIndex int64) (mempoolTxs []*MempoolTx, err error)
		GetLatestL2MempoolTxByAccountIndex(accountIndex int64) (mempoolTx *MempoolTx, err error)
	}

	defaultMempoolModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	MempoolTx struct {
		gorm.Model
		TxHash        string `gorm:"uniqueIndex"`
		TxType        int64
		GasFeeAssetId int64
		GasFee        string
		NftIndex      int64
		PairIndex     int64
		AssetId       int64
		TxAmount      string
		NativeAddress string
		TxInfo        string
		ExtraInfo     string
		Memo          string
		AccountIndex  int64
		Nonce         int64
		ExpiredAt     int64
		L2BlockHeight int64
		Status        int `gorm:"index"` // 0: pending tx; 1: committed tx; 2: verified tx;

		MempoolDetails []*MempoolTxDetail `json:"mempool_details" gorm:"foreignKey:TxId"`
	}
)

func NewMempoolModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) MempoolModel {
	return &defaultMempoolModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      MempoolTableName,
		DB:         db,
	}
}

func (*MempoolTx) TableName() string {
	return MempoolTableName
}

/*
	Func: CreateMempoolTxTable
	Params:
	Return: err error
	Description: create MempoolTx table
*/
func (m *defaultMempoolModel) CreateMempoolTxTable() error {
	return m.DB.AutoMigrate(MempoolTx{})
}

/*
	Func: DropMempoolTxTable
	Params:
	Return: err error
	Description: drop MempoolTx table
*/
func (m *defaultMempoolModel) DropMempoolTxTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

/*
	Func: GetAllMempoolTxsList
	Params:
	Return: []*MempoolTx, err error
	Description: used for Init globalMap
*/

func (m *defaultMempoolModel) OrderMempoolTxDetails(tx *MempoolTx) (err error) {
	var mempoolForeignKeyColumn = `MempoolDetails`
	var tmpMempoolTxDetails []*MempoolTxDetail
	err = m.DB.Model(&tx).Association(mempoolForeignKeyColumn).Find(&tmpMempoolTxDetails)
	tx.MempoolDetails = make([]*MempoolTxDetail, len(tmpMempoolTxDetails))
	for i := 0; i < len(tmpMempoolTxDetails); i++ {
		tx.MempoolDetails[tmpMempoolTxDetails[i].Order] = tmpMempoolTxDetails[i]
	}
	return err
}

func (m *defaultMempoolModel) GetAllMempoolTxsList() (mempoolTxs []*MempoolTx, err error) {
	dbTx := m.DB.Table(m.table).Order("created_at, id").Find(&mempoolTxs)
	if dbTx.Error != nil {
		logx.Errorf("[mempool.GetMempoolTxsList] %s", dbTx.Error)
		return nil, dbTx.Error
	}
	// TODO: cache operation
	for _, mempoolTx := range mempoolTxs {
		err := m.OrderMempoolTxDetails(mempoolTx)
		if err != nil {
			logx.Errorf("[mempool.GetMempoolTxsList] Get Associate MempoolDetails Error")
			return nil, err
		}
	}
	return mempoolTxs, nil
}

/*
	Func: GetMempoolTxsList
	Params: limit int, offset int
	Return: []*MempoolTx, err error
	Description: used for /api/v1/txVerification/getMempoolTxsList
*/
func (m *defaultMempoolModel) GetMempoolTxsList(limit int64, offset int64) (mempoolTxs []*MempoolTx, err error) {
	dbTx := m.DB.Table(m.table).Where("status = ?", PendingTxStatus).Limit(int(limit)).Offset(int(offset)).Order("created_at desc, id desc").Find(&mempoolTxs)
	if dbTx.Error != nil {
		logx.Errorf("[mempool.GetMempoolTxsList] %s", dbTx.Error)
		return nil, dbTx.Error
	}
	// TODO: cache operation
	for _, mempoolTx := range mempoolTxs {
		err := m.OrderMempoolTxDetails(mempoolTx)
		if err != nil {
			logx.Errorf("[mempool.GetMempoolTxsList] Get Associate MempoolDetails Error")
			return nil, err
		}
	}
	return mempoolTxs, nil
}

func (m *defaultMempoolModel) GetMempoolTxsByBlockHeight(l2BlockHeight int64) (rowsAffected int64, mempoolTxs []*MempoolTx, err error) {
	dbTx := m.DB.Table(m.table).Where("l2_block_height = ?", l2BlockHeight).Find(&mempoolTxs)
	if dbTx.Error != nil {
		logx.Errorf("[mempool.GetMempoolTxsByBlockHeight] %s", dbTx.Error)
		return 0, nil, dbTx.Error
	}
	// TODO: cache operation
	for _, mempoolTx := range mempoolTxs {
		err := m.OrderMempoolTxDetails(mempoolTx)
		if err != nil {
			logx.Errorf("[mempool.GetMempoolTxsByBlockHeight] Get Associate MempoolDetails Error")
			return 0, nil, err
		}
	}
	return dbTx.RowsAffected, mempoolTxs, nil
}

/*
	Func: GetMempoolTxsListForCommitter
	Return: []*MempoolTx, err error
	Description: query unhandled mempool txVerification
*/

func (m *defaultMempoolModel) GetMempoolTxsListForCommitter() (mempoolTxs []*MempoolTx, err error) {
	dbTx := m.DB.Table(m.table).Where("status = ?", PendingTxStatus).Order("created_at, id").Find(&mempoolTxs)
	if dbTx.Error != nil {
		logx.Errorf("[mempool.GetMempoolTxsList] %s", dbTx.Error)
		return nil, dbTx.Error
	}
	// TODO: cache operation
	for _, mempoolTx := range mempoolTxs {
		err := m.OrderMempoolTxDetails(mempoolTx)
		if err != nil {
			logx.Errorf("[mempool.GetMempoolTxsList] Get Associate MempoolDetails Error")
			return nil, err
		}
	}
	return mempoolTxs, nil
}

/*
	Func: GetMempoolTxsListByAccountIndex
	Params: accountIndex int64, limit int, offset int
	Return: []*MempoolTx, err error
	Description: used for /api/v1/txVerification/getMempoolTxsListByAccountIndex
*/

func (m *defaultMempoolModel) GetMempoolTxsListByAccountIndex(accountIndex int64, limit int64, offset int64) (mempoolTxs []*MempoolTx, err error) {
	var (
		mempoolDetailTable = `mempool_tx_detail`
		mempoolIds         []int64
	)
	var mempoolTxDetails []*MempoolTxDetail
	dbTx := m.DB.Table(mempoolDetailTable).Select("tx_id").Where("account_index = ?", accountIndex).Find(&mempoolTxDetails).Group("tx_id").Find(&mempoolIds)
	if dbTx.Error != nil {
		logx.Errorf("[mempool.GetMempoolTxsListByAccountIndex] %s", dbTx.Error)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Info("[mempool.GetMempoolTxsListByAccountIndex] No rows in mempool list")
		return nil, ErrNotFound
	}

	dbTx = m.DB.Table(m.table).Where("status = ?", PendingTxStatus).Order("created_at desc").Offset(int(offset)).Limit(int(limit)).Find(&mempoolTxs, mempoolIds)
	if dbTx.Error != nil {
		logx.Errorf("[mempool.GetMempoolTxsListByAccountIndex] %s", dbTx.Error)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Info("[mempool.GetMempoolTxsListByAccountIndex] No rows in mempool with Pending Status")
		return nil, ErrNotFound
	}
	// TODO: cache operation
	for _, mempoolTx := range mempoolTxs {
		err := m.OrderMempoolTxDetails(mempoolTx)
		if err != nil {
			logx.Errorf("[mempool.GetMempoolTxsListByAccountIndex] Get Associate MempoolDetails Error")
			return nil, err
		}
	}
	return mempoolTxs, nil
}

/*
	Func: GetMempoolTxsListByAccountIndexAndTxType
	Params: accountIndex int64, txType uint8, limit int64, offset int64
	Return: []*MempoolTx, err error
	Description:
*/

func (m *defaultMempoolModel) GetMempoolTxsListByAccountIndexAndTxType(accountIndex int64, txType uint8, limit int64, offset int64) (mempoolTxs []*MempoolTx, err error) {
	var (
		mempoolDetailTable = `mempool_tx_detail`
		mempoolIds         []int64
	)
	var mempoolTxDetails []*MempoolTxDetail
	dbTx := m.DB.Table(mempoolDetailTable).Select("tx_id").Where("account_index = ?", accountIndex).Find(&mempoolTxDetails).Group("tx_id").Find(&mempoolIds)
	if dbTx.Error != nil {
		logx.Errorf("[mempool.GetMempoolTxsListByAccountIndexAndTxType] %s", dbTx.Error)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Errorf("[mempool.GetMempoolTxsListByAccountIndexAndTxType] Get MempoolIds Error")
		return nil, ErrNotFound
	}
	dbTx = m.DB.Table(m.table).Where("status = ? and tx_type = ?", PendingTxStatus, txType).Order("created_at desc").Offset(int(offset)).Limit(int(limit)).Find(&mempoolTxs, mempoolIds)
	if dbTx.Error != nil {
		logx.Errorf("[mempool.GetMempoolTxsListByAccountIndexAndTxType] %s", dbTx.Error)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Errorf("[mempool.GetMempoolTxsListByAccountIndexAndTxType] Get MempoolTxs Error")
		return nil, ErrNotFound
	}
	// TODO: cache operation
	for _, mempoolTx := range mempoolTxs {
		err := m.OrderMempoolTxDetails(mempoolTx)
		if err != nil {
			logx.Errorf("[mempool.GetMempoolTxsListByAccountIndexAndTxType] Get Associate MempoolDetails Error")
			return nil, err
		}
	}
	return mempoolTxs, nil
}

func (m *defaultMempoolModel) GetMempoolTxsListByAccountIndexAndTxTypeArray(accountIndex int64, txTypeArray []uint8, limit int64, offset int64) (mempoolTxs []*MempoolTx, err error) {
	var (
		mempoolDetailTable = `mempool_tx_detail`
		mempoolIds         []int64
	)
	var mempoolTxDetails []*MempoolTxDetail
	dbTx := m.DB.Table(mempoolDetailTable).Select("tx_id").Where("account_index = ?", accountIndex).Find(&mempoolTxDetails).Group("tx_id").Find(&mempoolIds)
	if dbTx.Error != nil {
		logx.Errorf("[mempool.GetMempoolTxsListByAccountIndexAndTxType] %s", dbTx.Error)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Errorf("[mempool.GetMempoolTxsListByAccountIndexAndTxType] Get MempoolIds Error")
		return nil, ErrNotFound
	}
	dbTx = m.DB.Table(m.table).Where("status = ? and tx_type in (?)", PendingTxStatus, txTypeArray).Order("created_at desc").Offset(int(offset)).Limit(int(limit)).Find(&mempoolTxs, mempoolIds)
	if dbTx.Error != nil {
		logx.Errorf("[mempool.GetMempoolTxsListByAccountIndexAndTxType] %s", dbTx.Error)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Errorf("[mempool.GetMempoolTxsListByAccountIndexAndTxType] Get MempoolTxs Error")
		return nil, ErrNotFound
	}
	// TODO: cache operation
	for _, mempoolTx := range mempoolTxs {
		err := m.OrderMempoolTxDetails(mempoolTx)
		if err != nil {
			logx.Errorf("[mempool.GetMempoolTxsListByAccountIndexAndTxType] Get Associate MempoolDetails Error")
			return nil, err
		}
	}
	return mempoolTxs, nil
}

/*
	Func: GetMempoolTxsTotalCount
	Params:
	Return: count int64, err error
	Description: used for counting total transactions in mempool for explorer dashboard
*/
func (m *defaultMempoolModel) GetMempoolTxsTotalCount() (count int64, err error) {
	dbTx := m.DB.Table(m.table).Where("status = ? and deleted_at is NULL", PendingTxStatus).Count(&count)
	if dbTx.Error != nil {
		logx.Errorf("[txVerification.GetTxsTotalCount] %s", dbTx.Error)
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return count, nil
}

/*
	Func: GetMempoolTxsTotalCountByAccountIndex
	Params: accountIndex int64
	Return: count int64, err error
	Description: used for counting total transactions in mempool for explorer dashboard
*/
func (m *defaultMempoolModel) GetMempoolTxsTotalCountByAccountIndex(accountIndex int64) (count int64, err error) {
	var (
		mempoolDetailTable = `mempool_tx_detail`
		mempoolIds         []int64
	)
	var mempoolTxDetails []*MempoolTxDetail
	dbTx := m.DB.Table(mempoolDetailTable).Select("tx_id").Where("account_index = ?", accountIndex).Find(&mempoolTxDetails).Group("tx_id").Find(&mempoolIds)
	if dbTx.Error != nil {
		logx.Errorf("[mempool.GetMempoolTxsTotalCountByAccountIndex] %s", dbTx.Error)
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	dbTx = m.DB.Table(m.table).Where("status = ? and id in (?) and deleted_at is NULL", PendingTxStatus, mempoolIds).Count(&count)
	if dbTx.Error != nil {
		logx.Errorf("[mempool.GetMempoolTxsTotalCountByAccountIndex] %s", dbTx.Error)
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Infof("[mempool.GetMempoolTxsTotalCountByAccountIndex] no txVerification of account index %d in mempool", accountIndex)
		return 0, nil
	}
	return count, nil
}

/*
	Func: GetMempoolTxsTotalCountByAccountIndexAndTxType
	Params: accountIndex int64, txType uint8
	Return: count int64, err error
	Description: used for counting total transactions in mempool for explorer dashboard
*/
func (m *defaultMempoolModel) GetMempoolTxsTotalCountByAccountIndexAndTxType(accountIndex int64, txType uint8) (count int64, err error) {
	var (
		mempoolDetailTable = `mempool_tx_detail`
		mempoolIds         []int64
	)
	var mempoolTxDetails []*MempoolTxDetail
	dbTx := m.DB.Table(mempoolDetailTable).Select("tx_id").Where("account_index = ?", accountIndex).Find(&mempoolTxDetails).Group("tx_id").Find(&mempoolIds)
	if dbTx.Error != nil {
		logx.Errorf("[mempool.GetMempoolTxsTotalCountByAccountIndexAndTxType] %s", dbTx.Error)
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	dbTx = m.DB.Table(m.table).Where("status = ? and id in (?) and deleted_at is NULL  and tx_type = ?", PendingTxStatus, mempoolIds, txType).Count(&count)
	if dbTx.Error != nil {
		logx.Errorf("[mempool.GetMempoolTxsTotalCountByAccountIndexAndTxType] %s", dbTx.Error)
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Infof("[mempool.GetMempoolTxsTotalCountByAccountIndexAndTxType] no txVerification of account index %d and txVerification type = %d in mempool", accountIndex, txType)
		return 0, nil
	}
	return count, nil
}

/*
	Func: GetMempoolTxsTotalCountByAccountIndexAndTxTypeArray
	Params: accountIndex int64, txTypeArray []uint8
	Return: count int64, err error
	Description: used for counting total transactions in mempool for explorer dashboard
*/
func (m *defaultMempoolModel) GetMempoolTxsTotalCountByAccountIndexAndTxTypeArray(accountIndex int64, txTypeArray []uint8) (count int64, err error) {
	var (
		mempoolDetailTable = `mempool_tx_detail`
		mempoolIds         []int64
	)
	var mempoolTxDetails []*MempoolTxDetail
	dbTx := m.DB.Table(mempoolDetailTable).Select("tx_id").Where("account_index = ?", accountIndex).Find(&mempoolTxDetails).Group("tx_id").Find(&mempoolIds)
	if dbTx.Error != nil {
		logx.Errorf("[mempool.GetMempoolTxsTotalCountByAccountIndexAndTxType] %s", dbTx.Error)
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	dbTx = m.DB.Table(m.table).Where("status = ? and id in (?) and deleted_at is NULL  and tx_type in (?)", PendingTxStatus, mempoolIds, txTypeArray).Count(&count)
	if dbTx.Error != nil {
		logx.Errorf("[mempool.GetMempoolTxsTotalCountByAccountIndexAndTxType] %s", dbTx.Error)
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Infof("[mempool.GetMempoolTxsTotalCountByAccountIndexAndTxType] no txVerification of account index %d and txVerification type = %v in mempool", accountIndex, txTypeArray)
		return 0, nil
	}
	return count, nil
}

/*
	Func: GetMempoolTxsTotalCountByPublicKey
	Params: pk string
	Return: count int64, err error
	Description: used for counting total transactions in mempool for explorer dashboard
*/
func (m *defaultMempoolModel) GetMempoolTxsTotalCountByPublicKey(pk string) (count int64, err error) {
	var (
		accountTable       = `account`
		accountIndex       int64
		mempoolDetailTable = `mempool_tx_detail`
		mempoolIds         []int64
	)
	dbTx := m.DB.Table(accountTable).Select("account_index").Where("public_key = ?", pk).Find(&accountIndex)
	if dbTx.Error != nil {
		logx.Errorf("[mempool.GetMempoolTxsListByPublicKey] %s", dbTx.Error)
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	var mempoolTxDetails []*MempoolTxDetail
	dbTx = m.DB.Table(mempoolDetailTable).Select("tx_id").Where("account_index = ?", accountIndex).Find(&mempoolTxDetails).Group("tx_id").Find(&mempoolIds)
	if dbTx.Error != nil {
		logx.Errorf("[mempool.GetMempoolTxsListByPublicKey] %s", dbTx.Error)
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	dbTx = m.DB.Table(m.table).Where("status = ? and id in (?) and deleted_at is NULL", PendingTxStatus, mempoolIds).Count(&count)
	if dbTx.Error != nil {
		logx.Errorf("[mempool.GetMempoolTxsListByPublicKey] %s", dbTx.Error)
		return 0, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Infof("[mempool.GetMempoolTxsListByPublicKey] no txVerification of account index %d in mempool", accountIndex)
		return 0, nil
	}
	return count, nil
}

/*
	Func: GetMempoolTxByTxHash
	Params: hash string
	Return: mempoolTxs *MempoolTx, err error
	Description: used for get  transactions in mempool by txVerification hash
*/
func (m *defaultMempoolModel) GetMempoolTxByTxHash(hash string) (mempoolTx *MempoolTx, err error) {
	dbTx := m.DB.Table(m.table).Where("status = ? and tx_hash = ?", PendingTxStatus, hash).Find(&mempoolTx)
	if dbTx.Error != nil {
		if dbTx.Error == ErrNotFound {
			return mempoolTx, dbTx.Error
		} else {
			err := fmt.Sprintf("[mempool.GetMempoolTxByTxHash] %s", dbTx.Error)
			logx.Errorf(err)
			return nil, dbTx.Error
		}
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[mempool.GetMempoolTxByTxHash] %s", ErrNotFound)
		logx.Info(err)
		return nil, ErrNotFound
	}
	err = m.OrderMempoolTxDetails(mempoolTx)
	if err != nil {
		logx.Errorf("[mempool.GetMempoolTxByTxHash] Get Associate MempoolDetails Error")
		return nil, err
	}
	return mempoolTx, nil
}

/*
	Func: CreateBatchedMempoolTxs
	Params: []*MempoolTx
	Return: error
	Description: Insert MempoolTxs when sendTx request.
*/

func (m *defaultMempoolModel) CreateBatchedMempoolTxs(mempoolTxs []*MempoolTx) error {
	return m.DB.Transaction(func(tx *gorm.DB) error { // transact
		dbTx := tx.Table(m.table).Create(mempoolTxs)
		if dbTx.Error != nil {
			logx.Errorf("[mempool.CreateMempoolTxs] %s", dbTx.Error)
			return dbTx.Error
		}
		if dbTx.RowsAffected == 0 {
			logx.Errorf("[mempool.CreateMempoolTxs] Create Invalid Mempool Tx")
			return ErrInvalidMempoolTx
		}
		return nil
	})
}

/*
	Func: DeleteMempoolTxs
	Params: TxId []*int64
	Return: error
	Description: Delete MempoolTxs when Committer pack new layer2 block.
*/
func (m *defaultMempoolModel) DeleteMempoolTxs(txIds []*int64) error {
	//var mempoolDetailTable = `mempool_tx_detail`
	// TODO: clean cache operation
	return m.DB.Transaction(func(tx *gorm.DB) error { // transact
		for _, txId := range txIds {
			var mempoolTx *MempoolTx
			dbTx := tx.Table(m.table).Where("id = ?", txId).Delete(&mempoolTx)
			if dbTx.Error != nil {
				logx.Errorf("[mempool.DeleteMempoolTxs] %s", dbTx.Error)
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				logx.Errorf("[mempool.DeleteMempoolTxs] Delete Invalid Mempool Tx")
				return ErrInvalidMempoolTx
			}

			//var mempoolTxDetail *MempoolTxDetail
			//dbTx = m.DB.Table(mempoolDetailTable).Where("tx_id = ?", txId).Delete(&mempoolTxDetail)
			//if dbTx.Error != nil {
			//	logx.Errorf("[mempool.DeleteMempoolTxs] %s", dbTx.Error)
			//	return dbTx.Error
			//}
			//if dbTx.RowsAffected == 0 {
			//	logx.Errorf("[mempool.DeleteMempoolTxs] Delete Invalid Mempool TxDtail")
			//	return ErrInvalidMempoolTxDetail
			//}
		}
		return nil
	})
}

/*
	Func: GetMempoolTxIdsListByL2BlockHeight
	Params: blockHeight
	Return: []*MempoolTx, err error
	Description: used for verifier get txIds from Mempool and deleting the transaction in mempool table after
*/
func (m *defaultMempoolModel) GetMempoolTxsListByL2BlockHeight(blockHeight int64) (mempoolTxs []*MempoolTx, err error) {
	dbTx := m.DB.Table(m.table).Where("status = ? and l2_block_height <= ?", SuccessTxStatus, blockHeight).Find(&mempoolTxs)
	if dbTx.Error != nil {
		logx.Errorf("[mempool.GetMempoolTxsListByL2BlockHeight] %s", dbTx.Error)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Errorf("[mempool.GetMempoolTxsListByL2BlockHeight] Get MempoolTxs Error")
		return nil, ErrNotFound
	}

	return mempoolTxs, nil
}

func (m *defaultMempoolModel) GetLatestL2MempoolTxByAccountIndex(accountIndex int64) (mempoolTx *MempoolTx, err error) {
	dbTx := m.DB.Table(m.table).Where("account_index = ? and nonce != -1", accountIndex).
		Order("created_at desc, id desc").Find(&mempoolTx)
	if dbTx.Error != nil {
		logx.Errorf("[GetLatestL2MempoolTxByAccountIndex] %s", dbTx.Error)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Errorf("[GetLatestL2MempoolTxByAccountIndex] Get MempoolTxs Error")
		return nil, ErrNotFound
	}
	return mempoolTx, nil
}

func (m *defaultMempoolModel) GetPendingMempoolTxsByAccountIndex(accountIndex int64) (mempoolTxs []*MempoolTx, err error) {
	dbTx := m.DB.Table(m.table).Where("status = ? AND account_index = ?", PendingTxStatus, accountIndex).
		Order("created_at, id").Find(&mempoolTxs)
	if dbTx.Error != nil {
		logx.Errorf("[GetPendingMempoolTxsByAccountIndex] %s", dbTx.Error)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Errorf("[GetPendingMempoolTxsByAccountIndex] Get MempoolTxs Error")
		return nil, ErrNotFound
	}
	for _, mempoolTx := range mempoolTxs {
		err = m.OrderMempoolTxDetails(mempoolTx)
		if err != nil {
			logx.Errorf("[GetPendingMempoolTxsByAccountIndex] Get Associate MempoolDetails Error")
			return nil, err
		}
	}
	return mempoolTxs, nil
}

func (m *defaultMempoolModel) GetPendingLiquidityTxs() (mempoolTxs []*MempoolTx, err error) {
	dbTx := m.DB.Table(m.table).Where("status = ? and pair_index != ?", PendingTxStatus, commonConstant.NilPairIndex).
		Find(&mempoolTxs)
	if dbTx.Error != nil {
		errInfo := fmt.Sprintf("[mempool.GetMempoolTxByTxHash] %s", dbTx.Error)
		logx.Errorf(errInfo)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[mempool.GetMempoolTxByTxHash] %s", ErrNotFound)
		logx.Info(err)
		return nil, ErrNotFound
	}
	for _, mempoolTx := range mempoolTxs {
		err = m.OrderMempoolTxDetails(mempoolTx)
		if err != nil {
			logx.Errorf("[mempool.GetMempoolTxByTxHash] Get Associate MempoolDetails Error")
			return nil, err
		}
	}
	return mempoolTxs, nil
}

func (m *defaultMempoolModel) GetPendingNftTxs() (mempoolTxs []*MempoolTx, err error) {
	dbTx := m.DB.Table(m.table).Where("status = ? and nft_index != ?", PendingTxStatus, commonConstant.NilTxNftIndex).
		Find(&mempoolTxs)
	if dbTx.Error != nil {
		errInfo := fmt.Sprintf("[mempool.GetMempoolTxByTxHash] %s", dbTx.Error)
		logx.Errorf(errInfo)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[mempool.GetMempoolTxByTxHash] %s", ErrNotFound)
		logx.Info(err)
		return nil, ErrNotFound
	}
	for _, mempoolTx := range mempoolTxs {
		err = m.OrderMempoolTxDetails(mempoolTx)
		if err != nil {
			logx.Errorf("[mempool.GetMempoolTxByTxHash] Get Associate MempoolDetails Error")
			return nil, err
		}
	}
	return mempoolTxs, nil
}

func (m *defaultMempoolModel) CreateMempoolTxAndL2CollectionAndNonce(mempoolTx *MempoolTx, nftCollectionInfo *nft.L2NftCollection) error {
	return m.DB.Transaction(func(db *gorm.DB) error { // transact
		dbTx := db.Table(m.table).Create(mempoolTx)
		if dbTx.Error != nil {
			logx.Errorf("[mempool.CreateMempoolTxAndL2Collection] %s", dbTx.Error)
			return dbTx.Error
		}
		if dbTx.RowsAffected == 0 {
			logx.Errorf("[mempool.CreateMempoolTxAndL2Collection] Create Invalid Mempool Tx")
			return ErrInvalidMempoolTx
		}
		dbTx = db.Table(nft.L2NftCollectionTableName).Create(nftCollectionInfo)
		if dbTx.Error != nil {
			logx.Errorf("[mempool.CreateMempoolTxAndL2Collection] %s", dbTx.Error)
			return dbTx.Error
		}
		if dbTx.RowsAffected == 0 {
			logx.Errorf("[mempool.CreateMempoolTxAndL2Collection] Create Invalid nft collection info")
			return ErrInvalidMempoolTx
		}
		err := db.Model(&account.Account{}).Where("account_index = ?", nftCollectionInfo.AccountIndex).Update("collection_nonce", nftCollectionInfo.CollectionId)
		if err != nil {
			logx.Errorf("[mempool.CreateMempoolTxAndL2Collection] %s", err)
			return dbTx.Error
		}
		return nil
	})
}

func (m *defaultMempoolModel) CreateMempoolTxAndL2Nft(mempoolTx *MempoolTx, nftInfo *nft.L2Nft) error {
	return m.DB.Transaction(func(tx *gorm.DB) error { // transact
		dbTx := tx.Table(m.table).Create(mempoolTx)
		if dbTx.Error != nil {
			logx.Errorf("[mempool.CreateMempoolTxAndL2Nft] %s", dbTx.Error)
			return dbTx.Error
		}
		if dbTx.RowsAffected == 0 {
			logx.Errorf("[mempool.CreateMempoolTxAndL2Nft] Create Invalid Mempool Tx")
			return ErrInvalidMempoolTx
		}
		dbTx = tx.Table(nft.L2NftTableName).Create(nftInfo)
		if dbTx.Error != nil {
			logx.Errorf("[mempool.CreateMempoolTxAndL2Nft] %s", dbTx.Error)
			return dbTx.Error
		}
		if dbTx.RowsAffected == 0 {
			logx.Errorf("[mempool.CreateMempoolTxAndL2Nft] Create Invalid nft info")
			return ErrInvalidMempoolTx
		}
		return nil
	})
}

func (m *defaultMempoolModel) CreateMempoolTxAndL2NftExchange(mempoolTx *MempoolTx, offers []*nft.Offer, nftExchange *nft.L2NftExchange) error {
	return m.DB.Transaction(func(tx *gorm.DB) error { // transact
		dbTx := tx.Table(m.table).Create(mempoolTx)
		if dbTx.Error != nil {
			logx.Errorf("[mempool.CreateMempoolTxAndL2NftExchange] %s", dbTx.Error)
			return dbTx.Error
		}
		if dbTx.RowsAffected == 0 {
			logx.Errorf("[mempool.CreateMempoolTxAndL2NftExchange] Create Invalid Mempool Tx")
			return ErrInvalidMempoolTx
		}
		if len(offers) != 0 {
			dbTx = tx.Table(nft.OfferTableName).CreateInBatches(offers, len(offers))
			if dbTx.Error != nil {
				logx.Errorf("[mempool.CreateMempoolTxAndL2NftExchange] %s", dbTx.Error)
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				logx.Errorf("[mempool.CreateMempoolTxAndL2NftExchange] Create Invalid nft info")
				return ErrInvalidMempoolTx
			}
		}
		dbTx = tx.Table(nft.L2NftExchangeTableName).Create(nftExchange)
		if dbTx.Error != nil {
			logx.Errorf("[mempool.CreateMempoolTxAndL2NftExchange] %s", dbTx.Error)
			return dbTx.Error
		}
		if dbTx.RowsAffected == 0 {
			logx.Errorf("[mempool.CreateMempoolTxAndL2NftExchange] Create Invalid nft info")
			return ErrInvalidMempoolTx
		}
		return nil
	})
}

func (m *defaultMempoolModel) CreateMempoolTxAndUpdateOffer(mempoolTx *MempoolTx, offer *nft.Offer, isUpdate bool) error {
	return m.DB.Transaction(func(tx *gorm.DB) error { // transact
		dbTx := tx.Table(m.table).Create(mempoolTx)
		if dbTx.Error != nil {
			logx.Errorf("[mempool.CreateMempoolTxAndUpdateOffer] %s", dbTx.Error)
			return dbTx.Error
		}
		if dbTx.RowsAffected == 0 {
			logx.Errorf("[mempool.CreateMempoolTxAndUpdateOffer] Create Invalid Mempool Tx")
			return ErrInvalidMempoolTx
		}
		if isUpdate {
			dbTx = tx.Table(nft.OfferTableName).Where("id = ?", offer.ID).Select("*").Updates(&offer)
			if dbTx.Error != nil {
				logx.Errorf("[mempool.CreateMempoolTxAndUpdateOffer] %s", dbTx.Error)
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				logx.Errorf("[mempool.CreateMempoolTxAndUpdateOffer] Create Invalid nft info")
				return ErrInvalidMempoolTx
			}
		} else {
			dbTx = tx.Table(nft.OfferTableName).Create(offer)
			if dbTx.Error != nil {
				logx.Errorf("[mempool.CreateMempoolTxAndUpdateOffer] %s", dbTx.Error)
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				logx.Errorf("[mempool.CreateMempoolTxAndUpdateOffer] Create Invalid nft info")
				return ErrInvalidMempoolTx
			}
		}
		return nil
	})
}

func (m *defaultMempoolModel) GetMempoolTxByTxId(id uint) (mempoolTx *MempoolTx, err error) {
	dbTx := m.DB.Table(m.table).Where("id = ?", id).
		Find(&mempoolTx)
	if dbTx.Error != nil {
		errInfo := fmt.Sprintf("[mempool.GetMempoolTxByTxId] %s", dbTx.Error)
		logx.Errorf(errInfo)
		return nil, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[mempool.GetMempoolTxByTxId] %s", ErrNotFound)
		logx.Info(err)
		return nil, ErrNotFound
	}
	err = m.OrderMempoolTxDetails(mempoolTx)
	if err != nil {
		logx.Errorf("[mempool.GetMempoolTxByTxHash] Get Associate MempoolDetails Error")
		return nil, err
	}
	return mempoolTx, nil
}
