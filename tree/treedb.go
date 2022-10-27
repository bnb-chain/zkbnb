package tree

import (
	"encoding/json"
	"errors"
	"hash"
	"strings"
	"time"

	bsmt "github.com/bnb-chain/zkbnb-smt"
	"github.com/bnb-chain/zkbnb-smt/database"
	"github.com/bnb-chain/zkbnb-smt/database/leveldb"
	"github.com/bnb-chain/zkbnb-smt/database/memory"
	"github.com/bnb-chain/zkbnb-smt/database/redis"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/panjf2000/ants/v2"
)

const defaultBatchReloadSize = 1000

var (
	ErrUnsupportedDriver = errors.New("unsupported db driver")
)

type Driver string

type LevelDBOption struct {
	File string
	//nolint:staticcheck
	Cache int `json:",optional"`
	//nolint:staticcheck
	Handles int `json:",optional"`
}

type RedisDBOption struct {
	//nolint:staticcheck
	ClusterAddr []string `json:",optional"`
	//nolint:staticcheck
	Addr string `json:",optional"`
	// Use the specified Username to authenticate the current connection
	// with one of the connections defined in the ACL list when connecting
	// to a Redis 6.0 instance, or greater, that is using the Redis ACL system.
	//nolint:staticcheck
	Username string `json:",optional"`
	// Optional password. Must match the password specified in the
	// requirepass server configuration option (if connecting to a Redis 5.0 instance, or lower),
	// or the User Password when connecting to a Redis 6.0 instance, or greater,
	// that is using the Redis ACL system.
	//nolint:staticcheck
	Password string `json:",optional"`

	// The maximum number of retries before giving up. Command is retried
	// on network errors and MOVED/ASK redirects.
	// Default is 3 retries.
	//nolint:staticcheck
	MaxRedirects int `json:",optional"`

	// Enables read-only commands on slave nodes.
	//nolint:staticcheck
	ReadOnly bool `json:",optional"`
	// Allows routing read-only commands to the closest master or slave node.
	// It automatically enables ReadOnly.
	//nolint:staticcheck
	RouteByLatency bool `json:",optional"`
	// Allows routing read-only commands to the random master or slave node.
	// It automatically enables ReadOnly.
	//nolint:staticcheck
	RouteRandomly bool `json:",optional"`

	// Maximum number of retries before giving up.
	// Default is 3 retries; -1 (not 0) disables retries.
	//nolint:staticcheck
	MaxRetries int `json:",optional"`
	// Minimum backoff between each retry.
	// Default is 8 milliseconds; -1 disables backoff.
	//nolint:staticcheck
	MinRetryBackoff time.Duration `json:",optional"`
	// Maximum backoff between each retry.
	// Default is 512 milliseconds; -1 disables backoff.
	//nolint:staticcheck
	MaxRetryBackoff time.Duration `json:",optional"`

	// Dial timeout for establishing new connections.
	// Default is 5 seconds.
	//nolint:staticcheck
	DialTimeout time.Duration `json:",optional"`
	// Timeout for socket reads. If reached, commands will fail
	// with a timeout instead of blocking. Use value -1 for no timeout and 0 for default.
	// Default is 3 seconds.
	//nolint:staticcheck
	ReadTimeout time.Duration `json:",optional"`
	// Timeout for socket writes. If reached, commands will fail
	// with a timeout instead of blocking.
	// Default is ReadTimeout.
	//nolint:staticcheck
	WriteTimeout time.Duration `json:",optional"`

	// Type of connection pool.
	// true for FIFO pool, false for LIFO pool.
	// Note that fifo has higher overhead compared to lifo.
	//nolint:staticcheck
	PoolFIFO bool `json:",optional"`
	// Maximum number of socket connections.
	// Default is 10 connections per every available CPU as reported by runtime.GOMAXPROCS.
	//nolint:staticcheck
	PoolSize int `json:",optional"`
	// Minimum number of idle connections which is useful when establishing
	// new connection is slow.
	//nolint:staticcheck
	MinIdleConns int `json:",optional"`
	// Connection age at which client retires (closes) the connection.
	// Default is to not close aged connections.
	//nolint:staticcheck
	MaxConnAge time.Duration `json:",optional"`
	// Amount of time client waits for connection if all connections
	// are busy before returning an error.
	// Default is ReadTimeout + 1 second.
	//nolint:staticcheck
	PoolTimeout time.Duration `json:",optional"`
	// Amount of time after which client closes idle connections.
	// Should be less than server's timeout.
	// Default is 5 minutes. -1 disables idle timeout check.
	//nolint:staticcheck
	IdleTimeout time.Duration `json:",optional"`
	// Frequency of idle checks made by idle connection reaper.
	// Default is 1 minute. -1 disables idle connection reaper,
	// but idle connections are still discarded by the client
	// if IdleTimeout is set.
	//nolint:staticcheck
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

const (
	defaultTreeRoutinePoolSize = 10240
)

func NewContext(
	name string, driver Driver,
	reload bool, routineSize int,
	levelDBOption *LevelDBOption,
	redisDBOption *RedisDBOption) (*Context, error) {

	if routineSize <= 0 {
		routineSize = defaultTreeRoutinePoolSize
	}
	pool, err := ants.NewPool(routineSize)
	if err != nil {
		return nil, err
	}
	return &Context{
		Name:           name,
		Driver:         driver,
		LevelDBOption:  levelDBOption,
		RedisDBOption:  redisDBOption,
		reload:         reload,
		routinePool:    pool,
		hasher:         bsmt.NewHasherPool(func() hash.Hash { return mimc.NewMiMC() }),
		defaultOptions: []bsmt.Option{bsmt.GoRoutinePool(pool)},
	}, nil
}

type Context struct {
	Name          string
	Driver        Driver
	LevelDBOption *LevelDBOption
	RedisDBOption *RedisDBOption

	TreeDB          database.TreeDB
	defaultOptions  []bsmt.Option
	reload          bool
	batchReloadSize int
	routinePool     *ants.Pool
	hasher          *bsmt.Hasher
}

func (ctx *Context) IsLoad() bool {
	if ctx.reload {
		return true
	}
	return ctx.Driver == MemoryDB
}

func (ctx *Context) Options(blockHeight int64) []bsmt.Option {
	var opts []bsmt.Option
	for i := range ctx.defaultOptions {
		opts = append(opts, ctx.defaultOptions[i])
	}
	if ctx.Driver == MemoryDB {
		opts = append(opts, bsmt.InitializeVersion(bsmt.Version(blockHeight)))
	}
	return opts
}

func (ctx *Context) SetOptions(opts ...bsmt.Option) {
	ctx.defaultOptions = append(ctx.defaultOptions, opts...)
}

func (ctx *Context) BatchReloadSize() int {
	if ctx.batchReloadSize <= 0 {
		return defaultBatchReloadSize // default
	}

	return ctx.batchReloadSize
}

func (ctx *Context) SetBatchReloadSize(size int) {
	ctx.batchReloadSize = size
}

func (ctx *Context) RoutinePool() *ants.Pool {
	return ctx.routinePool
}

func (ctx *Context) Hasher() *bsmt.Hasher {
	return ctx.hasher
}
