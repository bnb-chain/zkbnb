/*
 * Copyright © 2021 ZkBAS Protocol
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

package chain

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/ethereum/go-ethereum/common"

	curve "github.com/bnb-chain/zkbas-crypto/ecc/ztwistededwards/tebn254"
	"github.com/bnb-chain/zkbas-crypto/ffmath"
	common2 "github.com/bnb-chain/zkbas/common"
)

func TestPubDataComputation(t *testing.T) {

	oldStateRoot, _ := new(big.Int).SetString("15043264495212376832665268192414242291394558777525090122806455607283976407362", 10)
	newStateRoot, _ := new(big.Int).SetString("12297177442334280409244260380119123763383333089941937037498066619533324895858", 10)
	pubData1, _ := new(big.Int).SetString("452312848688578680041881346888105167735506309919053548679680298785221640192", 10)
	pubData2, _ := new(big.Int).SetString("2983915526511271752764711863570950202204664407512986633745743434343302299646", 10)
	pubData3, _ := new(big.Int).SetString("4651953279359961953883431774279468060228027927824100750635971402026910833027", 10)
	pubData4, _ := new(big.Int).SetString("19965822911226554215779061565234345854457524215316903843664120688444160684410", 10)
	pubData5, _ := new(big.Int).SetString("8734016109108763008334396672504977758060680100901855772709016788881531390238", 10)
	pubData6, _ := new(big.Int).SetString("0", 10)

	assert.Equal(t, common.Bytes2Hex(oldStateRoot.FillBytes(make([]byte, 32))), "21422f9bebac15af8ddc504da0dbb88020c1a4de7e7b6722fe00acb0ed968942")
	assert.Equal(t, common.Bytes2Hex(newStateRoot.FillBytes(make([]byte, 32))), "1b2ff4ae0d507a971fb267849af6a28000b1d483865c5a610cc47db6f196c672")

	var buf bytes.Buffer
	buf.Write(pubData1.FillBytes(make([]byte, 32)))
	buf.Write(pubData2.FillBytes(make([]byte, 32)))
	buf.Write(pubData3.FillBytes(make([]byte, 32)))
	buf.Write(pubData4.FillBytes(make([]byte, 32)))
	buf.Write(pubData5.FillBytes(make([]byte, 32)))
	buf.Write(pubData6.FillBytes(make([]byte, 32)))
	assert.Equal(t, common.Bytes2Hex(buf.Bytes()), "01000000010000000000000000000000000000000000000000000000000000000698d61a3d9cbfac8f5f7492fcfd4f45af982f6f0c8d1edd783c14d81ffffffe0a48e9892a45a04d0c5b0f235a3aeb07b92137ba71a59b9c457774bafde959832c24415b75651673b0d7bbf145ac8d7cb744ba6926963d1d014836336df1317a134f4726b89983a8e7babbf6973e7ee16311e24328edf987bb0fbe7a494ec91e0000000000000000000000000000000000000000000000000000000000000000")

	commitment, _ := new(big.Int).SetString("2001904096268940627870837110796902048094724981649621731904769518577628368633", 10)
	assert.Equal(t, common.Bytes2Hex(commitment.FillBytes(make([]byte, 32))), "046d099ddea2c1ef130f85916df7e73d761454bd847cee12bb5919227c9a4ef9")
}

func TestPubData2(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(new(big.Int).SetInt64(1).FillBytes(make([]byte, 32)))
	buf.Write(new(big.Int).SetInt64(1654843322039).FillBytes(make([]byte, 32)))
	buf.Write(common.FromHex("14e4e8ad4848558d7200530337052e1ad30f5385b3c7187c80ad85f48547b74f"))
	buf.Write(common.FromHex("21422f9bebac15af8ddc504da0dbb88020c1a4de7e7b6722fe00acb0ed968942"))
	buf.Write(common.FromHex("01000000000000000000000000000000000000000000000000000000000000007472656173757279000000000000000000000000000000000000000000000000167c5363088a40a4839912a872f43164270740c7e986ec55397b2d583317ab4a2005db7af2bdcfae1fa8d28833ae2f1995e9a8e0825377cff121db64b0db21b718a96ca582a72b16f464330c89ab73277cb96e42df105ebf5c9ac5330d47b8fc0000000000000000000000000000000000000000000000000000000000000000"))
	buf.Write(new(big.Int).SetInt64(1).FillBytes(make([]byte, 32)))
	hFunc := mimc.NewMiMC()
	assert.Equal(t, common.Bytes2Hex(buf.Bytes()), "0000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000001814c592eb714e4e8ad4848558d7200530337052e1ad30f5385b3c7187c80ad85f48547b74f21422f9bebac15af8ddc504da0dbb88020c1a4de7e7b6722fe00acb0ed96894201000000000000000000000000000000000000000000000000000000000000007472656173757279000000000000000000000000000000000000000000000000167c5363088a40a4839912a872f43164270740c7e986ec55397b2d583317ab4a2005db7af2bdcfae1fa8d28833ae2f1995e9a8e0825377cff121db64b0db21b718a96ca582a72b16f464330c89ab73277cb96e42df105ebf5c9ac5330d47b8fc00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001")
	hFunc.Write(buf.Bytes())
	hashVal := hFunc.Sum(nil)
	assert.Equal(t, common.Bytes2Hex(hashVal), "0e4df6d7053619400d712319012c47b2cb7dcb2d83c203391547148b4f17741a")
}

func TestMiMCHash(t *testing.T) {
	hFunc := mimc.NewMiMC()
	hFunc.Write(new(big.Int).SetInt64(123123123).FillBytes(make([]byte, 32)))
	a := hFunc.Sum(nil)
	assert.Equal(t, new(big.Int).SetBytes(a).String(), "6158863128777714998435927227085268531294199267913818508594792142833376806078")
	assert.Equal(t, common.Bytes2Hex(curve.Modulus.Bytes()), "30644e72e131a029b85045b68181585d2833e84879b9709143e1f593f0000001")
}

func TestParsePubKey(t *testing.T) {
	pk, err := common2.ParsePubKey("58130e24cd20d9de8a110a20751f0a9b36089400ac0f20ca1993c28ee663318a")
	if err != nil {
		t.Fatal(err)
	}
	a := curve.ScalarBaseMul(big.NewInt(2))
	f, _ := new(big.Int).SetString("15527681003928902128179717624703512672403908117992798440346960750464748824729", 10)
	assert.Equal(t, ffmath.DivMod(new(big.Int).SetBytes(a.X.Marshal()), f, curve.Modulus).Int64(), int64(0))
	assert.Equal(t, a.Y.String(), "633281375905621697187330766174974863687049529291089048651929454608812697683")
	assert.True(t, pk.A.IsOnCurve())
	assert.Equal(t, pk.A.X.String(), "15824650925573404919778443019341920124666294571462377929750266189082392233365")
	assert.Equal(t, pk.A.Y.String(), "4610393480717259196086276896776664313868698522523751289329834907930790335320")
	assert.Equal(t, common.Bytes2Hex(pk.Bytes()), "58130e24cd20d9de8a110a20751f0a9b36089400ac0f20ca1993c28ee663318a")
}
