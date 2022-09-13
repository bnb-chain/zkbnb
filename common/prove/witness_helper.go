/*
 * Copyright © 2021 ZkBNB Protocol
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

package prove

import (
	"fmt"
	"github.com/bnb-chain/zkbnb/types"
)

func FillTxWitness(witness *TxWitness, oTx *Tx) error {
	switch oTx.TxType {
	case types.TxTypeRegisterZns:
		return fillRegisterZnsTxWitness(witness, oTx)
	case types.TxTypeCreatePair:
		return fillCreatePairTxWitness(witness, oTx)
	case types.TxTypeUpdatePairRate:
		return fillUpdatePairRateTxWitness(witness, oTx)
	case types.TxTypeDeposit:
		return fillDepositTxWitness(witness, oTx)
	case types.TxTypeDepositNft:
		return fillDepositNftTxWitness(witness, oTx)
	case types.TxTypeTransfer:
		return fillTransferTxWitness(witness, oTx)
	case types.TxTypeSwap:
		return fillSwapTxWitness(witness, oTx)
	case types.TxTypeAddLiquidity:
		return fillAddLiquidityTxWitness(witness, oTx)
	case types.TxTypeRemoveLiquidity:
		return fillRemoveLiquidityTxWitness(witness, oTx)
	case types.TxTypeWithdraw:
		return fillWithdrawTxWitness(witness, oTx)
	case types.TxTypeCreateCollection:
		return fillCreateCollectionTxWitness(witness, oTx)
	case types.TxTypeMintNft:
		return fillMintNftTxWitness(witness, oTx)
	case types.TxTypeTransferNft:
		return fillTransferNftTxWitness(witness, oTx)
	case types.TxTypeAtomicMatch:
		return fillAtomicMatchTxWitness(witness, oTx)
	case types.TxTypeCancelOffer:
		return fillCancelOfferTxWitness(witness, oTx)
	case types.TxTypeWithdrawNft:
		return fillWithdrawNftTxWitness(witness, oTx)
	case types.TxTypeFullExit:
		return fillFullExitTxWitness(witness, oTx)
	case types.TxTypeFullExitNft:
		return fillFullExitNftTxWitness(witness, oTx)
	default:
		return fmt.Errorf("tx type error")
	}
}

func SetFixedAccountArray(proof [][]byte) (res [AccountMerkleLevels][]byte, err error) {
	if len(proof) != AccountMerkleLevels {
		return res, fmt.Errorf("invalid size")
	}
	copy(res[:], proof[:])
	return res, nil
}

func SetFixedAccountAssetArray(proof [][]byte) (res [AssetMerkleLevels][]byte, err error) {
	if len(proof) != AssetMerkleLevels {
		return res, fmt.Errorf("invalid size")
	}
	copy(res[:], proof[:])
	return res, nil
}

func SetFixedLiquidityArray(proof [][]byte) (res [LiquidityMerkleLevels][]byte, err error) {
	if len(proof) != LiquidityMerkleLevels {
		return res, fmt.Errorf("invalid size")
	}
	copy(res[:], proof[:])
	return res, nil
}

func SetFixedNftArray(proof [][]byte) (res [NftMerkleLevels][]byte, err error) {
	if len(proof) != NftMerkleLevels {
		return res, fmt.Errorf("invalid size")
	}
	copy(res[:], proof[:])
	return res, nil
}
