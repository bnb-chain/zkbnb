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

package util

import (
	"errors"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/zeromicro/go-zero/core/logx"
)

func ParseRegisterZnsPubData(pubData []byte) (tx *RegisterZnsTxInfo, err error) {
	/*
		struct RegisterZNS {
			uint8 txType;
			bytes32 accountName;
			bytes32 accountNameHash;
			bytes32 pubKeyX;
			bytes32 pubKeyY;
		}
	*/
	if len(pubData) != RegisterZnsPubDataSize {
		logx.Errorf("[ParseRegisterZnsPubData] invalid size")
		return nil, errors.New("[ParseRegisterZnsPubData] invalid size")
	}
	offset := 0
	offset, txType := ReadUint8(pubData, offset)
	offset, accountIndex := ReadUint32(pubData, offset)
	offset, accountName := ReadBytes32(pubData, offset)
	offset, accountNameHash := ReadBytes32(pubData, offset)
	offset, pubKeyX := ReadBytes32(pubData, offset)
	offset, pubKeyY := ReadBytes32(pubData, offset)
	pk := new(eddsa.PublicKey)
	pk.A.X.SetBytes(pubKeyX)
	pk.A.Y.SetBytes(pubKeyY)
	tx = &RegisterZnsTxInfo{
		TxType:          txType,
		AccountIndex:    int64(accountIndex),
		AccountName:     CleanAccountName(SerializeAccountName(accountName)),
		AccountNameHash: accountNameHash,
		PubKey:          common.Bytes2Hex(pk.Bytes()),
	}
	return tx, nil
}

func ParseCreatePairPubData(pubData []byte) (tx *CreatePairTxInfo, err error) {
	if len(pubData) != CreatePairPubDataSize {
		logx.Errorf("[ParseCreatePairPubData] invalid size")
		return nil, errors.New("[ParseCreatePairPubData] invalid size")
	}
	offset := 0
	offset, txType := ReadUint8(pubData, offset)
	offset, pairIndex := ReadUint16(pubData, offset)
	offset, assetAId := ReadUint16(pubData, offset)
	offset, assetBId := ReadUint16(pubData, offset)
	offset, feeRate := ReadUint16(pubData, offset)
	offset, treasuryAccountIndex := ReadUint32(pubData, offset)
	offset, treasuryRate := ReadUint16(pubData, offset)
	tx = &CreatePairTxInfo{
		TxType:               txType,
		PairIndex:            int64(pairIndex),
		AssetAId:             int64(assetAId),
		AssetBId:             int64(assetBId),
		FeeRate:              int64(feeRate),
		TreasuryAccountIndex: int64(treasuryAccountIndex),
		TreasuryRate:         int64(treasuryRate),
	}
	return tx, nil
}

func ParseUpdatePairRatePubData(pubData []byte) (tx *UpdatePairRateTxInfo, err error) {
	if len(pubData) != UpdatePairRatePubdataSize {
		logx.Errorf("[ParseUpdatePairRatePubData] invalid size")
		return nil, errors.New("[ParseUpdatePairRatePubData] invalid size")
	}
	offset := 0
	offset, txType := ReadUint8(pubData, offset)
	offset, pairIndex := ReadUint16(pubData, offset)
	offset, feeRate := ReadUint16(pubData, offset)
	offset, treasuryAccountIndex := ReadUint32(pubData, offset)
	offset, treasuryRate := ReadUint16(pubData, offset)
	tx = &UpdatePairRateTxInfo{
		TxType:               txType,
		PairIndex:            int64(pairIndex),
		FeeRate:              int64(feeRate),
		TreasuryAccountIndex: int64(treasuryAccountIndex),
		TreasuryRate:         int64(treasuryRate),
	}
	return tx, nil
}

func ParseDepositPubData(pubData []byte) (tx *DepositTxInfo, err error) {
	/*
		struct Deposit {
			uint8 txType;
			uint32 accountIndex;
			bytes32 accountNameHash;
			uint16 assetId;
			uint128 amount;
		}
	*/
	if len(pubData) != DepositPubDataSize {
		logx.Errorf("[ParseDepositPubData] invalid size")
		return nil, errors.New("[ParseDepositPubData] invalid size")
	}
	offset := 0
	offset, txType := ReadUint8(pubData, offset)
	offset, accountIndex := ReadUint32(pubData, offset)
	offset, accountNameHash := ReadBytes32(pubData, offset)
	offset, assetId := ReadUint16(pubData, offset)
	offset, amount := ReadUint128(pubData, offset)
	tx = &DepositTxInfo{
		TxType:          txType,
		AccountIndex:    int64(accountIndex),
		AccountNameHash: accountNameHash,
		AssetId:         int64(assetId),
		AssetAmount:     amount,
	}
	return tx, nil
}

