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
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	"github.com/ethereum/go-ethereum/common"

	cryptoTypes "github.com/bnb-chain/zkbnb-crypto/circuit/types"
	"github.com/bnb-chain/zkbnb-crypto/wasm/txtypes"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

func (w *WitnessHelper) constructMintNftTxWitness(cryptoTx *TxWitness, oTx *tx.Tx) (*TxWitness, error) {
	txInfo, err := types.ParseMintNftTxInfo(oTx.TxInfo)
	if err != nil {
		return nil, err
	}
	cryptoTxInfo, err := toCryptoMintNftTx(txInfo)
	if err != nil {
		return nil, err
	}
	cryptoTx.MintNftTxInfo = cryptoTxInfo
	cryptoTx.ExpiredAt = txInfo.ExpiredAt
	cryptoTx.Signature = new(eddsa.Signature)
	_, err = cryptoTx.Signature.SetBytes(txInfo.Sig)
	if err != nil {
		return nil, err
	}
	return cryptoTx, nil
}

func toCryptoMintNftTx(txInfo *txtypes.MintNftTxInfo) (info *cryptoTypes.MintNftTx, err error) {
	packedFee, err := common2.ToPackedFee(txInfo.GasFeeAssetAmount)
	if err != nil {
		return nil, err
	}
	info = &cryptoTypes.MintNftTx{
		CreatorAccountIndex: txInfo.CreatorAccountIndex,
		ToAccountIndex:      txInfo.ToAccountIndex,
		ToL1Address:         common2.AddressStrToBytes(txInfo.ToL1Address),
		NftIndex:            txInfo.NftIndex,
		NftContentHash:      common2.PrefixPaddingBufToChunkSize(common.FromHex(txInfo.NftContentHash)),
		RoyaltyRate:         txInfo.RoyaltyRate,
		GasAccountIndex:     txInfo.GasAccountIndex,
		GasFeeAssetId:       txInfo.GasFeeAssetId,
		GasFeeAssetAmount:   packedFee,
		CollectionId:        txInfo.NftCollectionId,
		ExpiredAt:           txInfo.ExpiredAt,
		NftContentType:      txInfo.NftContentType,
	}
	return info, nil
}
