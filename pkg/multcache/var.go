package multcache

import (
	"errors"
)

// error got from other package
var (
	errRedisCacheKeyNotExist = errors.New("redis: nil")
	errGoCacheKeyNotExist    = errors.New("Value not found in GoCache store")
)

// cache key register
const (
	// account
	KeyAccountAccountName = "cache:account_accountName"
)
