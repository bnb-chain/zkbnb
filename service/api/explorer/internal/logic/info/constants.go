package info

import "errors"

const (
	TypeError = iota
	TypeNil
	TypeBlock
	TypeTx
	TypeAccount
)

var (
	contractAddressesNames = []string{
		"Zecrey_Contract_Ethereum",
		"Zecrey_Contract_Polygon",
		"Zecrey_Contract_NEAR_Aurora",
		"Zecrey_Contract_Avalanche",
		"Zecrey_Contract_BSC",
	}
	ErrBlockNumOutOfRange = errors.New("[info]Block Number is out of range")
	ErrNoInfoMatched      = errors.New("[info]No Info Matched")
)
