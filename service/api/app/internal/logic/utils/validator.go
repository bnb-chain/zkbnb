package utils

import (
	"math"
	"regexp"
	"strings"
)

const (
	accountNamePrefixReg = "^[a-z0-9]{3,32}$"

	accountPkLength = 64

	minSymbolLength = 3

	maxAssetId = math.MaxUint32

	minAccountIndex = 0
	maxAccountIndex = math.MaxUint32
)

func ValidateAccountName(accountName string) bool {
	if !strings.Contains(accountName, ".") {
		return false
	}

	splits := strings.Split(accountName, ".")
	if len(splits) != 2 {
		return false
	}

	match, _ := regexp.MatchString(accountNamePrefixReg, splits[0])
	return match
}

func ValidateAccountPk(accountPk string) bool {
	return len(accountPk) == accountPkLength
}

func ValidateAssetId(assetId uint32) bool {
	return assetId <= maxAssetId
}

func ValidateSymbol(symbol string) bool {
	return len(symbol) >= minSymbolLength
}

func ValidatePairIndex(pairIndex uint32) bool {
	return pairIndex >= minAccountIndex && pairIndex <= maxAccountIndex
}
