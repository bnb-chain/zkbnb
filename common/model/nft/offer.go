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

package nft

import (
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/errorcode"
)

type (
	OfferModel interface {
		CreateOfferTable() error
		DropOfferTable() error
		GetOfferByAccountIndexAndOfferId(accountIndex int64, offerId int64) (offer *Offer, err error)
		GetLatestOfferId(accountIndex int64) (offerId int64, err error)
		CreateOffer(offer *Offer) (err error)
	}
	defaultOfferModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	Offer struct {
		gorm.Model
		OfferType    int64
		OfferId      int64
		AccountIndex int64
		NftIndex     int64
		AssetId      int64
		AssetAmount  string
		ListedAt     int64
		ExpiredAt    int64
		TreasuryRate int64
		Sig          string
		Status       int
	}
)

func NewOfferModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) OfferModel {
	return &defaultOfferModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      OfferTableName,
		DB:         db,
	}
}

func (*Offer) TableName() string {
	return OfferTableName
}

/*
	Func: CreateOfferTable
	Params:
	Return: err error
	Description: create account l2 nft table
*/
func (m *defaultOfferModel) CreateOfferTable() error {
	return m.DB.AutoMigrate(Offer{})
}

/*
	Func: DropOfferTable
	Params:
	Return: err error
	Description: drop account l2 nft history table
*/
func (m *defaultOfferModel) DropOfferTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

func (m *defaultOfferModel) GetLatestOfferId(accountIndex int64) (offerId int64, err error) {
	var offer *Offer
	dbTx := m.DB.Table(m.table).Where("account_index = ?", accountIndex).Order("offer_id desc").Find(&offer)
	if dbTx.Error != nil {
		logx.Errorf("[GetLatestOfferId] unable to get latest offer info: %s", dbTx.Error.Error())
		return -1, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		return -1, errorcode.DbErrNotFound
	}
	return offer.OfferId, nil
}

func (m *defaultOfferModel) CreateOffer(offer *Offer) (err error) {
	dbTx := m.DB.Table(m.table).Create(offer)
	if dbTx.Error != nil {
		logx.Errorf("[CreateOffer] unable to create offer: %s", dbTx.Error.Error())
		return dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		logx.Errorf("[CreateOffer] invalid offer info")
		return errors.New("[CreateOffer] invalid offer info")
	}
	return nil
}

func (m *defaultOfferModel) GetOfferByAccountIndexAndOfferId(accountIndex int64, offerId int64) (offer *Offer, err error) {
	dbTx := m.DB.Table(m.table).Where("account_index = ? AND offer_id = ?", accountIndex, offerId).Find(&offer)
	if dbTx.Error != nil {
		logx.Errorf("[CreateOffer] unable to create offer: %s", dbTx.Error.Error())
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		logx.Errorf("[CreateOffer] invalid offer info")
		return nil, errorcode.DbErrNotFound
	}
	return offer, nil
}
