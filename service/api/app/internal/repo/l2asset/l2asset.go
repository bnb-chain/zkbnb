package l2asset

import (
	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"gorm.io/gorm"
)

var (
	cacheZecreyL2AssetInfoIdPrefix          = "cache:zecrey:l2AssetInfo:id:"
	cacheZecreyL2AssetInfoL2AssetIdPrefix   = "cache:zecrey:l2AssetInfo:l2AssetId:"
	cacheZecreyL2AssetInfoL2AssetNamePrefix = "cache:zecrey:l2AssetInfo:l2AssetName:"
)

type l2asset struct {
	cachedConn sqlc.CachedConn
	table      string
	db         *gorm.DB
	cache      multcache.MultCache
}

/*
	Func: GetL2AssetsList
	Params:
	Return: err error
	Description: create account table
*/
func (m *l2asset) GetL2AssetsList() (res []*L2AssetInfo, err error) {
	dbTx := m.db.Table(m.table).Find(&res)
	if dbTx.Error != nil {
		return nil, dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return nil, ErrNotExistInSql
	}
	for _, asset := range res {
		err := m.db.Model(&asset).Association("L1AssetsInfo").Find(&asset.L1AssetsInfo)
		if err != nil {
			return nil, err
		}
	}
	return res, nil
}

/*
	Func: GetL2AssetInfoBySymbol
	Params: symbol string
	Return: res *L2AssetInfo, err error
	Description: get l2 asset info by l2 symbol
*/
func (m *l2asset) GetL2AssetInfoBySymbol(symbol string) (res *L2AssetInfo, err error) {
	dbTx := m.db.Table(m.table).Where("l2_symbol = ?", symbol).Find(&res)
	if dbTx.Error != nil {
		return nil, dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return nil, ErrNotExistInSql
	}
	return res, nil
}

