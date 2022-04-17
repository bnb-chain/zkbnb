package cryptoUtils

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"github.com/google/uuid"
	curve "github.com/zecrey-labs/zecrey-crypto/ecc/ztwistededwards/tebn254"
	"math/big"
)

/*
	ParseSkStr: parse private key
*/
func ParseSkStr(skStr string) (sk *big.Int, err error) {
	sk, b := new(big.Int).SetString(skStr, 10)
	if !b {
		return nil, ErrInvalidSkStr
	}
	return sk, nil
}

func Base64Encode(buf []byte) string {
	return base64.StdEncoding.EncodeToString(buf)
}

func Base64Decode(bufStr string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(bufStr)
}

/*
	ComputeHashVal: compute hash val
*/
func ComputeHashVal(buf []byte) string {
	hLock.Lock()
	defer hLock.Unlock()
	h.Reset()
	h.Write(buf)
	return hex.EncodeToString(h.Sum([]byte{}))
}

/*
	Uint32ToBytes: uint32 to bytes
*/
func Uint32ToBytes(a uint32) []byte {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], a)
	return buf[:]
}

/*
	Uint64ToBytes: uint64 to bytes
*/
func Uint64ToBytes(a uint64) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], a)
	return buf[:]
}

/*
	GetRandomUUID: get random value
*/
func GetRandomUUID() string {
	u := uuid.New()
	return u.String()
}

func IsValidPublicKey(pkStr string) (bool, error) {
	pk, err := curve.FromString(pkStr)
	if err != nil {
		return false, err
	}
	isInSubGroup := curve.IsInSubGroup(pk)
	return isInSubGroup, nil
}
