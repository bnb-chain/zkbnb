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

package util

import (
	"errors"
	"reflect"
)

var (
	ErrReflectInvalid     = errors.New("reflect invalid")
	ErrReflectIntInvalid  = errors.New("reflect invalid int type")
	ErrReflectTypeInvalid = errors.New("reflect invalid type")

	ErrAccountIndexInvalid = errors.New("invalid account index")
	ErrAssetIdInvalid      = errors.New("invalid asset id")
	ErrAccountNameInvalid  = errors.New("invalid account name")
	ErrAccountPkInvalid    = errors.New("invalid account publickey")
	ErrChainIdInvalid      = errors.New("invalid chain id")
	ErrPairIndexInvalid    = errors.New("invalid pair index")
	ErrLimitInvalid        = errors.New("invalid limit")
	ErrOffsetInvalid       = errors.New("invalid offset")
	ErrHashInvalid         = errors.New("invalid hash or commitment")
	ErrBlockHeightInvalid  = errors.New("invalid block height")
	ErrTxInvalid           = errors.New("invalid txVerification")
	ErrLPAmountInvalid     = errors.New("invalid LP amount")
	ErrAssetAmountInvalid  = errors.New("invalid asset amount")
	ErrBooleanInvalid      = errors.New("invalid boolean")
	ErrGasFeeInvalid       = errors.New("invalid gas fee")
)

func CheckRequestParam(dataType uint8, v reflect.Value) error {
	var (
		dataUint   uint64
		dataInt    int64
		dataString string
		dataBool   bool
	)
	switch v.Kind() {
	case reflect.Invalid:
		return ErrReflectInvalid
	case reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64:
		dataInt = v.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		dataUint = v.Uint()
	case reflect.String:
		dataString = v.String()
	case reflect.Bool:
		dataBool = v.Bool()
	default: // reflect.Array, reflect.Struct, reflect.Interface
		return ErrReflectTypeInvalid
	}

	switch dataType {
	case TypeAccountIndex:
		if dataUint > maxAccountIndex {
			return ErrAccountIndexInvalid
		}
		break
	case TypeAssetId:
		if dataInt > maxAssetId {
			return ErrAssetIdInvalid
		}
		break
	case TypeAccountName:
		if len(dataString) > maxAccountNameLength {
			return ErrAccountNameInvalid
		}
		break
	case TypeAccountNameOmitSpace:
		if len(dataString) > maxAccountNameLengthOmitSpace {
			return ErrAccountNameInvalid
		}
		break
	case TypeAccountPk:
		if len(dataString) > maxPublicKeyLength {
			return ErrAccountPkInvalid
		}
		break
	case TypePairIndex:
		if dataUint > maxPairIndex {
			return ErrPairIndexInvalid
		}
		break
	case TypeLimit:
		if dataUint > maxLimit {
			return ErrLimitInvalid
		}
		break
	case TypeOffset:
		if dataUint > maxOffset {
			return ErrOffsetInvalid
		}
		break
	case TypeHash:
		if len(dataString) > maxHashLength {
			return ErrHashInvalid
		}
		break
	case TypeBlockHeight:
		if dataUint > maxBlockHeight {
			return ErrBlockHeightInvalid
		}
		break
	case TypeTxType:
		if dataUint > maxTxtype {
			return ErrTxInvalid
		}
		break
	case TypeLPAmount:
		if dataUint > maxLPAmount {
			return ErrLPAmountInvalid
		}
		break
	case TypeAssetAmount:
		if dataUint > maxAssetAmount {
			return ErrAssetAmountInvalid
		}
	case TypeBoolean:
		if dataBool != true && dataBool != false {
			return ErrBooleanInvalid
		}
	case TypeGasFee:
		if dataUint > maxGasFee {
			return ErrGasFeeInvalid
		}
	}

	return nil
}
