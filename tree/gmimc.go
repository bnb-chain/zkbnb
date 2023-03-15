package tree

import (
	"errors"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"hash"
	"math/big"
)

const (
	BlockSize = fr.Bytes // BlockSize size that GMimc consumes
)

// GMimcT2 is a hasher for t = 2
var GMimcT2 GMimcHasher

// GMimcT4 is a hasher for t = 4
var GMimcT4 GMimcHasher

// GMimcT5 is a hasher for t = 5
var GMimcT5 GMimcHasher

// GMimcT6 is a hasher for t = 6
var GMimcT6 GMimcHasher

// GMimcT8 is a hasher for t = 8
var GMimcT8 GMimcHasher

func init() {
	initArk()
	initGMimc()
}

func initGMimc() {
	GMimcT2 = GMimcHasher{t: 2, nRounds: 91}
	GMimcT4 = GMimcHasher{t: 4, nRounds: 91}
	GMimcT5 = GMimcHasher{t: 5, nRounds: 91}
	GMimcT6 = GMimcHasher{t: 6, nRounds: 91}
	GMimcT8 = GMimcHasher{t: 8, nRounds: 91}
}

// GMimcHasher contains all the parameters to describe a GMimc function
type GMimcHasher struct {
	t       int // size of Cauchy matrix
	nRounds int // number of rounds of the Mimc hash function
}

// Hash hashes a full message
func (g *GMimcHasher) Hash(msg []*fr.Element) *fr.Element {
	state := make([]fr.Element, g.t)

	for i := 0; i < len(msg); i += g.t {
		block := make([]fr.Element, g.t)
		if i+g.t >= len(msg) {
			// Only zero-pad the input
			for j, w := range msg[i:] {
				block[j] = *w
			}
		} else {
			// Take a full chunk
			for j, w := range msg[i : i+g.t] {
				block[j] = *w
			}
		}
		g.UpdateInplace(state, block)
	}

	return &state[0]
}

func GMimcBytes(input ...[]byte) []byte {
	inputElements := make([]*fr.Element, len(input))
	for i, ele := range input {
		num := new(big.Int).SetBytes(ele)
		if num.Cmp(fr.Modulus()) >= 0 {
			panic("not support bytes bigger than modulus")
		}
		e := fr.Element{0, 0, 0, 0}
		e.SetBigInt(num)
		inputElements[i] = &e
	}

	var res [32]byte
	switch len(input) {
	case 2:
		res = GMimcT2.Hash(inputElements).Bytes()
	case 4:
		res = GMimcT4.Hash(inputElements).Bytes()
	case 5:
		res = GMimcT5.Hash(inputElements).Bytes()
	case 6:
		res = GMimcT6.Hash(inputElements).Bytes()
	case 8:
		res = GMimcT8.Hash(inputElements).Bytes()
	default:
		panic("invalid length of input to hash by GKR Mimc")
	}
	return res[:]
}

// UpdateInplace updates the state with the provided block of data
func (g *GMimcHasher) UpdateInplace(state []fr.Element, block []fr.Element) {
	oldState := append([]fr.Element{}, state...)
	for i := 0; i < g.nRounds; i++ {
		AddArkAndKeysInplace(state, block, Arks[i])
		SBoxInplace(&state[0])
		InPlaceCircularPermutation(state)
	}

	// Recombine with the old state
	for i := range state {
		state[i].Add(&state[i], &oldState[i])
		state[i].Add(&state[i], &block[i])
	}
}

// InPlaceCircularPermutation moves all the element to the left and place the first element
// at the end of the state
// ie: [a, b, c, d] -> [b, c, d, a]
func InPlaceCircularPermutation(state []fr.Element) {
	for i := 1; i < len(state); i++ {
		state[i-1], state[i] = state[i], state[i-1]
	}
}

// AddArkAndKeysInplace adds the
func AddArkAndKeysInplace(state []fr.Element, keys []fr.Element, ark fr.Element) {
	for i := range state {
		state[i].Add(&state[i], &keys[i])
		state[i].Add(&state[i], &ark)
	}
}

// SBoxInplace computes x^7 in-place
func SBoxInplace(x *fr.Element) {
	tmp := *x
	x.Square(x)
	x.Mul(x, &tmp)
	x.Square(x)
	x.Mul(x, &tmp)
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
