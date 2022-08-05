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
 *
 */

package proverUtil

import (
	"errors"

	bsmt "github.com/bnb-chain/bas-smt"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/pkg/treedb"
)

func ConstructMintNftCryptoTx(
	oTx *Tx,
	treeCtx *treedb.Context,
	finalityBlockNr uint64,
	accountTree bsmt.SparseMerkleTree,
	accountAssetsTree *[]bsmt.SparseMerkleTree,
	liquidityTree bsmt.SparseMerkleTree,
	nftTree bsmt.SparseMerkleTree,
	accountModel AccountModel,
) (cryptoTx *CryptoTx, err error) {
	if oTx.TxType != commonTx.TxTypeMintNft {
		logx.Errorf("[ConstructMintNftCryptoTx] invalid tx type")
		return nil, errors.New("[ConstructMintNftCryptoTx] invalid tx type")
	}
	if oTx == nil || accountTree == nil || accountAssetsTree == nil || liquidityTree == nil || nftTree == nil {
		logx.Errorf("[ConstructMintNftCryptoTx] invalid params")
		return nil, errors.New("[ConstructMintNftCryptoTx] invalid params")
	}
	txInfo, err := commonTx.ParseMintNftTxInfo(oTx.TxInfo)
	if err != nil {
		logx.Errorf("[ConstructMintNftCryptoTx] unable to parse register zns tx info:%s", err.Error())
		return nil, err
	}
	cryptoTxInfo, err := ToCryptoMintNftTx(txInfo)
	if err != nil {
		logx.Errorf("[ConstructMintNftCryptoTx] unable to convert to crypto register zns tx: %s", err.Error())
		return nil, err
	}
	accountKeys, proverAccounts, proverLiquidityInfo, proverNftInfo, err := ConstructProverInfo(oTx, accountModel)
	if err != nil {
		logx.Errorf("[ConstructMintNftCryptoTx] unable to construct prover info: %s", err.Error())
		return nil, err
	}
	cryptoTx, err = ConstructWitnessInfo(
		oTx,
		accountModel,
		treeCtx,
		finalityBlockNr,
		accountTree,
		accountAssetsTree,
		liquidityTree,
		nftTree,
		accountKeys,
		proverAccounts,
		proverLiquidityInfo,
		proverNftInfo,
	)
	if err != nil {
		logx.Errorf("[ConstructMintNftCryptoTx] unable to construct witness info: %s", err.Error())
		return nil, err
	}
	cryptoTx.TxType = uint8(oTx.TxType)
	cryptoTx.MintNftTxInfo = cryptoTxInfo
	cryptoTx.Nonce = oTx.Nonce
	cryptoTx.ExpiredAt = txInfo.ExpiredAt
	cryptoTx.Signature = new(eddsa.Signature)
	_, err = cryptoTx.Signature.SetBytes(txInfo.Sig)
	if err != nil {
		logx.Errorf("[ConstructMintNftCryptoTx] invalid sig bytes: %s", err.Error())
		return nil, err
	}
	return cryptoTx, nil
}

func ToCryptoMintNftTx(txInfo *commonTx.MintNftTxInfo) (info *CryptoMintNftTx, err error) {
	packedFee, err := util.ToPackedFee(txInfo.GasFeeAssetAmount)
	if err != nil {
		logx.Errorf("[ToCryptoSwapTx] unable to convert to packed fee: %s", err.Error())
		return nil, err
	}
	info = &CryptoMintNftTx{
		CreatorAccountIndex: txInfo.CreatorAccountIndex,
		ToAccountIndex:      txInfo.ToAccountIndex,
		ToAccountNameHash:   common.FromHex(txInfo.ToAccountNameHash),
		NftIndex:            txInfo.NftIndex,
		NftContentHash:      common.FromHex(txInfo.NftContentHash),
		CreatorTreasuryRate: txInfo.CreatorTreasuryRate,
		GasAccountIndex:     txInfo.GasAccountIndex,
		GasFeeAssetId:       txInfo.GasFeeAssetId,
		GasFeeAssetAmount:   packedFee,
		CollectionId:        txInfo.NftCollectionId,
		ExpiredAt:           txInfo.ExpiredAt,
	}
	return info, nil
}
