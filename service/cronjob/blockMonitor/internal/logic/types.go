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

import "math/big"

type L1EventInfo struct {
	// deposit / lock / committed / verified / reverted
	EventType uint8
	// tx hash
	TxHash string
}

type ChainConfig struct {
	L2ChainId                uint8
	NativeChainId            *big.Int
	NetworkRPC               string
	ZecreyLegendContractAddr string
	GovernanceContractAddr   string
	StartL1BlockHeight       int64
	PendingBlocksCount       uint64
}
