package l2asset

import (
	"errors"

	"github.com/zecrey-labs/zecrey-legend/pkg/zerror"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	ErrNotFound            = sqlx.ErrNotFound
	ErrInvalidL2AssetInput = errors.New("[ErrInvalidL2AssetInput] Invalid L2AssetInfo input")
)

var (
	ErrNewHttpRequest = zerror.New(40000, "http.NewRequest err")
	ErrHttpClientDo   = zerror.New(40001, "http.Client.Do err")
	ErrIoutilReadAll  = zerror.New(40002, "ioutil.ReadAll err")
	ErrJsonUnmarshal  = zerror.New(40003, "json.Unmarshal err")
	ErrJsonMarshal    = zerror.New(40004, "json.Marshal err")
	ErrTypeAssertion  = zerror.New(40005, "type assertion err")
	ErrSetCache       = zerror.New(40006, "set cache err")
)
