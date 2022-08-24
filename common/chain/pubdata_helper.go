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

package chain

import (
	"errors"

	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"

	common2 "github.com/bnb-chain/zkbas/common"
	"github.com/bnb-chain/zkbas/types"
)

func ParseRegisterZnsPubData(pubData []byte) (tx *types.RegisterZnsTxInfo, err error) {
	/*
		struct RegisterZNS {
			uint8 txType;
			bytes32 accountName;
			bytes32 accountNameHash;
			bytes32 pubKeyX;
			bytes32 pubKeyY;
		}
	*/
	if len(pubData) != types.RegisterZnsPubDataSize {
		logx.Errorf("[ParseRegisterZnsPubData] invalid size")
		return nil, errors.New("[ParseRegisterZnsPubData] invalid size")
	}
	offset := 0
	offset, txType := common2.ReadUint8(pubData, offset)
	offset, accountIndex := common2.ReadUint32(pubData, offset)
	offset, accountName := common2.ReadBytes32(pubData, offset)
	offset, accountNameHash := common2.ReadBytes32(pubData, offset)
	offset, pubKeyX := common2.ReadBytes32(pubData, offset)
	offset, pubKeyY := common2.ReadBytes32(pubData, offset)
	pk := new(eddsa.PublicKey)
	pk.A.X.SetBytes(pubKeyX)
	pk.A.Y.SetBytes(pubKeyY)
	tx = &types.RegisterZnsTxInfo{
		TxType:          txType,
		AccountIndex:    int64(accountIndex),
		AccountName:     common2.CleanAccountName(common2.SerializeAccountName(accountName)),
		AccountNameHash: accountNameHash,
		PubKey:          common.Bytes2Hex(pk.Bytes()),
	}
	return tx, nil
}

