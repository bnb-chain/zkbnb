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
 */

package logic

import (
	"fmt"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/ethereum/go-ethereum/common"
	"log"
	"math/big"
	"testing"
	"time"
)

func TestTryLock(t *testing.T) {
	var m Mutex
	go func() {
		m.Lock()
	}()
	time.Sleep(time.Second)
	fmt.Printf("TryLock1: %t\n", m.TryLock()) //false
	fmt.Printf("TryLock1: %t\n", m.TryLock()) //false
	fmt.Printf("TryLock1: %t\n", m.TryLock()) //false
	fmt.Printf("TryLock1: %t\n", m.TryLock()) //false
	fmt.Printf("TryLock1: %t\n", m.TryLock()) //false
	fmt.Printf("TryLock1: %t\n", m.TryLock()) //false
	fmt.Printf("TryLock1: %t\n", m.TryLock()) //false
	fmt.Printf("TryLock1: %t\n", m.TryLock()) //false
	m.Unlock()
	fmt.Printf("TryLock3: %t\n", m.TryLock()) //true
	m.Unlock()
}

func TestGetLatestUnprovedBlockHeight(t *testing.T) {
	hFunc := mimc.NewMiMC()
	a, _ := new(big.Int).SetString("52670410704698717926084766743862344617313117190279404211741912966309903073280", 10)
	hFunc.Write(a.FillBytes(make([]byte, 32)))
	aHash := hFunc.Sum(nil)
	log.Println(common.Bytes2Hex(aHash))
	b, _ := new(big.Int).SetString("8893924961020167481591955253347794440216388389447335524345504593158286082046", 10)
	hFunc.Reset()
	hFunc.Write(b.FillBytes(make([]byte, 32)))
	bHash := hFunc.Sum(nil)
	log.Println(common.Bytes2Hex(bHash))
}
