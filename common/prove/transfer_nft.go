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

package prove

import (
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"

	common2 "github.com/bnb-chain/zkbas/common"
	"github.com/bnb-chain/zkbas/types"
)

func (w *WitnessHelper) constructTransferNftCryptoTx(cryptoTx *CryptoTx, oTx *Tx) (*CryptoTx, error) {
	txInfo, err := types.ParseTransferNftTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("unable to parse transfer nft tx info:%s", err.Error())
		return nil, err
	}
	cryptoTxInfo, err := ToCryptoTransferNftTx(txInfo)
	if err != nil {
		logx.Errorf("unable to convert to crypto transfer nft tx: %s", err.Error())
		return nil, err
	}
	cryptoTx.TransferNftTxInfo = cryptoTxInfo
	cryptoTx.ExpiredAt = txInfo.ExpiredAt
	cryptoTx.Signature = new(eddsa.Signature)
	_, err = cryptoTx.Signature.SetBytes(txInfo.Sig)
	if err != nil {
		logx.Errorf("invalid sig bytes: %s", err.Error())
		return nil, err
	}
	return cryptoTx, nil
}

func ToCryptoTransferNftTx(txInfo *types.TransferNftTxInfo) (info *CryptoTransferNftTx, err error) {
	packedFee, err := common2.ToPackedFee(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("unable to convert to packed fee: %s", err.Error())
		return nil, err
	}
	info = &CryptoTransferNftTx{
		FromAccountIndex:  txInfo.FromAccountIndex,
		ToAccountIndex:    txInfo.ToAccountIndex,
		ToAccountNameHash: common.FromHex(txInfo.ToAccountNameHash),
		NftIndex:          txInfo.NftIndex,
		GasAccountIndex:   txInfo.GasAccountIndex,
		GasFeeAssetId:     txInfo.GasFeeAssetId,
		GasFeeAssetAmount: packedFee,
		CallDataHash:      txInfo.CallDataHash,
	}
	return info, nil
}
