package tree

import (
	"github.com/bnb-chain/zkbnb-crypto/wasm/txtypes"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestMIMCFrElements(t *testing.T) {
	inputs := []*big.Int{big.NewInt(0), big.NewInt(1), big.NewInt(2)}
	resEle := GMimcElements([]*fr.Element{txtypes.FromBigIntToFr(inputs[0]), txtypes.FromBigIntToFr(inputs[1]), txtypes.FromBigIntToFr(inputs[2])})
	res := resEle.Bytes()
	t.Logf("GMimcElements:%x\n", res[:])

	resBytes := GMimcBytes(inputs[0].Bytes(), inputs[1].Bytes(), inputs[2].Bytes())
	t.Logf("GMimcBytes:%x\n", resBytes)
	assert.Equal(t, res[:], resBytes)
}
