package treedb

import (
	"encoding/json"
	"strings"
	"time"

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

type LevelDBOption struct {
	File    string
	Cache   int `json:",optional"`
	Handles int `json:",optional"`
}

type RedisDBOption struct {
	ClusterAddr []string `json:",optional"`
	Addr        string   `json:",optional"`
	// Use the specified Username to authenticate the current connection
	// with one of the connections defined in the ACL list when connecting
	// to a Redis 6.0 instance, or greater, that is using the Redis ACL system.
	Username string `json:",optional"`
	// Optional password. Must match the password specified in the
	// requirepass server configuration option (if connecting to a Redis 5.0 instance, or lower),
	// or the User Password when connecting to a Redis 6.0 instance, or greater,
	// that is using the Redis ACL system.
	Password string `json:",optional"`

	// The maximum number of retries before giving up. Command is retried
	// on network errors and MOVED/ASK redirects.
	// Default is 3 retries.
	MaxRedirects int `json:",optional"`

	// Enables read-only commands on slave nodes.
	ReadOnly bool `json:",optional"`
	// Allows routing read-only commands to the closest master or slave node.
	// It automatically enables ReadOnly.
	RouteByLatency bool `json:",optional"`
	// Allows routing read-only commands to the random master or slave node.
	// It automatically enables ReadOnly.
	RouteRandomly bool `json:",optional"`

	// Maximum number of retries before giving up.
	// Default is 3 retries; -1 (not 0) disables retries.
	MaxRetries int `json:",optional"`
	// Minimum backoff between each retry.
	// Default is 8 milliseconds; -1 disables backoff.
	MinRetryBackoff time.Duration `json:",optional"`
	// Maximum backoff between each retry.
	// Default is 512 milliseconds; -1 disables backoff.
	MaxRetryBackoff time.Duration `json:",optional"`

	// Dial timeout for establishing new connections.
	// Default is 5 seconds.
	DialTimeout time.Duration `json:",optional"`
	// Timeout for socket reads. If reached, commands will fail
	// with a timeout instead of blocking. Use value -1 for no timeout and 0 for default.
	// Default is 3 seconds.
	ReadTimeout time.Duration `json:",optional"`
	// Timeout for socket writes. If reached, commands will fail
	// with a timeout instead of blocking.
	// Default is ReadTimeout.
	WriteTimeout time.Duration `json:",optional"`

	// Type of connection pool.
	// true for FIFO pool, false for LIFO pool.
	// Note that fifo has higher overhead compared to lifo.
	PoolFIFO bool `json:",optional"`
	// Maximum number of socket connections.
	// Default is 10 connections per every available CPU as reported by runtime.GOMAXPROCS.
	PoolSize int `json:",optional"`
	// Minimum number of idle connections which is useful when establishing
	// new connection is slow.
	MinIdleConns int `json:",optional"`
	// Connection age at which client retires (closes) the connection.
	// Default is to not close aged connections.
	MaxConnAge time.Duration `json:",optional"`
	// Amount of time client waits for connection if all connections
	// are busy before returning an error.
	// Default is ReadTimeout + 1 second.
	PoolTimeout time.Duration `json:",optional"`
	// Amount of time after which client closes idle connections.
	// Should be less than server's timeout.
	// Default is 5 minutes. -1 disables idle timeout check.
	IdleTimeout time.Duration `json:",optional"`
	// Frequency of idle checks made by idle connections reaper.
	// Default is 1 minute. -1 disables idle connections reaper,
	// but idle connections are still discarded by the client
	// if IdleTimeout is set.
	IdleCheckFrequency time.Duration `json:",optional"`
}

const (
	MemoryDB Driver = "memorydb"
	LevelDB  Driver = "leveldb"
	RedisDB  Driver = "redis"
)

func SetupTreeDB(
	context *Context,
) error {
	switch context.Driver {
	case MemoryDB:
		context.TreeDB = memory.NewMemoryDB()
		return nil
	case LevelDB:
		db, err := leveldb.New(context.LevelDBOption.File, context.LevelDBOption.Cache, context.LevelDBOption.Handles, false)
		if err != nil {
			return err
		}
		context.TreeDB = db
		return nil
	case RedisDB:
		bytes, err := json.Marshal(context.RedisDBOption)
		if err != nil {
			return err
		}
		redisOption := &redis.RedisConfig{}
		err = json.Unmarshal(bytes, redisOption)
		if err != nil {
			return err
		}
		db, err := redis.New(redisOption)
		if err != nil {
			return err
		}
		context.TreeDB = db
		return nil
	}
	return ErrUnsupportedDriver
}

func SetNamespace(
	context *Context,
	namespace string,
) database.TreeDB {
	switch context.Driver {
	case MemoryDB:
		return memory.NewMemoryDB()
	case LevelDB:
		return leveldb.WrapWithNamespace(context.TreeDB.(*leveldb.Database), strings.Join([]string{context.Name, namespace}, ":"))
	case RedisDB:
		return redis.WrapWithNamespace(context.TreeDB.(*redis.Database), strings.Join([]string{context.Name, namespace}, ":"))
	}
	return context.TreeDB
}