/*
	Func: GetL2AssetInfoByAssetId
	Params: assetId uint32
	Return: L2AssetInfo, error
	Description: get layer-2 asset info by assetId
*/
func (m *l2asset) GetL2AssetInfoByAssetId(assetId uint32) (res *L2AssetInfo, err error) {
	var L2AssetInfoForeignKeyColumn = "L1AssetsInfo"
	dbTx := m.db.Table(m.table).Where("l2_asset_id = ?", assetId).Find(&res)
	if dbTx.Error != nil {
		return nil, dbTx.Error
	}
	if dbTx.RowsAffected == 0 {
		return nil, ErrNotExistInSql
	}
	err = m.db.Model(&res).Association(L2AssetInfoForeignKeyColumn).Find(&res.L1AssetsInfo)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// func (*l2asset) TableName() string {
// 	return `l2_asset_info`
// }

// /*
// 	Func: CreateL2AssetInfoTable
// 	Params:
// 	Return: err error
// 	Description: create l2 asset info table
// */
// func (m *defaultL2AssetInfoModel) CreateL2AssetInfoTable() error {
// 	return m.DB.AutoMigrate(L2AssetInfo{})
// }

// /*
// 	Func: DropL2AssetInfoTable
// 	Params:
// 	Return: err error
// 	Description: drop l2 asset info table
// */
// func (m *defaultL2AssetInfoModel) DropL2AssetInfoTable() error {
// 	return m.DB.Migrator().DropTable(m.table)
// }

// /*
// 	Func: GetL2AssetsListWithoutL1AssetsInfo
// 	Params:
// 	Return: err error
// 	Description: GetL2AssetsListWithoutL1AssetsInfo
// */
// func (m *defaultL2AssetInfoModel) GetL2AssetsListWithoutL1AssetsInfo() (res []*L2AssetInfo, err error) {
// 	dbTx := m.DB.Table(m.table).Find(&res)
// 	if dbTx.Error != nil {
// 		err := fmt.Sprintf("[l2asset.GetL2AssetsList] %s", dbTx.Error)
// 		logx.Error(err)
// 		return nil, dbTx.Error
// 	}
// 	if dbTx.RowsAffected == 0 {
// 		err := fmt.Sprintf("[l2asset.GetL2AssetsList] %s", ErrNotFound)
// 		logx.Error(err)
// 		return nil, ErrNotFound
// 	}
// 	return res, nil
// }

// /*
// 	Func: CreateL2AssetInfo
// 	Params: l2AssetInfo *L2AssetInfo
// 	Return: bool, error
// 	Description: create L2AssetsInfo batches
// */
// func (m *defaultL2AssetInfoModel) CreateL2AssetInfo(l2AssetInfo *L2AssetInfo) (bool, error) {
// 	dbTx := m.DB.Table(m.table).Create(l2AssetInfo)
// 	if dbTx.Error != nil {
// 		err := fmt.Sprintf("[l2asset.CreateL2AssetInfo] %s", dbTx.Error)
// 		logx.Error(err)
// 		return false, dbTx.Error
// 	}
// 	if dbTx.RowsAffected == 0 {
// 		err := fmt.Sprintf("[l2asset.CreateL2AssetInfo] %s", ErrInvalidL2AssetInput)
// 		logx.Error(err)
// 		return false, ErrInvalidL2AssetInput
// 	}
// 	return true, nil
// }

// /*
// 	Func: CreateL2AssetsInfoInBatches
// 	Params: []*L2AssetInfo
// 	Return: rowsAffected int64, err error
// 	Description: create L2AssetsInfo batches
// */
// func (m *defaultL2AssetInfoModel) CreateL2AssetsInfoInBatches(l2AssetsInfo []*L2AssetInfo) (rowsAffected int64, err error) {
// 	dbTx := m.DB.Table(m.table).CreateInBatches(l2AssetsInfo, len(l2AssetsInfo))
// 	if dbTx.Error != nil {
// 		err := fmt.Sprintf("[l2asset.CreateL2AssetsInfoInBatches] %s", dbTx.Error)
// 		logx.Error(err)
// 		return 0, dbTx.Error
// 	}
// 	if dbTx.RowsAffected == 0 {
// 		return 0, nil
// 	}
// 	return dbTx.RowsAffected, nil
// }

// /*
// 	Func: GetL2AssetsCount
// 	Params:
// 	Return: latestHeight int64, err error
// 	Description: get latest l1asset id to active accounts
// */
// func (m *defaultL2AssetInfoModel) GetL2AssetsCount() (latestHeight int64, err error) {
// 	var asset *L2AssetInfo
// 	dbTx := m.DB.Table(m.table).Order("l2_asset_id desc").First(&asset)
// 	if dbTx.Error != nil {
// 		err := fmt.Sprintf("[l2asset.GetL2AssetsCount] %s", dbTx.Error)
// 		logx.Error(err)
// 		return 0, dbTx.Error
// 	}
// 	if dbTx.RowsAffected == 0 {
// 		err := fmt.Sprintf("[l2asset.GetL2AssetsCount] %s", ErrNotFound)
// 		logx.Error(err)
// 		return 0, ErrNotFound
// 	}
// 	return asset.L2AssetId + 1, nil
// }

// /*
// 	Func: GetL2AssetIdByChainIdAndAssetId
// 	Params:chainId uint8, assetId uint32
// 	Return: l2AssetId int64, err error
// 	Description: get layer-2 l1asset id by chainId and l1asset id, which will be used for deposit tx
// */
// func (m *defaultL2AssetInfoModel) GetL2AssetIdByChainIdAndAssetId(chainId uint8, assetId uint32) (l2AssetId int64, err error) {
// 	var asset *l1asset.L1AssetInfo
// 	dbTx := m.DB.Table(asset.TableName()).Where("chain_id = ? AND asset_id = ?", chainId, assetId).Find(&asset)
// 	if dbTx.Error != nil {
// 		err := fmt.Sprintf("[l2asset.GetL2AssetIdByChainIdAndAssetId] %s", dbTx.Error)
// 		logx.Error(err)
// 		return 0, dbTx.Error
// 	}
// 	if dbTx.RowsAffected == 0 {
// 		err := fmt.Sprintf("[l2asset.GetL2AssetIdByChainIdAndAssetId] %s", ErrNotFound)
// 		logx.Error(err)
// 		return 0, ErrNotFound
// 	}
// 	var l2Asset *L2AssetInfo
// 	dbTx = m.DB.Table(m.table).Where("id = ?", asset.L2AssetPk).Find(&l2Asset)
// 	if dbTx.Error != nil {
// 		err := fmt.Sprintf("[l2asset.GetL2AssetIdByChainIdAndAssetId] %s", dbTx.Error)
// 		logx.Error(err)
// 		return 0, dbTx.Error
// 	}
// 	if dbTx.RowsAffected == 0 {
// 		err := fmt.Sprintf("[l2asset.GetL2AssetIdByChainIdAndAssetId] %s", ErrNotFound)
// 		logx.Error(err)
// 		return 0, ErrNotFound
// 	}
// 	return l2Asset.L2AssetId, nil
// }

// /*
// 	Func: GetSimpleL2AssetInfoByAssetId
// 	Params: assetId uint32
// 	Return: L2AssetInfo, error
// 	Description: get layer-2 asset info by assetId
// */
// func (m *defaultL2AssetInfoModel) GetSimpleL2AssetInfoByAssetId(assetId uint32) (res *L2AssetInfo, err error) {
// 	dbTx := m.DB.Table(m.table).Where("l2_asset_id = ?", assetId).Find(&res)
// 	if dbTx.Error != nil {
// 		errInfo := fmt.Sprintf("[l2asset.GetL2AssetInfoByAssetId] %s", dbTx.Error)
// 		logx.Error(errInfo)
// 		return nil, dbTx.Error
// 	}
// 	if dbTx.RowsAffected == 0 {
// 		errInfo := fmt.Sprintf("[l2asset.GetL2AssetInfoByAssetId] %s", ErrNotFound)
// 		logx.Error(errInfo)
// 		return nil, ErrNotFound
// 	}
// 	return res, nil
// }

// /*
// 	Func: GetAssetIdCount
// 	Params:
// 	Return: res int64, err error
// 	Description: get l2 asset id count
// */
// func (m *defaultL2AssetInfoModel) GetAssetIdCount() (res int64, err error) {
// 	dbTx := m.DB.Table(m.table).Where("deleted_at is NULL").Count(&res)
// 	if dbTx.Error != nil {
// 		errInfo := fmt.Sprintf("[l2asset.GetAssetIdCount] %s", dbTx.Error)
// 		logx.Error(errInfo)
// 		// TODO : to be modified
// 		return 0, dbTx.Error
// 	} else {
// 		return res, nil
// 	}
// }
