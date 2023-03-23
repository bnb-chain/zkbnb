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

package chain

import (
	types2 "github.com/bnb-chain/zkbnb-crypto/circuit/types"
	"github.com/bnb-chain/zkbnb-crypto/util"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"
	"math/big"

	"github.com/bnb-chain/zkbnb-crypto/wasm/txtypes"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/types"
)

func ParseDepositPubData(pubData []byte) (tx *txtypes.DepositTxInfo, err error) {
	/*
		struct Deposit {
			uint8 txType;
			uint32 accountIndex;
			bytes20 L1Address;
			uint16 assetId;
			uint128 amount;
		}
	*/
	if len(pubData) != types.DepositPubDataSize {
		logx.Error("[ParseDepositPubData] invalid size")
		return nil, types.AppErrDepositPubDataInvalidSize
	}
	offset := 0
	offset, txType := common2.ReadUint8(pubData, offset)
	offset, accountIndex := common2.ReadUint32(pubData, offset)
	offset, l1Address := common2.ReadAddress(pubData, offset)
	offset, assetId := common2.ReadUint16(pubData, offset)
	_, amount := common2.ReadUint128(pubData, offset)
	tx = &txtypes.DepositTxInfo{
		TxType:       txType,
		AccountIndex: int64(accountIndex),
		L1Address:    l1Address,
		AssetId:      int64(assetId),
		AssetAmount:  amount,
	}
	return tx, nil
}

func ParseDepositNftPubData(pubData []byte) (tx *txtypes.DepositNftTxInfo, err error) {
	if len(pubData) != types.DepositNftPubDataSize {
		logx.Error("[ParseDepositNftPubData] invalid size")
		return nil, types.AppErrDepositNFTPubDataInvalidSize
	}
	offset := 0
	offset, txType := common2.ReadUint8(pubData, offset)
	offset, accountIndex := common2.ReadUint32(pubData, offset)
	offset, creatorAccountIndex := common2.ReadUint32(pubData, offset)
	offset, royaltyRate := common2.ReadUint16(pubData, offset)
	offset, nftIndex := common2.ReadUint40(pubData, offset)
	offset, collectionId := common2.ReadUint16(pubData, offset)
	offset, l1Address := common2.ReadAddress(pubData, offset)
	offset, nftContentHash := common2.ReadBytes32(pubData, offset)
	_, nftContentType := common2.ReadUint8(pubData, offset)

	tx = &txtypes.DepositNftTxInfo{
		TxType:              txType,
		AccountIndex:        int64(accountIndex),
		NftIndex:            nftIndex,
		CreatorAccountIndex: int64(creatorAccountIndex),
		RoyaltyRate:         int64(royaltyRate),
		NftContentHash:      nftContentHash,
		L1Address:           l1Address,
		CollectionId:        int64(collectionId),
		NftContentType:      int64(nftContentType),
	}
	return tx, nil
}

func ParseFullExitPubData(pubData []byte) (tx *txtypes.FullExitTxInfo, err error) {
	if len(pubData) != types.FullExitPubDataSize {
		logx.Error("[ParseFullExitPubData] invalid size")
		return nil, types.AppErrFullExitPubDataInvalidSize
	}
	offset := 0
	offset, txType := common2.ReadUint8(pubData, offset)
	offset, accountIndex := common2.ReadUint32(pubData, offset)
	offset, assetId := common2.ReadUint16(pubData, offset)
	offset, assetAmount := common2.ReadUint128(pubData, offset)
	_, l1Address := common2.ReadAddress(pubData, offset)
	tx = &txtypes.FullExitTxInfo{
		TxType:       txType,
		AccountIndex: int64(accountIndex),
		L1Address:    l1Address,
		AssetId:      int64(assetId),
		AssetAmount:  assetAmount,
	}
	return tx, nil
}

func ParseFullExitNftPubData(pubData []byte) (tx *txtypes.FullExitNftTxInfo, err error) {
	if len(pubData) != types.FullExitNftPubDataSize {
		logx.Error("[ParseFullExitNftPubData] invalid size")
		return nil, types.AppErrFullExitNftPubDataInvalidSize
	}
	offset := 0
	offset, txType := common2.ReadUint8(pubData, offset)
	offset, accountIndex := common2.ReadUint32(pubData, offset)
	offset, creatorAccountIndex := common2.ReadUint32(pubData, offset)
	offset, royaltyRate := common2.ReadUint16(pubData, offset)
	offset, nftIndex := common2.ReadUint40(pubData, offset)
	offset, collectionId := common2.ReadUint16(pubData, offset)
	offset, l1Address := common2.ReadAddress(pubData, offset)
	offset, creatorL1Address := common2.ReadAddress(pubData, offset)
	offset, nftContentHash := common2.ReadBytes32(pubData, offset)
	_, nftContentType := common2.ReadUint8(pubData, offset)

	tx = &txtypes.FullExitNftTxInfo{
		TxType:              txType,
		AccountIndex:        int64(accountIndex),
		CreatorAccountIndex: int64(creatorAccountIndex),
		RoyaltyRate:         int64(royaltyRate),
		NftIndex:            nftIndex,
		CollectionId:        int64(collectionId),
		L1Address:           l1Address,
		CreatorL1Address:    creatorL1Address,
		NftContentHash:      nftContentHash,
		NftContentType:      int64(nftContentType),
	}
	return tx, nil
}

func ParsePubData(pubData []byte) {
	fromOffset := 0
	toOffset := types2.PubDataBitsSizePerTx / 8
	len := len(pubData)
	for toOffset <= len {
		res := make([]byte, types2.PubDataBitsSizePerTx/8)
		copy(res[:], pubData[fromOffset:toOffset])
		fromOffset = toOffset
		toOffset += types2.PubDataBitsSizePerTx / 8
		str := common.Bytes2Hex(pubData)
		logx.Info(str)

		ParseCreateCollectionPubData(res)
	}

}
func ParseCreateCollectionPubData(pubData []byte) (tx *txtypes.CreateCollectionTxInfo, err error) {
	offset := 0
	offset, txType := common2.ReadUint8(pubData, offset)
	if types.TxTypeCreateCollection == txType {
		offset, accountIndex := common2.ReadUint32(pubData, offset)
		offset, collectionId := common2.ReadUint16(pubData, offset)
		offset, gasFeeAssetId := common2.ReadUint16(pubData, offset)
		offset, gasFeeAssetAmount := common2.ReadUint16(pubData, offset)
		gasFeeAssetAmountBigInt, _ := util.CleanPackedFee(big.NewInt(int64(gasFeeAssetAmount)))

		tx = &txtypes.CreateCollectionTxInfo{
			//TxType:                 txType,
			AccountIndex:      int64(accountIndex),
			GasFeeAssetId:     int64(gasFeeAssetId),
			GasFeeAssetAmount: gasFeeAssetAmountBigInt,
			CollectionId:      int64(collectionId),
		}
	}

	return tx, nil
}
