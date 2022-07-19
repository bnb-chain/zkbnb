package treedb

import (
	"github.com/bnb-chain/bas-smt/database"
	"github.com/bnb-chain/bas-smt/database/leveldb"
	"github.com/bnb-chain/bas-smt/database/memory"
	"github.com/bnb-chain/bas-smt/database/redis"
)

const (
	NFTPrefix          = "nft:"
	LiquidityPrefix    = "liquidity:"
	AccountPrefix      = "account:"
	AccountAssetPrefix = "account_asset:"
)

type Driver string

type LevelDBOptions struct {
	File    string
	Cache   int
	Handles int
}

type RedisDBOptions = redis.RedisConfig

const (
	MemoryDB Driver = "memorydb"
	LevelDB  Driver = "leveldb"
	RedisDB  Driver = "redis"
)

func NewTreeDB(
	driver Driver,
	levelDBOptions LevelDBOptions,
	redisDBOption RedisDBOptions,
) (database.TreeDB, error) {
	switch driver {
	case MemoryDB:
		return memory.NewMemoryDB(), nil
	case LevelDB:
		return leveldb.New(levelDBOptions.File, levelDBOptions.Cache, levelDBOptions.Handles, false)
	case RedisDB:
		return redis.New(&redisDBOption)
	}
	return nil, ErrUnsupportedDriver
}

func SetNamespace(
	driver Driver,
	db database.TreeDB,
	namespace string,
) database.TreeDB {
	switch driver {
	case MemoryDB:
		return memory.NewMemoryDB()
	case LevelDB:
		return leveldb.WrapWithNamespace(db.(*leveldb.Database), namespace)
	case RedisDB:
		return redis.WrapWithNamespace(db.(*redis.Database), namespace)
	}
	return db
}
