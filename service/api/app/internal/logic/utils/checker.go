package utils

import (
	"math"
	"strings"
)

const (
	maxAccountNameLength          = 30
	maxAccountNameLengthOmitSpace = 20

	accountPkLength = 64

	minSymbolLength = 3
	maxSymbolLength = 8

	maxAssetId = math.MaxUint32

	minAccountIndex = 0
	maxAccountIndex = math.MaxUint32
)

func CheckAccountName(accountName string) bool {
	return len(accountName) > maxAccountNameLength
}

func CheckFormatAccountName(accountName string) bool {
	return len(accountName) > maxAccountNameLengthOmitSpace
}

func CheckAccountPk(accountPk string) bool {
	return len(accountPk) != accountPkLength
}

func CheckAssetId(assetId uint32) bool {
	return assetId > maxAssetId
}

func CheckSymbol(symbol string) bool {
	return len(symbol) > maxSymbolLength || len(symbol) < minSymbolLength
}

func CheckPairIndex(pairIndex uint32) bool {
	return pairIndex > maxAccountIndex || pairIndex < minAccountIndex
}

func FormatAccountName(name string) string {
	name = strings.ToLower(name)
	name = strings.Replace(name, "\n", "", -1)
	return name
}
