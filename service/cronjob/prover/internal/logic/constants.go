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
 */

package logic

import (
	cryptoBlock "github.com/bnb-chain/zkbas-crypto/legend/circuit/bn254/block"
	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/frontend"
)

type (
	CryptoBlock = cryptoBlock.Block
)

const RedisLockKey = "prover_mutex_key"

var (
	VerifyingKeys []plonk.VerifyingKey
	ProvingKeys   []plonk.ProvingKey
	KeyTxCounts   []int
	R1cs          []frontend.CompiledConstraintSystem
)
