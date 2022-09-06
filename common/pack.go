package common

import (
	"math/big"

	"github.com/bnb-chain/zkbnb-crypto/util"
)

// ToPackedAmount : convert big int to 40 bit, 5 bits for 10^x, 35 bits for a * 10^x
func ToPackedAmount(amount *big.Int) (res int64, err error) {
	return util.ToPackedAmount(amount)
}

func CleanPackedAmount(amount *big.Int) (nAmount *big.Int, err error) {
	return util.CleanPackedAmount(amount)
}

// ToPackedFee : convert big int to 16 bit, 5 bits for 10^x, 11 bits for a * 10^x
func ToPackedFee(amount *big.Int) (res int64, err error) {
	return util.ToPackedFee(amount)
}

func CleanPackedFee(amount *big.Int) (nAmount *big.Int, err error) {
	return util.CleanPackedFee(amount)
}
