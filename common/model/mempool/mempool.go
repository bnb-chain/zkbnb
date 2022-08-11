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

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/common/model/nft"
)

type (
	MemPoolModel interface {
		CreateMempoolTxTable() error
		DropMempoolTxTable() error
		GetMempoolTxByTxId(id int64) (mempoolTx *MempoolTx, err error)
		GetMempoolTxsListForCommitter() (mempoolTxs []*MempoolTx, err error)
		GetMempoolTxsList(limit int64, offset int64) (mempoolTxs []*MempoolTx, err error)
		GetMempoolTxsTotalCount() (count int64, err error)
		GetMempoolTxByTxHash(hash string) (mempoolTxs *MempoolTx, err error)
		GetMempoolTxsByBlockHeight(l2BlockHeight int64) (rowsAffected int64, mempoolTxs []*MempoolTx, err error)
		GetPendingLiquidityTxs() (mempoolTxs []*MempoolTx, err error)
		GetPendingNftTxs() (mempoolTxs []*MempoolTx, err error)
		CreateBatchedMempoolTxs(mempoolTxs []*MempoolTx) error
		CreateMempoolTxAndL2CollectionAndNonce(mempoolTx *MempoolTx, nftInfo *nft.L2NftCollection) error
		CreateMempoolTxAndL2Nft(mempoolTx *MempoolTx, nftInfo *nft.L2Nft) error
		CreateMempoolTxAndL2NftExchange(mempoolTx *MempoolTx, offers []*nft.Offer, nftExchange *nft.L2NftExchange) error
		CreateMempoolTxAndUpdateOffer(mempoolTx *MempoolTx, offer *nft.Offer, isUpdate bool) error

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

func NewMempoolModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) MemPoolModel {
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

/*
Func: GetMempoolTxsList
Params: limit int, offset int
Return: []*MempoolTx, err error
Description: used for /api/v1/txVerification/getMempoolTxsList
*/
func (m *defaultMempoolModel) GetMempoolTxsList(limit int64, offset int64) (mempoolTxs []*MempoolTx, err error) {
	dbTx := m.DB.Table(m.table).Where("status = ?", PendingTxStatus).Limit(int(limit)).Offset(int(offset)).Order("created_at desc, id desc").Find(&mempoolTxs)
	if dbTx.Error != nil {
		logx.Errorf("get mempool tx errors, err: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	}
	// TODO: cache operation
	for _, mempoolTx := range mempoolTxs {
		err := m.OrderMempoolTxDetails(mempoolTx)
		if err != nil {
			logx.Errorf("get associate mempool details error, err: %s", err.Error())
			return nil, err
		}
	}
	return mempoolTxs, nil
}

func (m *defaultMempoolModel) GetMempoolTxsByBlockHeight(l2BlockHeight int64) (rowsAffected int64, mempoolTxs []*MempoolTx, err error) {
	dbTx := m.DB.Table(m.table).Where("l2_block_height = ?", l2BlockHeight).Find(&mempoolTxs)
	if dbTx.Error != nil {
		logx.Errorf("get mempool tx errors, err: %s", dbTx.Error.Error())
		return 0, nil, errorcode.DbErrSqlOperation
	}
	// TODO: cache operation
	for _, mempoolTx := range mempoolTxs {
		err := m.OrderMempoolTxDetails(mempoolTx)
		if err != nil {
			logx.Errorf("get associate mempool details error, err: %s", err.Error())
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
		logx.Errorf("get mempool tx errors, err: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	}
	// TODO: cache operation
	for _, mempoolTx := range mempoolTxs {
		err := m.OrderMempoolTxDetails(mempoolTx)
		if err != nil {
			logx.Errorf("get associate mempool details error, err: %s", err.Error())
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
		logx.Errorf("get mempool tx count error, err: %s", dbTx.Error.Error())
		return 0, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
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
		if dbTx.Error == errorcode.DbErrNotFound {
			return mempoolTx, dbTx.Error
		} else {
			logx.Errorf("get mempool tx error, err: %s", dbTx.Error.Error())
			return nil, errorcode.DbErrSqlOperation
		}
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}
	err = m.OrderMempoolTxDetails(mempoolTx)
	if err != nil {
		logx.Errorf("get associate mempool details error, err: %s", err.Error())
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
			logx.Errorf("create mempool tx error, err: %s", dbTx.Error.Error())
			return dbTx.Error
		}
		if dbTx.RowsAffected == 0 {
			return errorcode.DbErrFailToCreateMempoolTx
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
		logx.Errorf("get mempool txs error, err: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}

	return mempoolTxs, nil
}

func (m *defaultMempoolModel) GetLatestL2MempoolTxByAccountIndex(accountIndex int64) (mempoolTx *MempoolTx, err error) {
	dbTx := m.DB.Table(m.table).Where("account_index = ? and nonce != -1", accountIndex).
		Order("created_at desc, id desc").Find(&mempoolTx)
	if dbTx.Error != nil {
		logx.Errorf("get mempool txs error, err: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}
	return mempoolTx, nil
}

func (m *defaultMempoolModel) GetPendingMempoolTxsByAccountIndex(accountIndex int64) (mempoolTxs []*MempoolTx, err error) {
	dbTx := m.DB.Table(m.table).Where("status = ? AND account_index = ?", PendingTxStatus, accountIndex).
		Order("created_at, id").Find(&mempoolTxs)
	if dbTx.Error != nil {
		logx.Errorf("get mempool txs error, err: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}
	for _, mempoolTx := range mempoolTxs {
		err = m.OrderMempoolTxDetails(mempoolTx)
		if err != nil {
			logx.Errorf("get associate mempool details error, err: %s", err.Error())
			return nil, err
		}
	}
	return mempoolTxs, nil
}

func (m *defaultMempoolModel) GetPendingLiquidityTxs() (mempoolTxs []*MempoolTx, err error) {
	dbTx := m.DB.Table(m.table).Where("status = ? and pair_index != ?", PendingTxStatus, commonConstant.NilPairIndex).
		Find(&mempoolTxs)
	if dbTx.Error != nil {
		logx.Errorf("get mempool txs error, err: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}
	for _, mempoolTx := range mempoolTxs {
		err = m.OrderMempoolTxDetails(mempoolTx)
		if err != nil {
			logx.Errorf("get associate mempool details error, err: %s", err.Error())
			return nil, err
		}
	}
	return mempoolTxs, nil
}

func (m *defaultMempoolModel) GetPendingNftTxs() (mempoolTxs []*MempoolTx, err error) {
	dbTx := m.DB.Table(m.table).Where("status = ? and nft_index != ?", PendingTxStatus, commonConstant.NilTxNftIndex).
		Find(&mempoolTxs)
	if dbTx.Error != nil {
		logx.Errorf("get pending nft txs error, err: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		return nil, errorcode.DbErrNotFound
	}
	for _, mempoolTx := range mempoolTxs {
		err = m.OrderMempoolTxDetails(mempoolTx)
		if err != nil {
			logx.Errorf("get associate mempool details error, err: %s", err.Error())
			return nil, err
		}
	}
	return mempoolTxs, nil
}

func (m *defaultMempoolModel) CreateMempoolTxAndL2CollectionAndNonce(mempoolTx *MempoolTx, nftCollectionInfo *nft.L2NftCollection) error {
	return m.DB.Transaction(func(db *gorm.DB) error { // transact
		dbTx := db.Table(m.table).Create(mempoolTx)
		if dbTx.Error != nil {
			logx.Errorf("create mempool tx error, err: %s", dbTx.Error.Error())
			return errorcode.DbErrSqlOperation
		}
		if dbTx.RowsAffected == 0 {
			return errorcode.DbErrFailToCreateMempoolTx
		}
		dbTx = db.Table(nft.L2NftCollectionTableName).Create(nftCollectionInfo)
		if dbTx.Error != nil {
			logx.Errorf("create collection error, err: %s", dbTx.Error.Error())
			return errorcode.DbErrSqlOperation
		}
		if dbTx.RowsAffected == 0 {
			return errorcode.DbErrFailToCreateMempoolTx
		}
		return nil
	})
}

func (m *defaultMempoolModel) CreateMempoolTxAndL2Nft(mempoolTx *MempoolTx, nftInfo *nft.L2Nft) error {
	return m.DB.Transaction(func(tx *gorm.DB) error { // transact
		dbTx := tx.Table(m.table).Create(mempoolTx)
		if dbTx.Error != nil {
			logx.Errorf("create mempool tx error, err: %s", dbTx.Error.Error())
			return errorcode.DbErrSqlOperation
		}
		if dbTx.RowsAffected == 0 {
			return errorcode.DbErrFailToCreateMempoolTx
		}
		dbTx = tx.Table(nft.L2NftTableName).Create(nftInfo)
		if dbTx.Error != nil {
			logx.Errorf("create nft error, err: %s", dbTx.Error.Error())
			return errorcode.DbErrSqlOperation
		}
		if dbTx.RowsAffected == 0 {
			return errorcode.DbErrFailToCreateMempoolTx
		}
		return nil
	})
}

func (m *defaultMempoolModel) CreateMempoolTxAndL2NftExchange(mempoolTx *MempoolTx, offers []*nft.Offer, nftExchange *nft.L2NftExchange) error {
	return m.DB.Transaction(func(tx *gorm.DB) error { // transact
		dbTx := tx.Table(m.table).Create(mempoolTx)
		if dbTx.Error != nil {
			logx.Errorf("create mempool tx error, err: %s", dbTx.Error.Error())
			return dbTx.Error
		}
		if dbTx.RowsAffected == 0 {
			return errorcode.DbErrFailToCreateMempoolTx
		}
		if len(offers) != 0 {
			dbTx = tx.Table(nft.OfferTableName).CreateInBatches(offers, len(offers))
			if dbTx.Error != nil {
				logx.Errorf("create offers error, err: %s", dbTx.Error.Error())
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				return errorcode.DbErrFailToCreateMempoolTx
			}
		}
		dbTx = tx.Table(nft.L2NftExchangeTableName).Create(nftExchange)
		if dbTx.Error != nil {
			logx.Errorf("create nft exchange error, err: %s", dbTx.Error.Error())
			return dbTx.Error
		}
		if dbTx.RowsAffected == 0 {
			return errorcode.DbErrFailToCreateMempoolTx
		}
		return nil
	})
}

func (m *defaultMempoolModel) CreateMempoolTxAndUpdateOffer(mempoolTx *MempoolTx, offer *nft.Offer, isUpdate bool) error {
	return m.DB.Transaction(func(tx *gorm.DB) error { // transact
		dbTx := tx.Table(m.table).Create(mempoolTx)
		if dbTx.Error != nil {
			logx.Errorf("create mempool tx error, err: %s", dbTx.Error.Error())
			return dbTx.Error
		}
		if dbTx.RowsAffected == 0 {
			return errorcode.DbErrFailToCreateMempoolTx
		}
		if isUpdate {
			dbTx = tx.Table(nft.OfferTableName).Where("id = ?", offer.ID).Select("*").Updates(&offer)
			if dbTx.Error != nil {
				logx.Errorf("update offer error, err: %s", dbTx.Error.Error())
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				return errorcode.DbErrFailToCreateMempoolTx
			}
		} else {
			dbTx = tx.Table(nft.OfferTableName).Create(offer)
			if dbTx.Error != nil {
				logx.Errorf("create offer error, err: %s", dbTx.Error.Error())
				return dbTx.Error
			}
			if dbTx.RowsAffected == 0 {
				return errorcode.DbErrFailToCreateMempoolTx
			}
		}
		return nil
	})
}

func (m *defaultMempoolModel) GetMempoolTxByTxId(id int64) (mempoolTx *MempoolTx, err error) {
	dbTx := m.DB.Table(m.table).Where("id = ?", id).
		Find(&mempoolTx)
	if dbTx.Error != nil {
		logx.Errorf("get mempool tx by id error, err: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[mempool.GetMempoolTxByTxId] %s", errorcode.DbErrNotFound)
		logx.Info(err)
		return nil, errorcode.DbErrNotFound
	}
	err = m.OrderMempoolTxDetails(mempoolTx)
	if err != nil {
		logx.Errorf("get associate mempool details error, err: %s", err.Error())
		return nil, err
	}
	return mempoolTx, nil
}
