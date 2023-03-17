package tree

import (
	"fmt"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark/backend/hint"
	"math/big"
	"testing"
)

func TestGkrMimc(t *testing.T) {
	element0 := new(fr.Element).SetBigInt(big.NewInt(0))
	element1 := new(fr.Element).SetBigInt(big.NewInt(1))
	ele := hint.GMimcElements([]*fr.Element{element0, element1})
	res := ele.Bytes()
	fmt.Printf("%x\n", res)
}