func ParseDepositNftPubData(pubData []byte) (tx *DepositNftTxInfo, err error) {
	if len(pubData) != DepositNftPubDataSize {
		logx.Errorf("[ParseDepositNftPubData] invalid size")
		return nil, errors.New("[ParseDepositNftPubData] invalid size")
	}
	offset := 0
	offset, txType := ReadUint8(pubData, offset)
	offset, accountIndex := ReadUint32(pubData, offset)
	offset, nftIndex := ReadUint40(pubData, offset)
	offset, nftL1Address := ReadAddress(pubData, offset)
	offset, creatorAccountIndex := ReadUint32(pubData, offset)
	offset, creatorTreasuryRate := ReadUint16(pubData, offset)
	offset, nftContentHash := ReadBytes32(pubData, offset)
	offset, nftL1TokenId := ReadUint256(pubData, offset)
	offset, accountNameHash := ReadBytes32(pubData, offset)
	offset, collectionId := ReadUint16(pubData, offset)
	tx = &DepositNftTxInfo{
		TxType:              txType,
		AccountIndex:        int64(accountIndex),
		NftIndex:            int64(nftIndex),
		NftL1Address:        nftL1Address,
		CreatorAccountIndex: int64(creatorAccountIndex),
		CreatorTreasuryRate: int64(creatorTreasuryRate),
		NftContentHash:      nftContentHash,
		NftL1TokenId:        nftL1TokenId,
		AccountNameHash:     accountNameHash,
		CollectionId:        int64(collectionId),
	}
	return tx, nil
}

func ParseFullExitPubData(pubData []byte) (tx *FullExitTxInfo, err error) {
	if len(pubData) != FullExitPubDataSize {
		logx.Errorf("[ParseFullExitPubData] invalid size")
		return nil, errors.New("[ParseFullExitPubData] invalid size")
	}
	offset := 0
	offset, txType := ReadUint8(pubData, offset)
	offset, accountIndex := ReadUint32(pubData, offset)
	offset, assetId := ReadUint16(pubData, offset)
	offset, assetAmount := ReadUint128(pubData, offset)
	offset, accountNameHash := ReadBytes32(pubData, offset)
	tx = &FullExitTxInfo{
		TxType:          txType,
		AccountIndex:    int64(accountIndex),
		AccountNameHash: accountNameHash,
		AssetId:         int64(assetId),
		AssetAmount:     assetAmount,
	}
	return tx, nil
}

func ParseFullExitNftPubData(pubData []byte) (tx *commonTx.FullExitNftTxInfo, err error) {
	if len(pubData) != FullExitNftPubDataSize {
		logx.Errorf("[ParseFullExitNftPubData] invalid size")
		return nil, errors.New("[ParseFullExitNftPubData] invalid size")
	}
	offset := 0
	offset, txType := ReadUint8(pubData, offset)
	offset, accountIndex := ReadUint32(pubData, offset)
	offset, creatorAccountIndex := ReadUint32(pubData, offset)
	offset, creatorTreasuryRate := ReadUint16(pubData, offset)
	offset, nftIndex := ReadUint40(pubData, offset)
	offset, collectionId := ReadUint16(pubData, offset)
	offset, nftL1Address := ReadAddress(pubData, offset)
	offset, accountNameHash := ReadBytes32(pubData, offset)
	offset, creatorAccountNameHash := ReadBytes32(pubData, offset)
	offset, nftContentHash := ReadBytes32(pubData, offset)
	offset, nftL1TokenId := ReadUint256(pubData, offset)
	tx = &FullExitNftTxInfo{
		TxType:                 txType,
		AccountIndex:           int64(accountIndex),
		CreatorAccountIndex:    int64(creatorAccountIndex),
		CreatorTreasuryRate:    int64(creatorTreasuryRate),
		NftIndex:               int64(nftIndex),
		CollectionId:           int64(collectionId),
		NftL1Address:           nftL1Address,
		AccountNameHash:        accountNameHash,
		CreatorAccountNameHash: creatorAccountNameHash,
		NftContentHash:         nftContentHash,
		NftL1TokenId:           nftL1TokenId,
	}
	return tx, nil
}
