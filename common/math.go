package common

import (
	"math"
	"math/big"
)

func MinInt64(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

func MaxInt64(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}

func GetFeeFromWei(fee *big.Int) float64 {
	fbalance := new(big.Float)
	fbalance.SetString(fee.String())
	ethValue := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(18)))
	f, _ := ethValue.Float64()
	return f
}
