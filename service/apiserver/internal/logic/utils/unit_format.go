package utils

import (
	"github.com/ethereum/go-ethereum/params"
	"math/big"
)

func FormatWeiToEther(weiAmount *big.Int) *big.Float {
	f := new(big.Float)
	f.SetPrec(236) //  IEEE 754 octuple-precision binary floating-point format: binary256
	f.SetMode(big.ToNearestEven)
	fWei := new(big.Float)
	fWei.SetPrec(236) //  IEEE 754 octuple-precision binary floating-point format: binary256
	fWei.SetMode(big.ToNearestEven)
	etherAmount := f.Quo(fWei.SetInt(weiAmount), big.NewFloat(params.Ether))
	return etherAmount
}
