/*
 * Copyright © 2021 Zkbas Protocol
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package util

import (
	"bytes"
	"errors"
	"math/big"
	"strings"
	"unicode"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func LowerCase(s string) string {
	return strings.ToLower(s)
}

func OmitSpace(s string) string {
	return strings.TrimSpace(s)
}

func OmitSpaceMiddle(s string) (rs string) {
	for _, v := range strings.FieldsFunc(s, unicode.IsSpace) {
		rs = rs + v
	}
	return rs
}

func CleanAccountName(name string) string {
	name = LowerCase(name)
	name = OmitSpace(name)
	name = OmitSpaceMiddle(name)
	return name
}

func SerializeAccountName(a []byte) string {
	return string(bytes.Trim(a[:], "\x00")) + ".legend"
}

func keccakHash(value []byte) []byte {
	hashVal := crypto.Keccak256Hash(value)
	return hashVal[:]
}

func AccountNameHash(accountName string) (res string, err error) {
	// TODO Keccak256
	words := strings.Split(accountName, ".")
	if len(words) != 2 {
		return "", errors.New("[AccountNameHash] invalid account name")
	}

	q, _ := big.NewInt(0).SetString("21888242871839275222246405745257275088548364400416034343698204186575808495617", 10)

	rootNode := make([]byte, 32)
	hashOfBaseNode := keccakHash(append(rootNode, keccakHash([]byte(words[1]))...))

	baseNode := big.NewInt(0).Mod(big.NewInt(0).SetBytes(hashOfBaseNode), q)
	baseNodeBytes := make([]byte, 32)
	baseNode.FillBytes(baseNodeBytes)

	nameHash := keccakHash([]byte(words[0]))
	subNameHash := keccakHash(append(baseNodeBytes, nameHash...))

	subNode := big.NewInt(0).Mod(big.NewInt(0).SetBytes(subNameHash), q)
	subNodeBytes := make([]byte, 32)
	subNode.FillBytes(subNodeBytes)

	res = common.Bytes2Hex(subNodeBytes)
	return res, nil
}
