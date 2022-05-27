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

package txVerification

import (
	"errors"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/zecrey-labs/zecrey-crypto/wasm/zecrey-legend/legendTxTypes"
	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zeromicro/go-zero/core/logx"
	"log"
)

func VerifyOfferTxInfo(
	accountInfoMap map[int64]*AccountInfo,
	nftInfo *NftInfo,
	txInfo *OfferTxInfo,
) (err error) {
	// verify params
	if accountInfoMap[txInfo.AccountIndex] == nil ||
		accountInfoMap[txInfo.AccountIndex].AssetInfo == nil ||
		accountInfoMap[txInfo.AccountIndex].AssetInfo[txInfo.AssetId] == nil ||
		accountInfoMap[txInfo.AccountIndex].AssetInfo[txInfo.AssetId].Balance.Cmp(ZeroBigInt) <= 0 ||
		nftInfo.NftIndex != txInfo.NftIndex ||
		txInfo.AssetAmount.Cmp(ZeroBigInt) <= 0 {
		logx.Errorf("[VerifyMintNftTxInfo] invalid params")
		return errors.New("[VerifyMintNftTxInfo] invalid params")
	}
	// check if it is a buy offer, check enough balance
	if txInfo.Type == commonAsset.BuyOfferType {
		if accountInfoMap[txInfo.AccountIndex].AssetInfo[txInfo.AssetId].Balance.Cmp(txInfo.AssetAmount) < 0 {
			logx.Errorf("[VerifyMintNftTxInfo] you don't have enough balance")
			return errors.New("[VerifyMintNftTxInfo] you don't have enough balance")
		}
	}
	// verify sig
	hFunc := mimc.NewMiMC()
	msgHash := legendTxTypes.ComputeOfferMsgHash(txInfo, hFunc)
	// verify signature
	hFunc.Reset()
	pk, err := ParsePkStr(accountInfoMap[txInfo.AccountIndex].PublicKey)
	if err != nil {
		return err
	}
	isValid, err := pk.Verify(txInfo.Sig, msgHash, hFunc)
	if err != nil {
		log.Println("[VerifyCancelOfferTxInfo] unable to verify signature:", err)
		return err
	}
	if !isValid {
		log.Println("[VerifyCancelOfferTxInfo] invalid signature")
		return errors.New("[VerifyCancelOfferTxInfo] invalid signature")
	}
	return nil
}
