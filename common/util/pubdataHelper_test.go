/*
 * Copyright © 2021 Zecrey Protocol
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
	"fmt"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/ethereum/go-ethereum/common"
	curve "github.com/zecrey-labs/zecrey-crypto/ecc/ztwistededwards/tebn254"
	"github.com/zecrey-labs/zecrey-legend/common/model/basic"
	"github.com/zecrey-labs/zecrey-legend/common/model/mempool"
	"log"
	"math/big"
	"testing"
)

var (
	mempoolModel = mempool.NewMempoolModel(basic.Connection, basic.CacheConf, basic.DB)
)

func TestConvertTxToRegisterZNSPubData(t *testing.T) {
	txInfo, err := mempoolModel.GetMempoolTxByTxId(1)
	if err != nil {
		t.Fatal(err)
	}
	pubData, err := ConvertTxToRegisterZNSPubData(txInfo)
	if err != nil {
		t.Fatal(err)
	}
	log.Println(common.Bytes2Hex(pubData))
}

func TestPubDataComputation(t *testing.T) {

	oldStateRoot, _ := new(big.Int).SetString("15043264495212376832665268192414242291394558777525090122806455607283976407362", 10)
	newStateRoot, _ := new(big.Int).SetString("12297177442334280409244260380119123763383333089941937037498066619533324895858", 10)
	pubData1, _ := new(big.Int).SetString("452312848688578680041881346888105167735506309919053548679680298785221640192", 10)
	pubData2, _ := new(big.Int).SetString("2983915526511271752764711863570950202204664407512986633745743434343302299646", 10)
	pubData3, _ := new(big.Int).SetString("4651953279359961953883431774279468060228027927824100750635971402026910833027", 10)
	pubData4, _ := new(big.Int).SetString("19965822911226554215779061565234345854457524215316903843664120688444160684410", 10)
	pubData5, _ := new(big.Int).SetString("8734016109108763008334396672504977758060680100901855772709016788881531390238", 10)
	pubData6, _ := new(big.Int).SetString("0", 10)

	fmt.Println(common.Bytes2Hex(oldStateRoot.FillBytes(make([]byte, 32))))
	fmt.Println(common.Bytes2Hex(newStateRoot.FillBytes(make([]byte, 32))))

	var buf bytes.Buffer
	buf.Write(pubData1.FillBytes(make([]byte, 32)))
	buf.Write(pubData2.FillBytes(make([]byte, 32)))
	buf.Write(pubData3.FillBytes(make([]byte, 32)))
	buf.Write(pubData4.FillBytes(make([]byte, 32)))
	buf.Write(pubData5.FillBytes(make([]byte, 32)))
	buf.Write(pubData6.FillBytes(make([]byte, 32)))
	fmt.Println(common.Bytes2Hex(buf.Bytes()))

	commitment, _ := new(big.Int).SetString("2001904096268940627870837110796902048094724981649621731904769518577628368633", 10)
	fmt.Println(common.Bytes2Hex(commitment.FillBytes(make([]byte, 32))))
}

func TestPubData2(t *testing.T) {
	var buf bytes.Buffer
	//bytesType, _ := abi.NewType("bytes", "", nil)
	//uint32Type, _ := abi.NewType("uint32", "", nil)
	//uint64Type, _ := abi.NewType("uint64", "", nil)
	//bytes32Type, _ := abi.NewType("bytes32", "", nil)
	buf.Write(new(big.Int).SetInt64(1).FillBytes(make([]byte, 32)))
	buf.Write(new(big.Int).SetInt64(1654843322039).FillBytes(make([]byte, 32)))
	buf.Write(common.FromHex("14e4e8ad4848558d7200530337052e1ad30f5385b3c7187c80ad85f48547b74f"))
	buf.Write(common.FromHex("21422f9bebac15af8ddc504da0dbb88020c1a4de7e7b6722fe00acb0ed968942"))
	buf.Write(common.FromHex("01000000000000000000000000000000000000000000000000000000000000007472656173757279000000000000000000000000000000000000000000000000167c5363088a40a4839912a872f43164270740c7e986ec55397b2d583317ab4a2005db7af2bdcfae1fa8d28833ae2f1995e9a8e0825377cff121db64b0db21b718a96ca582a72b16f464330c89ab73277cb96e42df105ebf5c9ac5330d47b8fc0000000000000000000000000000000000000000000000000000000000000000"))
	buf.Write(new(big.Int).SetInt64(1).FillBytes(make([]byte, 32)))
	//hFunc.Write(buf.Bytes())
	//hashVal := hFunc.Sum(nil)
	hFunc := mimc.NewMiMC()
	//arguments := abi.Arguments{{Type: uint64Type}, {Type: uint64Type}, {Type: bytes32Type}, {Type: bytes32Type}, {Type: bytesType}, {Type: uint32Type}}
	//info, _ := arguments.Pack(
	//	uint64(1),
	//	uint64(1654843322039),
	//	common.FromHex("14e4e8ad4848558d7200530337052e1ad30f5385b3c7187c80ad85f48547b74f"),
	//	common.FromHex("21422f9bebac15af8ddc504da0dbb88020c1a4de7e7b6722fe00acb0ed968942"),
	//	common.FromHex("01000000000000000000000000000000000000000000000000000000000000007472656173757279000000000000000000000000000000000000000000000000167c5363088a40a4839912a872f43164270740c7e986ec55397b2d583317ab4a2005db7af2bdcfae1fa8d28833ae2f1995e9a8e0825377cff121db64b0db21b718a96ca582a72b16f464330c89ab73277cb96e42df105ebf5c9ac5330d47b8fc0000000000000000000000000000000000000000000000000000000000000000"),
	//	uint32(1),
	//)
	//log.Println(common.Bytes2Hex(info))
	log.Println(common.Bytes2Hex(buf.Bytes()))
	//hashVal := KeccakHash(info)
	hFunc.Write(buf.Bytes())
	hashVal := hFunc.Sum(nil)
	fmt.Println(common.Bytes2Hex(hashVal))
}

func TestMiMCHash(t *testing.T) {
	hFunc := mimc.NewMiMC()
	hFunc.Write(new(big.Int).SetInt64(123123123).FillBytes(make([]byte, 32)))
	a := hFunc.Sum(nil)
	fmt.Println(new(big.Int).SetBytes(a).String())
	fmt.Println(common.Bytes2Hex(curve.Modulus.Bytes()))
}
