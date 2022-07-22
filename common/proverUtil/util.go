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
	"github.com/bnb-chain/bas-smt/database"
	cryptoBlock "github.com/bnb-chain/zkbas-crypto/legend/circuit/bn254/block"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/pkg/treedb"
)

func SetFixedAccountArray(proof [][]byte) (res [AccountMerkleLevels][]byte, err error) {
	if len(proof) != AccountMerkleLevels {
		logx.Errorf("[SetFixedAccountArray] invalid size")
		return res, errors.New("[SetFixedAccountArray] invalid size")
	}
	copy(res[:], proof[:])
	return res, nil
}

func SetFixedAccountAssetArray(proof [][]byte) (res [AssetMerkleLevels][]byte, err error) {
	if len(proof) != AssetMerkleLevels {
		logx.Errorf("[SetFixedAccountAssetArray] invalid size")
		return res, errors.New("[SetFixedAccountAssetArray] invalid size")
	}
	copy(res[:], proof[:])
	return res, nil
}

func SetFixedLiquidityArray(proof [][]byte) (res [LiquidityMerkleLevels][]byte, err error) {
	if len(proof) != LiquidityMerkleLevels {
		logx.Errorf("[SetFixedLiquidityArray] invalid size")
		return res, errors.New("[SetFixedLiquidityArray] invalid size")
	}
	copy(res[:], proof[:])
	return res, nil
}

func SetFixedNftArray(proof [][]byte) (res [NftMerkleLevels][]byte, err error) {
	if len(proof) != NftMerkleLevels {
		logx.Errorf("[SetFixedNftArray] invalid size")
		return res, errors.New("[SetFixedNftArray] invalid size")
	}
	copy(res[:], proof[:])
	return res, nil
}

