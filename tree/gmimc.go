package tree

import (
	"errors"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark/backend/hint"
	"hash"
	"math/big"
)

const (
	BlockSize = fr.Bytes // BlockSize size that GMimc consumes
)

type digest struct {
	h    fr.Element
	data [][]byte // data to hash
}

func NewGMimc() hash.Hash {
	d := new(digest)
	d.Reset()
	return d
}

// Reset resets the Hash to its initial state.
func (d *digest) Reset() {
	d.data = nil
	d.h = fr.Element{0, 0, 0, 0}
}

// Only receive byte slice less than fr.Modulus()
func (d *digest) Write(p []byte) (n int, err error) {
	n = len(p)
	num := new(big.Int).SetBytes(p)
	if num.Cmp(fr.Modulus()) >= 0 {
		return 0, errors.New("not support bytes bigger than modulus")
	}
	d.data = append(d.data, p)
	return n, nil
}

func (d *digest) Size() int {
	return BlockSize
}

// BlockSize returns the number of bytes Sum will return.
func (d *digest) BlockSize() int {
	return BlockSize
}

// Sum appends the current hash to b and returns the resulting slice.
// It does not change the underlying hash state.
func (d *digest) Sum(b []byte) []byte {
	e := fr.Element{0, 0, 0, 0}
	e.SetBigInt(new(big.Int).SetBytes(hint.GMimcBytes(d.data...)))
	d.h = e
	d.data = nil // flush the data already hashed
	hash := d.h.Bytes()
	b = append(b, hash[:]...)
	return b
}
