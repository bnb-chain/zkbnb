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
		"ZecreyLegendContract",
	}
	ErrBlockNumOutOfRange = errors.New("[info]Block Number is out of range")
	ErrNoInfoMatched      = errors.New("[info]No Info Matched")
)