func ParseCreatePairPubData(pubData []byte) (tx *types.CreatePairTxInfo, err error) {
	if len(pubData) != types.CreatePairPubDataSize {
		logx.Errorf("[ParseCreatePairPubData] invalid size")
		return nil, errors.New("[ParseCreatePairPubData] invalid size")
	}
	offset := 0
	offset, txType := common2.ReadUint8(pubData, offset)
	offset, pairIndex := common2.ReadUint16(pubData, offset)
	offset, assetAId := common2.ReadUint16(pubData, offset)
	offset, assetBId := common2.ReadUint16(pubData, offset)
	offset, feeRate := common2.ReadUint16(pubData, offset)
	offset, treasuryAccountIndex := common2.ReadUint32(pubData, offset)
	offset, treasuryRate := common2.ReadUint16(pubData, offset)
	tx = &types.CreatePairTxInfo{
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

func ParseUpdatePairRatePubData(pubData []byte) (tx *types.UpdatePairRateTxInfo, err error) {
	if len(pubData) != types.UpdatePairRatePubdataSize {
		logx.Errorf("[ParseUpdatePairRatePubData] invalid size")
		return nil, errors.New("[ParseUpdatePairRatePubData] invalid size")
	}
	offset := 0
	offset, txType := common2.ReadUint8(pubData, offset)
	offset, pairIndex := common2.ReadUint16(pubData, offset)
	offset, feeRate := common2.ReadUint16(pubData, offset)
	offset, treasuryAccountIndex := common2.ReadUint32(pubData, offset)
	offset, treasuryRate := common2.ReadUint16(pubData, offset)
	tx = &types.UpdatePairRateTxInfo{
		TxType:               txType,
		PairIndex:            int64(pairIndex),
		FeeRate:              int64(feeRate),
		TreasuryAccountIndex: int64(treasuryAccountIndex),
		TreasuryRate:         int64(treasuryRate),
	}
	return tx, nil
}

func ParseDepositPubData(pubData []byte) (tx *types.DepositTxInfo, err error) {
	/*
		struct Deposit {
			uint8 txType;
			uint32 accountIndex;
			bytes32 accountNameHash;
			uint16 assetId;
			uint128 amount;
		}
	*/
	if len(pubData) != types.DepositPubDataSize {
		logx.Errorf("[ParseDepositPubData] invalid size")
		return nil, errors.New("[ParseDepositPubData] invalid size")
	}
	offset := 0
	offset, txType := common2.ReadUint8(pubData, offset)
	offset, accountIndex := common2.ReadUint32(pubData, offset)
	offset, accountNameHash := common2.ReadBytes32(pubData, offset)
	offset, assetId := common2.ReadUint16(pubData, offset)
	offset, amount := common2.ReadUint128(pubData, offset)
	tx = &types.DepositTxInfo{
		TxType:          txType,
		AccountIndex:    int64(accountIndex),
		AccountNameHash: accountNameHash,
		AssetId:         int64(assetId),
		AssetAmount:     amount,
	}
	return tx, nil
}

func ParseDepositNftPubData(pubData []byte) (tx *types.DepositNftTxInfo, err error) {
	if len(pubData) != types.DepositNftPubDataSize {
		logx.Errorf("[ParseDepositNftPubData] invalid size")
		return nil, errors.New("[ParseDepositNftPubData] invalid size")
	}
	offset := 0
	offset, txType := common2.ReadUint8(pubData, offset)
	offset, accountIndex := common2.ReadUint32(pubData, offset)
	offset, nftIndex := common2.ReadUint40(pubData, offset)
	offset, nftL1Address := common2.ReadAddress(pubData, offset)
	offset, creatorAccountIndex := common2.ReadUint32(pubData, offset)
	offset, creatorTreasuryRate := common2.ReadUint16(pubData, offset)
	offset, nftContentHash := common2.ReadBytes32(pubData, offset)
	offset, nftL1TokenId := common2.ReadUint256(pubData, offset)
	offset, accountNameHash := common2.ReadBytes32(pubData, offset)
	offset, collectionId := common2.ReadUint16(pubData, offset)
	tx = &types.DepositNftTxInfo{
		TxType:              txType,
		AccountIndex:        int64(accountIndex),
		NftIndex:            nftIndex,
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

func ParseFullExitPubData(pubData []byte) (tx *types.FullExitTxInfo, err error) {
	if len(pubData) != types.FullExitPubDataSize {
		logx.Errorf("[ParseFullExitPubData] invalid size")
		return nil, errors.New("[ParseFullExitPubData] invalid size")
	}
	offset := 0
	offset, txType := common2.ReadUint8(pubData, offset)
	offset, accountIndex := common2.ReadUint32(pubData, offset)
	offset, assetId := common2.ReadUint16(pubData, offset)
	offset, assetAmount := common2.ReadUint128(pubData, offset)
	offset, accountNameHash := common2.ReadBytes32(pubData, offset)
	tx = &types.FullExitTxInfo{
		TxType:          txType,
		AccountIndex:    int64(accountIndex),
		AccountNameHash: accountNameHash,
		AssetId:         int64(assetId),
		AssetAmount:     assetAmount,
	}
	return tx, nil
}

func ParseFullExitNftPubData(pubData []byte) (tx *legendTxTypes.FullExitNftTxInfo, err error) {
	if len(pubData) != types.FullExitNftPubDataSize {
		logx.Errorf("[ParseFullExitNftPubData] invalid size")
		return nil, errors.New("[ParseFullExitNftPubData] invalid size")
	}
	offset := 0
	offset, txType := common2.ReadUint8(pubData, offset)
	offset, accountIndex := common2.ReadUint32(pubData, offset)
	offset, creatorAccountIndex := common2.ReadUint32(pubData, offset)
	offset, creatorTreasuryRate := common2.ReadUint16(pubData, offset)
	offset, nftIndex := common2.ReadUint40(pubData, offset)
	offset, collectionId := common2.ReadUint16(pubData, offset)
	offset, nftL1Address := common2.ReadAddress(pubData, offset)
	offset, accountNameHash := common2.ReadBytes32(pubData, offset)
	offset, creatorAccountNameHash := common2.ReadBytes32(pubData, offset)
	offset, nftContentHash := common2.ReadBytes32(pubData, offset)
	offset, nftL1TokenId := common2.ReadUint256(pubData, offset)
	tx = &types.FullExitNftTxInfo{
		TxType:                 txType,
		AccountIndex:           int64(accountIndex),
		CreatorAccountIndex:    int64(creatorAccountIndex),
		CreatorTreasuryRate:    int64(creatorTreasuryRate),
		NftIndex:               nftIndex,
		CollectionId:           int64(collectionId),
		NftL1Address:           nftL1Address,
		AccountNameHash:        accountNameHash,
		CreatorAccountNameHash: creatorAccountNameHash,
		NftContentHash:         nftContentHash,
		NftL1TokenId:           nftL1TokenId,
	}
	return tx, nil
}
