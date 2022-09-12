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

func FillInTxWitness(oTx *Tx) (witness *TxWitness, err error) {
	witness.TxType = uint8(oTx.TxType)
	witness.Nonce = oTx.Nonce
	switch oTx.TxType {
	case types.TxTypeRegisterZns:
		return constructRegisterZnsTxWitness(witness, oTx)
	case types.TxTypeCreatePair:
		return constructCreatePairTxWitness(witness, oTx)
	case types.TxTypeUpdatePairRate:
		return constructUpdatePairRateTxWitness(witness, oTx)
	case types.TxTypeDeposit:
		return constructDepositTxWitness(witness, oTx)
	case types.TxTypeDepositNft:
		return constructDepositNftTxWitness(witness, oTx)
	case types.TxTypeTransfer:
		return constructTransferTxWitness(witness, oTx)
	case types.TxTypeSwap:
		return constructSwapTxWitness(witness, oTx)
	case types.TxTypeAddLiquidity:
		return constructAddLiquidityTxWitness(witness, oTx)
	case types.TxTypeRemoveLiquidity:
		return constructRemoveLiquidityTxWitness(witness, oTx)
	case types.TxTypeWithdraw:
		return constructWithdrawTxWitness(witness, oTx)
	case types.TxTypeCreateCollection:
		return constructCreateCollectionTxWitness(witness, oTx)
	case types.TxTypeMintNft:
		return constructMintNftTxWitness(witness, oTx)
	case types.TxTypeTransferNft:
		return constructTransferNftTxWitness(witness, oTx)
	case types.TxTypeAtomicMatch:
		return constructAtomicMatchTxWitness(witness, oTx)
	case types.TxTypeCancelOffer:
		return constructCancelOfferTxWitness(witness, oTx)
	case types.TxTypeWithdrawNft:
		return constructWithdrawNftTxWitness(witness, oTx)
	case types.TxTypeFullExit:
		return constructFullExitTxWitness(witness, oTx)
	case types.TxTypeFullExitNft:
		return constructFullExitNftTxWitness(witness, oTx)
	default:
		return nil, fmt.Errorf("tx type error")
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
