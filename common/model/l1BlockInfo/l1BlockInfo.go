package l1BlockInfo

import (
	"errors"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas/errorcode"
)

type (
	L1BlockInfoModel interface {
		CreateL1BlockInfoTable() error
		DropL1BlockInfoTable() error
		CreateL1BlockInfo(blockInfo *L1BlockInfo) (bool, error)
		CreateL1BlockInfosInBatches(blockInfos []*L1BlockInfo) (rowsAffected int64, err error)
		GetL1BlockInfos() (blockInfos []*L1BlockInfo, err error)
		GetLatestL1BlockInfo() (blockInfo *L1BlockInfo, err error)
		GetL1BlockInfosByL2BlockHeight(l2BlockHeight int64) (blockInfo *L1BlockInfo, err error)
	}

	defaultL1BlockInfoModel struct {
		sqlc.CachedConn
		table string
		DB    *gorm.DB
	}

	L1BlockInfo struct {
		gorm.Model
		// block type, 1 - Commit 2 - Verify
		BlockType uint8
		// block height
		L2BlockHeight int64 `gorm:"index;not null"`
		// commit block info
		BlockInfo string // base64.StdEncoding.EncodeToString(json.Marshal(BlockCommit))
	}
)

func (*L1BlockInfo) TableName() string {
	return TableName
}

func NewL1BlockInfoModel(conn sqlx.SqlConn, c cache.CacheConf, db *gorm.DB) L1BlockInfoModel {
	return &defaultL1BlockInfoModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      TableName,
		DB:         db,
	}
}

/*
	Func: CreateL1BlockInfoTable
	Params:
	Return: err error
	Description: create l2 txVerification event monitor table
*/
func (m *defaultL1BlockInfoModel) CreateL1BlockInfoTable() error {
	return m.DB.AutoMigrate(L1BlockInfo{})
}

/*
	Func: DropL1BlockInfoTable
	Params:
	Return: err error
	Description: drop l2 txVerification event monitor table
*/
func (m *defaultL1BlockInfoModel) DropL1BlockInfoTable() error {
	return m.DB.Migrator().DropTable(m.table)
}

/*
	Func: CreateL1BlockInfo
	Params: asset *L1BlockInfo
	Return: bool, error
	Description: create L1BlockInfo txVerification
*/
func (m *defaultL1BlockInfoModel) CreateL1BlockInfo(blockInfo *L1BlockInfo) (bool, error) {
	dbTx := m.DB.Table(m.table).Create(blockInfo)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l1BlockInfo.CreateL1BlockInfo] %s", dbTx.Error)
		logx.Error(err)
		return false, dbTx.Error
	} else if dbTx.RowsAffected == 0 {
		ErrInvalidL1BlockInfo := errors.New("invalid l1BlockInfo")
		err := fmt.Sprintf("[l1BlockInfo.CreateL1BlockInfo] %s", ErrInvalidL1BlockInfo)
		logx.Error(err)
		return false, ErrInvalidL1BlockInfo
	}
	return true, nil
}

/*
	Func: CreateL1BlockInfosInBatches
	Params: []*L1BlockInfo
	Return: rowsAffected int64, err error
	Description: create L1BlockInfo batches
*/
func (m *defaultL1BlockInfoModel) CreateL1BlockInfosInBatches(blockInfos []*L1BlockInfo) (rowsAffected int64, err error) {
	dbTx := m.DB.Table(m.table).CreateInBatches(blockInfos, len(blockInfos))
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l1BlockInfo.CreateL1AssetsMonitorInBatches] %s", dbTx.Error)
		logx.Error(err)
		return 0, dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return 0, nil
	}
	return dbTx.RowsAffected, nil
}

/*
	GetL1BlockInfos: get all L1BlockInfos
*/
func (m *defaultL1BlockInfoModel) GetL1BlockInfos() (blockInfos []*L1BlockInfo, err error) {
	dbTx := m.DB.Table(m.table).Find(&blockInfos).Order("l2_block_height")
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l1BlockInfo.GetL1BlockInfos] %s", dbTx.Error)
		logx.Error(err)
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[l1BlockInfo.GetL1BlockInfos] %s", errorcode.DbErrNotFound)
		logx.Error(err)
		return nil, errorcode.DbErrNotFound
	}
	return blockInfos, dbTx.Error
}

/*
	Func: GetLatestL1BlockInfo
	Return: blockInfos []*L1BlockInfo, err error
	Description: get latest l1 block monitor info
*/
func (m *defaultL1BlockInfoModel) GetLatestL1BlockInfo() (blockInfo *L1BlockInfo, err error) {
	dbTx := m.DB.Table(m.table).Order("l2_block_height desc").First(&blockInfo)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l1BlockInfo.GetLatestL1BlockInfo] %s", dbTx.Error)
		logx.Error(err)
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[l1BlockInfo.GetLatestL1BlockInfo] %s", errorcode.DbErrNotFound)
		logx.Error(err)
		return nil, errorcode.DbErrNotFound
	}
	return blockInfo, nil
}

/*
	Func: GetL1BlockInfosByChainIdAndL1BlockHeight
	Return: blockInfos []*L1BlockInfo, err error
	Description: get L1BlockInfo by chain id and l1 block height
*/
func (m *defaultL1BlockInfoModel) GetL1BlockInfosByL2BlockHeight(l2BlockHeight int64) (blockInfo *L1BlockInfo, err error) {
	dbTx := m.DB.Table(m.table).Where("l2_block_height = ?", l2BlockHeight).First(&blockInfo)
	if dbTx.Error != nil {
		err := fmt.Sprintf("[l1BlockInfo.GetL1BlockInfosByL2BlockHeight] %s", dbTx.Error)
		logx.Error(err)
		return nil, errorcode.DbErrSqlOperation
	} else if dbTx.RowsAffected == 0 {
		err := fmt.Sprintf("[l1BlockInfo.GetL1BlockInfosByL2BlockHeight] %s", errorcode.DbErrNotFound)
		logx.Error(err)
		return nil, errorcode.DbErrNotFound
	}
	return blockInfo, nil
}
