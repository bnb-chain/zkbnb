package nft

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"

	nftModel "github.com/zecrey-labs/zecrey-legend/common/model/nft"
	"github.com/zecrey-labs/zecrey-legend/pkg/multcache"
)

type nft struct {
	table     string
	db        *gorm.DB
	cache     multcache.MultCache
	redisConn *redis.Redis
}

func (n *nft) GetNftListByAccountIndex(ctx context.Context, accountIndex, limit, offset int64) (nfts []*nftModel.L2Nft, err error) {
	f := func() (interface{}, error) {
		nftList := make([]*nftModel.L2Nft, 0)
		dbTx := n.db.Table(n.table).Where("owner_account_index = ? and deleted_at is NULL", accountIndex).
			Limit(int(limit)).Offset(int(offset)).Order("nft_index desc").Find(&nftList)
		if dbTx.Error != nil {
			return nil, dbTx.Error
		} else if dbTx.RowsAffected == 0 {
			return nil, sqlx.ErrNotFound
		}
		return &nftList, nil
	}
	nftList := make([]*nftModel.L2Nft, 0)
	value, err := n.cache.GetWithSet(ctx, multcache.SpliceCacheKeyAccountNftList(accountIndex, offset, limit), &nftList, 1, f)
	if err != nil {
		return nil, err
	}
	nftListStored, ok := value.(*[]*nftModel.L2Nft)
	if !ok {
		return nil, fmt.Errorf("[GetNftListByAccountIndex] ErrConvertFail")
	}
	return *nftListStored, nil
}

func (n *nft) GetAccountNftTotalCount(ctx context.Context, accountIndex int64) (int64, error) {
	f := func() (interface{}, error) {
		var count int64
		dbTx := n.db.Table(n.table).Where("owner_account_index = ? and deleted_at is NULL", accountIndex).Count(&count)
		if dbTx.Error != nil {
			return 0, fmt.Errorf("[GetAccountNftTotalCount]: %v", dbTx.Error)
		} else if dbTx.RowsAffected == 0 {
			return 0, sqlx.ErrNotFound
		}
		return &count, nil
	}

	var count int64
	value, err := n.cache.GetWithSet(ctx, multcache.SpliceCacheKeyAccountTotalNftCount(accountIndex), &count, 5, f)
	if err != nil {
		return count, err
	}
	countStored, ok := value.(*int64)
	if !ok {
		return 0, fmt.Errorf("[GetAccountNftTotalCount] ErrConvertFail")
	}
	return *countStored, nil
}
