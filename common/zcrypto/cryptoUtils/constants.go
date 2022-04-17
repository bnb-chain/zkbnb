package cryptoUtils

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"sync"
)

var (
	h     = mimc.NewMiMC()
	hLock sync.Mutex
)
