package tree

import (
	"errors"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark/backend/hint"
	gkrhash "github.com/consensys/gnark/std/gkr/hash"
	"hash"
	"math/big"
)

const (
	BlockSize = fr.Bytes // BlockSize size that GMimc consumes
)

func GMIMCElements(q *big.Int, inputs []*big.Int, results []*big.Int) error {
	if len(inputs) < 2 {
		return errors.New("MIMCHash requires at least two input elements")
	}

	var err error

	// Compute the hash for the first pair of input elements
	err = hint.MIMC2Elements(q, inputs[:2], results)
	if err != nil {
		return err
	}

	// Compute the hash for the remaining input elements
	for i := 2; i < len(inputs); i++ {
		err = hint.MIMC2Elements(q, []*big.Int{results[0], inputs[i]}, results)
		if err != nil {
			return err
		}
	}

	// Copy the final hash result to the output parameter
	results[0].SetBytes(results[0].Bytes())

	return nil
}

func GMimcElements(msg []*fr.Element) *fr.Element {
	if len(msg) < 2 {
		return nil
	}

	// Convert msg []*fr.Element to inputs ...[]byte
	inputs := make([][]byte, len(msg))
	for i, e := range msg {
		res := e.Bytes()
		inputs[i] = res[:]
	}

	// Compute the hash of the inputs
	hashBytes := GMimcBytes(inputs...)

	// Convert the hash to an *fr.Element
	var hashFr fr.Element
	hashFr.SetBytes(hashBytes)
	return &hashFr
}

func GMimcBigInt(i1, i2 *big.Int) []byte {
	newState := new(fr.Element).SetBigInt(i2)
	block := new(fr.Element).SetBigInt(new(big.Int).Add(i1, i2))
	oldState := new(fr.Element).SetBigInt(i2)
	block.Sub(block, oldState)
	gkrhash.MimcPermutationInPlace(newState, *block)
	bytes := newState.Bytes()
	return bytes[:]
}

func GMimcBigInts(inputs ...*big.Int) []byte {
	if len(inputs) < 2 {
		return nil
	}

	// Compute the hash for the first pair of input elements
	hashBytes := GMimcBigInt(inputs[0], inputs[1])

	// Compute the hash for the remaining input elements
	for i := 2; i < len(inputs); i++ {
		hashBytes = GMimcBigInt(big.NewInt(0).SetBytes(hashBytes), inputs[i])
	}
	return hashBytes
}

func GMimcBytes(inputs ...[]byte) []byte {
	if len(inputs) < 2 {
		return nil
	}

	bigIntInputs := make([]*big.Int, len(inputs))
	for i, input := range inputs {
		bigIntInputs[i] = new(big.Int).SetBytes(input)
	}
	return GMimcBigInts(bigIntInputs...)
}

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
	e.SetBigInt(new(big.Int).SetBytes(GMimcBytes(d.data...)))
	d.h = e
	d.data = nil // flush the data already hashed
	hash := d.h.Bytes()
	b = append(b, hash[:]...)
	return b
}