func ConstructCryptoTx(
	oTx *Tx,
	treeDBDriver treedb.Driver,
	treeDB database.TreeDB,
	accountTree bsmt.SparseMerkleTree,
	assetTrees *[]bsmt.SparseMerkleTree,
	liquidityTree bsmt.SparseMerkleTree,
	nftTree bsmt.SparseMerkleTree,
	accountModel AccountModel,
) (cryptoTx *CryptoTx, err error) {
	switch oTx.TxType {
	case commonTx.TxTypeEmpty:
		logx.Error("[ConstructProverBlocks] there should be no empty tx")
		break
	case commonTx.TxTypeRegisterZns:
		cryptoTx, err = ConstructRegisterZnsCryptoTx(
			oTx,
			treeDBDriver,
			treeDB,
			accountTree,
			assetTrees,
			liquidityTree,
			nftTree,
			accountModel,
		)
		if err != nil {
			logx.Errorf("[ConstructProverBlocks] unable to construct registerZNS crypto tx: %x", err.Error())
			return nil, err
		}
		break
	case commonTx.TxTypeCreatePair:
		cryptoTx, err = ConstructCreatePairCryptoTx(
			oTx,
			treeDBDriver,
			treeDB,
			accountTree,
			assetTrees,
			liquidityTree,
			nftTree,
			accountModel,
		)
		if err != nil {
			logx.Errorf("[ConstructProverBlocks] unable to construct create pair crypto tx: %s", err.Error())
			return nil, err
		}
		break
	case commonTx.TxTypeUpdatePairRate:
		cryptoTx, err = ConstructUpdatePairRateCryptoTx(
			oTx,
			treeDBDriver,
			treeDB,
			accountTree,
			assetTrees,
			liquidityTree,
			nftTree,
			accountModel,
		)
		if err != nil {
			logx.Errorf("[ConstructProverBlocks] unable to construct update pair crypto tx: %s", err.Error())
			return nil, err
		}
		break
	case commonTx.TxTypeDeposit:
		cryptoTx, err = ConstructDepositCryptoTx(
			oTx,
			treeDBDriver,
			treeDB,
			accountTree,
			assetTrees,
			liquidityTree,
			nftTree,
			accountModel,
		)
		if err != nil {
			logx.Errorf("[ConstructProverBlocks] unable to construct deposit crypto tx: %s", err.Error())
			return nil, err
		}
		break
	case commonTx.TxTypeDepositNft:
		cryptoTx, err = ConstructDepositNftCryptoTx(
			oTx,
			treeDBDriver,
			treeDB,
			accountTree,
			assetTrees,
			liquidityTree,
			nftTree,
			accountModel,
		)
		if err != nil {
			logx.Errorf("[ConstructProverBlocks] unable to construct deposit nft crypto tx: %s", err.Error())
			return nil, err
		}
		break
	case commonTx.TxTypeTransfer:
		cryptoTx, err = ConstructTransferCryptoTx(
			oTx,
			treeDBDriver,
			treeDB,
			accountTree,
			assetTrees,
			liquidityTree,
			nftTree,
			accountModel,
		)
		if err != nil {
			logx.Errorf("[ConstructProverBlocks] unable to construct transfer crypto tx: %s", err.Error())
			return nil, err
		}
		break
	case commonTx.TxTypeSwap:
		cryptoTx, err = ConstructSwapCryptoTx(
			oTx,
			treeDBDriver,
			treeDB,
			accountTree,
			assetTrees,
			liquidityTree,
			nftTree,
			accountModel,
		)
		if err != nil {
			logx.Errorf("[ConstructProverBlocks] unable to construct swap crypto tx: %s", err.Error())
			return nil, err
		}
		break
	case commonTx.TxTypeAddLiquidity:
		cryptoTx, err = ConstructAddLiquidityCryptoTx(
			oTx,
			treeDBDriver,
			treeDB,
			accountTree,
			assetTrees,
			liquidityTree,
			nftTree,
			accountModel,
		)
		if err != nil {
			logx.Errorf("[ConstructProverBlocks] unable to construct add liquidity crypto tx: %s", err.Error())
			return nil, err
		}
		break
	case commonTx.TxTypeRemoveLiquidity:
		cryptoTx, err = ConstructRemoveLiquidityCryptoTx(
			oTx,
			treeDBDriver,
			treeDB,
			accountTree,
			assetTrees,
			liquidityTree,
			nftTree,
			accountModel,
		)
		if err != nil {
			logx.Errorf("[ConstructProverBlocks] unable to construct remove liquidity crypto tx: %s", err.Error())
			return nil, err
		}
		break
	case commonTx.TxTypeWithdraw:
		cryptoTx, err = ConstructWithdrawCryptoTx(
			oTx,
			treeDBDriver,
			treeDB,
			accountTree,
			assetTrees,
			liquidityTree,
			nftTree,
			accountModel,
		)
		if err != nil {
			logx.Errorf("[ConstructProverBlocks] unable to construct withdraw crypto tx: %s", err.Error())
			return nil, err
		}
		break
	case commonTx.TxTypeCreateCollection:
		cryptoTx, err = ConstructCreateCollectionCryptoTx(
			oTx,
			treeDBDriver,
			treeDB,
			accountTree,
			assetTrees,
			liquidityTree,
			nftTree,
			accountModel,
		)
		if err != nil {
			logx.Errorf("[ConstructProverBlocks] unable to construct create collection crypto tx: %s", err.Error())
			return nil, err
		}
		break
	case commonTx.TxTypeMintNft:
		cryptoTx, err = ConstructMintNftCryptoTx(
			oTx,
			treeDBDriver,
			treeDB,
			accountTree,
			assetTrees,
			liquidityTree,
			nftTree,
			accountModel,
		)
		if err != nil {
			logx.Errorf("[ConstructProverBlocks] unable to construct mint nft crypto tx: %s", err.Error())
			return nil, err
		}
		break
	case commonTx.TxTypeTransferNft:
		cryptoTx, err = ConstructTransferNftCryptoTx(
			oTx,
			treeDBDriver,
			treeDB,
			accountTree,
			assetTrees,
			liquidityTree,
			nftTree,
			accountModel,
		)
		if err != nil {
			logx.Errorf("[ConstructProverBlocks] unable to construct transfer nft crypto tx: %s", err.Error())
			return nil, err
		}
		break
	case commonTx.TxTypeAtomicMatch:
		cryptoTx, err = ConstructAtomicMatchCryptoTx(
			oTx,
			treeDBDriver,
			treeDB,
			accountTree,
			assetTrees,
			liquidityTree,
			nftTree,
			accountModel,
		)
		if err != nil {
			logx.Errorf("[ConstructProverBlocks] unable to construct atomic match crypto tx: %s", err.Error())
			return nil, err
		}
		break
	case commonTx.TxTypeCancelOffer:
		cryptoTx, err = ConstructCancelOfferCryptoTx(
			oTx,
			treeDBDriver,
			treeDB,
			accountTree,
			assetTrees,
			liquidityTree,
			nftTree,
			accountModel,
		)
		if err != nil {
			logx.Errorf("[ConstructProverBlocks] unable to construct cancel offer crypto tx: %s", err.Error())
			return nil, err
		}
		break
	case commonTx.TxTypeWithdrawNft:
		cryptoTx, err = ConstructWithdrawNftCryptoTx(
			oTx,
			treeDBDriver,
			treeDB,
			accountTree,
			assetTrees,
			liquidityTree,
			nftTree,
			accountModel,
		)
		if err != nil {
			logx.Errorf("[ConstructProverBlocks] unable to construct withdraw nft crypto tx: %s", err.Error())
			return nil, err
		}
		break
	case commonTx.TxTypeFullExit:
		cryptoTx, err = ConstructFullExitCryptoTx(
			oTx,
			treeDBDriver,
			treeDB,
			accountTree,
			assetTrees,
			liquidityTree,
			nftTree,
			accountModel,
		)
		if err != nil {
			logx.Errorf("[ConstructProverBlocks] unable to construct full exit crypto tx: %s", err.Error())
			return nil, err
		}
		break
	case commonTx.TxTypeFullExitNft:
		cryptoTx, err = ConstructFullExitNftCryptoTx(
			oTx,
			treeDBDriver,
			treeDB,
			accountTree,
			assetTrees,
			liquidityTree,
			nftTree,
			accountModel,
		)
		if err != nil {
			logx.Errorf("[ConstructProverBlocks] unable to construct full exit nft crypto tx: %s", err.Error())
			return nil, err
		}
		break
	default:
		return nil, errors.New("tx type error")
	}
	return cryptoTx, nil
}

func BlockToCryptoBlock(
	oBlock *block.Block,
	oldStateRoot, newStateRoot []byte,
	cryptoTxs []*cryptoBlock.Tx,
) (cBlock *cryptoBlock.Block, err error) {
	cBlock = &cryptoBlock.Block{
		BlockNumber:     oBlock.BlockHeight,
		CreatedAt:       oBlock.CreatedAt.UnixMilli(),
		OldStateRoot:    oldStateRoot,
		NewStateRoot:    newStateRoot,
		BlockCommitment: common.FromHex(oBlock.BlockCommitment),
	}
	for i := 0; i < len(cryptoTxs); i++ {
		cBlock.Txs = append(cBlock.Txs, cryptoTxs[i])
	}
	return cBlock, nil
}
