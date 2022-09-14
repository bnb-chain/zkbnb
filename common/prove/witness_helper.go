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
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/bnb-chain/zkbnb-crypto/legend/circuit/bn254/std"
	bsmt "github.com/bnb-chain/zkbnb-smt"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/common/chain"
	"github.com/bnb-chain/zkbnb/tree"
	"github.com/bnb-chain/zkbnb/types"
)

type WitnessHelper struct {
	treeCtx *tree.Context

	accountModel AccountModel

	// Trees
	accountTree   bsmt.SparseMerkleTree
	assetTrees    *[]tree.LazyTreeWrapper
	liquidityTree bsmt.SparseMerkleTree
	nftTree       bsmt.SparseMerkleTree
}

func NewWitnessHelper(treeCtx *tree.Context, accountTree, liquidityTree, nftTree bsmt.SparseMerkleTree,
	assetTrees *[]tree.LazyTreeWrapper, accountModel AccountModel) *WitnessHelper {
	return &WitnessHelper{
		treeCtx:       treeCtx,
		accountModel:  accountModel,
		accountTree:   accountTree,
		assetTrees:    assetTrees,
		liquidityTree: liquidityTree,
		nftTree:       nftTree,
	}
}

func (w *WitnessHelper) ConstructTxWitness(oTx *Tx, finalityBlockNr uint64,
) (cryptoTx *TxWitness, err error) {
	switch oTx.TxType {
	case types.TxTypeEmpty:
		return nil, fmt.Errorf("there should be no empty tx")
	default:
		cryptoTx, err = w.constructTxWitness(oTx, finalityBlockNr)
		if err != nil {
			return nil, err
		}
	}
	return cryptoTx, nil
}

func (w *WitnessHelper) constructTxWitness(oTx *Tx, finalityBlockNr uint64) (witness *TxWitness, err error) {
	if oTx == nil || w.accountTree == nil || w.assetTrees == nil || w.liquidityTree == nil || w.nftTree == nil {
		return nil, fmt.Errorf("failed because of nil tx or tree")
	}
	witness, err = w.constructWitnessInfo(oTx, finalityBlockNr)
	if err != nil {
		return nil, err
	}
	witness.TxType = uint8(oTx.TxType)
	witness.Nonce = oTx.Nonce
	switch oTx.TxType {
	case types.TxTypeRegisterZns:
		return w.constructRegisterZnsTxWitness(witness, oTx)
	case types.TxTypeCreatePair:
		return w.constructCreatePairTxWitness(witness, oTx)
	case types.TxTypeUpdatePairRate:
		return w.constructUpdatePairRateTxWitness(witness, oTx)
	case types.TxTypeDeposit:
		return w.constructDepositTxWitness(witness, oTx)
	case types.TxTypeDepositNft:
		return w.constructDepositNftTxWitness(witness, oTx)
	case types.TxTypeTransfer:
		return w.constructTransferTxWitness(witness, oTx)
	case types.TxTypeSwap:
		return w.constructSwapTxWitness(witness, oTx)
	case types.TxTypeAddLiquidity:
		return w.constructAddLiquidityTxWitness(witness, oTx)
	case types.TxTypeRemoveLiquidity:
		return w.constructRemoveLiquidityTxWitness(witness, oTx)
	case types.TxTypeWithdraw:
		return w.constructWithdrawTxWitness(witness, oTx)
	case types.TxTypeCreateCollection:
		return w.constructCreateCollectionTxWitness(witness, oTx)
	case types.TxTypeMintNft:
		return w.constructMintNftTxWitness(witness, oTx)
	case types.TxTypeTransferNft:
		return w.constructTransferNftTxWitness(witness, oTx)
	case types.TxTypeAtomicMatch:
		return w.constructAtomicMatchTxWitness(witness, oTx)
	case types.TxTypeCancelOffer:
		return w.constructCancelOfferTxWitness(witness, oTx)
	case types.TxTypeWithdrawNft:
		return w.constructWithdrawNftTxWitness(witness, oTx)
	case types.TxTypeFullExit:
		return w.constructFullExitTxWitness(witness, oTx)
	case types.TxTypeFullExitNft:
		return w.constructFullExitNftTxWitness(witness, oTx)
	default:
		return nil, fmt.Errorf("tx type error")
	}
}

func (w *WitnessHelper) constructWitnessInfo(
	oTx *Tx,
	finalityBlockNr uint64,
) (
	cryptoTx *TxWitness,
	err error,
) {
	accountKeys, proverAccounts, proverLiquidityInfo, proverNftInfo, err := w.constructSimpleWitnessInfo(oTx)
	if err != nil {
		return nil, err
	}
	// construct account witness
	accountRootBefore, accountsInfoBefore, merkleProofsAccountAssetsBefore, merkleProofsAccountBefore, err :=
		w.constructAccountWitness(oTx, finalityBlockNr, accountKeys, proverAccounts)
	if err != nil {
		return nil, err
	}
	// construct liquidity witness
	liquidityRootBefore, liquidityBefore, merkleProofsLiquidityBefore, err :=
		w.constructLiquidityWitness(proverLiquidityInfo)
	if err != nil {
		return nil, err
	}
	// construct nft witness
	nftRootBefore, nftBefore, merkleProofsNftBefore, err :=
		w.constructNftWitness(proverNftInfo)
	if err != nil {
		return nil, err
	}
	stateRootBefore := tree.ComputeStateRootHash(accountRootBefore, liquidityRootBefore, nftRootBefore)
	stateRootAfter := tree.ComputeStateRootHash(w.accountTree.Root(), w.liquidityTree.Root(), w.nftTree.Root())
	cryptoTx = &TxWitness{
		AccountRootBefore:               accountRootBefore,
		AccountsInfoBefore:              accountsInfoBefore,
		LiquidityRootBefore:             liquidityRootBefore,
		LiquidityBefore:                 liquidityBefore,
		NftRootBefore:                   nftRootBefore,
		NftBefore:                       nftBefore,
		StateRootBefore:                 stateRootBefore,
		MerkleProofsAccountAssetsBefore: merkleProofsAccountAssetsBefore,
		MerkleProofsAccountBefore:       merkleProofsAccountBefore,
		MerkleProofsLiquidityBefore:     merkleProofsLiquidityBefore,
		MerkleProofsNftBefore:           merkleProofsNftBefore,
		StateRootAfter:                  stateRootAfter,
	}
	return cryptoTx, nil
}

func (w *WitnessHelper) constructAccountWitness(
	oTx *Tx,
	finalityBlockNr uint64,
	accountKeys []int64,
	proverAccounts []*AccountWitnessInfo,
) (
	accountRootBefore []byte,
	// account before info, size is 5
	accountsInfoBefore [NbAccountsPerTx]*CryptoAccount,
	// before account asset merkle proof
	merkleProofsAccountAssetsBefore [NbAccountsPerTx][NbAccountAssetsPerAccount][AssetMerkleLevels][]byte,
	// before account merkle proof
	merkleProofsAccountBefore [NbAccountsPerTx][AccountMerkleLevels][]byte,
	err error,
) {
	accountRootBefore = w.accountTree.Root()
	var (
		accountCount = 0
	)
	for _, accountKey := range accountKeys {
		var (
			cryptoAccount = new(CryptoAccount)
			// get account asset before
			assetCount = 0
		)
		// get account before
		accountMerkleProofs, err := w.accountTree.GetProof(uint64(accountKey))
		if err != nil {
			return accountRootBefore, accountsInfoBefore, merkleProofsAccountAssetsBefore, merkleProofsAccountBefore, err
		}
		// it means this is a registerZNS tx
		if proverAccounts == nil {
			if accountKey != int64(len(*w.assetTrees)) {
				return accountRootBefore, accountsInfoBefore, merkleProofsAccountAssetsBefore, merkleProofsAccountBefore,
					fmt.Errorf("invalid key")
			}
			emptyAccountAssetTree := tree.NewEmptyAccountAssetTree(w.treeCtx, accountKey, finalityBlockNr)
			// if err != nil {
			// 	return accountRootBefore, accountsInfoBefore, merkleProofsAccountAssetsBefore, merkleProofsAccountBefore, err
			// }
			*w.assetTrees = append(*w.assetTrees, emptyAccountAssetTree)
			cryptoAccount = std.EmptyAccount(accountKey, tree.NilAccountAssetRoot)
			// update account info
			accountInfo, err := w.accountModel.GetConfirmedAccountByIndex(accountKey)
			if err != nil {
				return accountRootBefore, accountsInfoBefore, merkleProofsAccountAssetsBefore, merkleProofsAccountBefore, err
			}
			proverAccounts = append(proverAccounts, &AccountWitnessInfo{
				AccountInfo: &Account{
					AccountIndex:    accountInfo.AccountIndex,
					AccountName:     accountInfo.AccountName,
					PublicKey:       accountInfo.PublicKey,
					AccountNameHash: accountInfo.AccountNameHash,
					L1Address:       accountInfo.L1Address,
					Nonce:           types.EmptyNonce,
					CollectionNonce: types.EmptyCollectionNonce,
					AssetInfo:       types.EmptyAccountAssetInfo,
					AssetRoot:       common.Bytes2Hex(tree.NilAccountAssetRoot),
					Status:          accountInfo.Status,
				},
			})
		} else {
			proverAccountInfo := proverAccounts[accountCount]
			pk, err := common2.ParsePubKey(proverAccountInfo.AccountInfo.PublicKey)
			if err != nil {
				return accountRootBefore, accountsInfoBefore, merkleProofsAccountAssetsBefore, merkleProofsAccountBefore, err
			}
			cryptoAccount = &CryptoAccount{
				AccountIndex:    accountKey,
				AccountNameHash: common.FromHex(proverAccountInfo.AccountInfo.AccountNameHash),
				AccountPk:       pk,
				Nonce:           proverAccountInfo.AccountInfo.Nonce,
				CollectionNonce: proverAccountInfo.AccountInfo.CollectionNonce,
				AssetRoot:       (*w.assetTrees)[accountKey].Root(),
			}
			for i, accountAsset := range proverAccountInfo.AccountAssets {
				assetMerkleProof, err := (*w.assetTrees)[accountKey].GetProof(uint64(accountAsset.AssetId))
				if err != nil {
					return accountRootBefore, accountsInfoBefore, merkleProofsAccountAssetsBefore, merkleProofsAccountBefore, err
				}
				// set crypto account asset
				cryptoAccount.AssetsInfo[assetCount] = &CryptoAccountAsset{
					AssetId:                  accountAsset.AssetId,
					Balance:                  accountAsset.Balance,
					LpAmount:                 accountAsset.LpAmount,
					OfferCanceledOrFinalized: accountAsset.OfferCanceledOrFinalized,
				}

				// set merkle proof
				merkleProofsAccountAssetsBefore[accountCount][assetCount], err = SetFixedAccountAssetArray(assetMerkleProof)
				if err != nil {
					return accountRootBefore, accountsInfoBefore, merkleProofsAccountAssetsBefore, merkleProofsAccountBefore, err
				}
				// update asset merkle tree
				nBalance, err := chain.ComputeNewBalance(
					proverAccountInfo.AssetsRelatedTxDetails[i].AssetType,
					proverAccountInfo.AssetsRelatedTxDetails[i].Balance,
					proverAccountInfo.AssetsRelatedTxDetails[i].BalanceDelta,
				)
				if err != nil {
					return accountRootBefore, accountsInfoBefore, merkleProofsAccountAssetsBefore, merkleProofsAccountBefore, err
				}
				nAsset, err := types.ParseAccountAsset(nBalance)
				if err != nil {
					return accountRootBefore, accountsInfoBefore, merkleProofsAccountAssetsBefore, merkleProofsAccountBefore, err
				}
				nAssetHash, err := tree.ComputeAccountAssetLeafHash(nAsset.Balance.String(), nAsset.LpAmount.String(), nAsset.OfferCanceledOrFinalized.String())
				if err != nil {
					return accountRootBefore, accountsInfoBefore, merkleProofsAccountAssetsBefore, merkleProofsAccountBefore, err
				}
				err = (*w.assetTrees)[accountKey].Set(uint64(accountAsset.AssetId), nAssetHash)
				if err != nil {
					return accountRootBefore, accountsInfoBefore, merkleProofsAccountAssetsBefore, merkleProofsAccountBefore, err
				}

				assetCount++
			}
			if err != nil {
				return accountRootBefore, accountsInfoBefore, merkleProofsAccountAssetsBefore, merkleProofsAccountBefore, err
			}
		}
		// padding empty account asset
		for assetCount < NbAccountAssetsPerAccount {
			cryptoAccount.AssetsInfo[assetCount] = std.EmptyAccountAsset(LastAccountAssetId)
			assetMerkleProof, err := (*w.assetTrees)[accountKey].GetProof(LastAccountAssetId)
			if err != nil {
				return accountRootBefore, accountsInfoBefore, merkleProofsAccountAssetsBefore, merkleProofsAccountBefore, err
			}
			merkleProofsAccountAssetsBefore[accountCount][assetCount], err = SetFixedAccountAssetArray(assetMerkleProof)
			if err != nil {
				return accountRootBefore, accountsInfoBefore, merkleProofsAccountAssetsBefore, merkleProofsAccountBefore, err
			}
			assetCount++
		}
		// set account merkle proof
		merkleProofsAccountBefore[accountCount], err = SetFixedAccountArray(accountMerkleProofs)
		if err != nil {
			return accountRootBefore, accountsInfoBefore, merkleProofsAccountAssetsBefore, merkleProofsAccountBefore, err
		}
		// update account merkle tree
		nonce := cryptoAccount.Nonce
		collectionNonce := cryptoAccount.CollectionNonce
		if oTx.AccountIndex == accountKey && types.IsL2Tx(oTx.TxType) {
			nonce = oTx.Nonce + 1 // increase nonce if tx is initiated in l2
		}
		if oTx.AccountIndex == accountKey && oTx.TxType == types.TxTypeCreateCollection {
			collectionNonce++
		}
		nAccountHash, err := tree.ComputeAccountLeafHash(
			proverAccounts[accountCount].AccountInfo.AccountNameHash,
			proverAccounts[accountCount].AccountInfo.PublicKey,
			nonce,
			collectionNonce,
			(*w.assetTrees)[accountKey].Root(),
		)
		if err != nil {
			return accountRootBefore, accountsInfoBefore, merkleProofsAccountAssetsBefore, merkleProofsAccountBefore, err
		}
		err = w.accountTree.Set(uint64(accountKey), nAccountHash)
		if err != nil {
			return accountRootBefore, accountsInfoBefore, merkleProofsAccountAssetsBefore, merkleProofsAccountBefore, err
		}
		// set account info before
		accountsInfoBefore[accountCount] = cryptoAccount
		// add count
		accountCount++
	}
	if err != nil {
		return accountRootBefore, accountsInfoBefore, merkleProofsAccountAssetsBefore, merkleProofsAccountBefore, err
	}
	// padding empty account
	emptyAssetTree, err := tree.NewMemAccountAssetTree()
	if err != nil {
		return accountRootBefore, accountsInfoBefore, merkleProofsAccountAssetsBefore, merkleProofsAccountBefore, err
	}
	for accountCount < NbAccountsPerTx {
		accountsInfoBefore[accountCount] = std.EmptyAccount(LastAccountIndex, tree.NilAccountAssetRoot)
		// get account before
		accountMerkleProofs, err := w.accountTree.GetProof(LastAccountIndex)
		if err != nil {
			return accountRootBefore, accountsInfoBefore, merkleProofsAccountAssetsBefore, merkleProofsAccountBefore, err
		}
		// set account merkle proof
		merkleProofsAccountBefore[accountCount], err = SetFixedAccountArray(accountMerkleProofs)
		if err != nil {
			return accountRootBefore, accountsInfoBefore, merkleProofsAccountAssetsBefore, merkleProofsAccountBefore, err
		}
		for i := 0; i < NbAccountAssetsPerAccount; i++ {
			assetMerkleProof, err := emptyAssetTree.GetProof(0)
			if err != nil {
				return accountRootBefore, accountsInfoBefore, merkleProofsAccountAssetsBefore, merkleProofsAccountBefore, err
			}
			merkleProofsAccountAssetsBefore[accountCount][i], err = SetFixedAccountAssetArray(assetMerkleProof)
			if err != nil {
				return accountRootBefore, accountsInfoBefore, merkleProofsAccountAssetsBefore, merkleProofsAccountBefore, err
			}
		}
		accountCount++

	}
	return accountRootBefore, accountsInfoBefore, merkleProofsAccountAssetsBefore, merkleProofsAccountBefore, nil
}

func (w *WitnessHelper) constructLiquidityWitness(
	proverLiquidityInfo *LiquidityWitnessInfo,
) (
	// liquidity root before
	LiquidityRootBefore []byte,
	// liquidity before
	LiquidityBefore *CryptoLiquidity,
	// before liquidity merkle proof
	MerkleProofsLiquidityBefore [LiquidityMerkleLevels][]byte,
	err error,
) {
	LiquidityRootBefore = w.liquidityTree.Root()
	if proverLiquidityInfo == nil {
		liquidityMerkleProofs, err := w.liquidityTree.GetProof(LastPairIndex)
		if err != nil {
			return LiquidityRootBefore, LiquidityBefore, MerkleProofsLiquidityBefore, err
		}
		MerkleProofsLiquidityBefore, err = SetFixedLiquidityArray(liquidityMerkleProofs)
		if err != nil {
			return LiquidityRootBefore, LiquidityBefore, MerkleProofsLiquidityBefore, err
		}
		LiquidityBefore = std.EmptyLiquidity(LastPairIndex)
		return LiquidityRootBefore, LiquidityBefore, MerkleProofsLiquidityBefore, nil
	}
	liquidityMerkleProofs, err := w.liquidityTree.GetProof(uint64(proverLiquidityInfo.LiquidityInfo.PairIndex))
	if err != nil {
		return LiquidityRootBefore, LiquidityBefore, MerkleProofsLiquidityBefore, err
	}
	MerkleProofsLiquidityBefore, err = SetFixedLiquidityArray(liquidityMerkleProofs)
	if err != nil {
		return LiquidityRootBefore, LiquidityBefore, MerkleProofsLiquidityBefore, err
	}
	LiquidityBefore = &CryptoLiquidity{
		PairIndex:            proverLiquidityInfo.LiquidityInfo.PairIndex,
		AssetAId:             proverLiquidityInfo.LiquidityInfo.AssetAId,
		AssetA:               proverLiquidityInfo.LiquidityInfo.AssetA,
		AssetBId:             proverLiquidityInfo.LiquidityInfo.AssetBId,
		AssetB:               proverLiquidityInfo.LiquidityInfo.AssetB,
		LpAmount:             proverLiquidityInfo.LiquidityInfo.LpAmount,
		KLast:                proverLiquidityInfo.LiquidityInfo.KLast,
		FeeRate:              proverLiquidityInfo.LiquidityInfo.FeeRate,
		TreasuryAccountIndex: proverLiquidityInfo.LiquidityInfo.TreasuryAccountIndex,
		TreasuryRate:         proverLiquidityInfo.LiquidityInfo.TreasuryRate,
	}
	// update liquidity tree
	nBalance, err := chain.ComputeNewBalance(
		proverLiquidityInfo.LiquidityRelatedTxDetail.AssetType,
		proverLiquidityInfo.LiquidityRelatedTxDetail.Balance,
		proverLiquidityInfo.LiquidityRelatedTxDetail.BalanceDelta,
	)
	if err != nil {
		return LiquidityRootBefore, LiquidityBefore, MerkleProofsLiquidityBefore, err
	}
	nPoolInfo, err := types.ParseLiquidityInfo(nBalance)
	if err != nil {
		return LiquidityRootBefore, LiquidityBefore, MerkleProofsLiquidityBefore, err
	}
	nLiquidityHash, err := tree.ComputeLiquidityAssetLeafHash(
		nPoolInfo.AssetAId,
		nPoolInfo.AssetA.String(),
		nPoolInfo.AssetBId,
		nPoolInfo.AssetB.String(),
		nPoolInfo.LpAmount.String(),
		nPoolInfo.KLast.String(),
		nPoolInfo.FeeRate,
		nPoolInfo.TreasuryAccountIndex,
		nPoolInfo.TreasuryRate,
	)
	if err != nil {
		return LiquidityRootBefore, LiquidityBefore, MerkleProofsLiquidityBefore, err
	}
	err = w.liquidityTree.Set(uint64(proverLiquidityInfo.LiquidityInfo.PairIndex), nLiquidityHash)
	if err != nil {
		return LiquidityRootBefore, LiquidityBefore, MerkleProofsLiquidityBefore, err
	}
	return LiquidityRootBefore, LiquidityBefore, MerkleProofsLiquidityBefore, nil
}

func (w *WitnessHelper) constructNftWitness(
	proverNftInfo *NftWitnessInfo,
) (
	// nft root before
	nftRootBefore []byte,
	// nft before
	nftBefore *CryptoNft,
	// before nft tree merkle proof
	merkleProofsNftBefore [NftMerkleLevels][]byte,
	err error,
) {
	nftRootBefore = w.nftTree.Root()
	if proverNftInfo == nil {
		liquidityMerkleProofs, err := w.nftTree.GetProof(LastNftIndex)
		if err != nil {
			return nftRootBefore, nftBefore, merkleProofsNftBefore, err
		}
		merkleProofsNftBefore, err = SetFixedNftArray(liquidityMerkleProofs)
		if err != nil {
			return nftRootBefore, nftBefore, merkleProofsNftBefore, err
		}
		nftBefore = std.EmptyNft(LastNftIndex)
		return nftRootBefore, nftBefore, merkleProofsNftBefore, nil
	}
	nftMerkleProofs, err := w.nftTree.GetProof(uint64(proverNftInfo.NftInfo.NftIndex))
	if err != nil {
		return nftRootBefore, nftBefore, merkleProofsNftBefore, err
	}
	merkleProofsNftBefore, err = SetFixedNftArray(nftMerkleProofs)
	if err != nil {
		return nftRootBefore, nftBefore, merkleProofsNftBefore, err
	}
	nftL1TokenId, isValid := new(big.Int).SetString(proverNftInfo.NftInfo.NftL1TokenId, 10)
	if !isValid {
		return nftRootBefore, nftBefore, merkleProofsNftBefore, fmt.Errorf("unable to parse big int")
	}
	nftBefore = &CryptoNft{
		NftIndex:            proverNftInfo.NftInfo.NftIndex,
		NftContentHash:      common.FromHex(proverNftInfo.NftInfo.NftContentHash),
		CreatorAccountIndex: proverNftInfo.NftInfo.CreatorAccountIndex,
		OwnerAccountIndex:   proverNftInfo.NftInfo.OwnerAccountIndex,
		NftL1Address:        new(big.Int).SetBytes(common.FromHex(proverNftInfo.NftInfo.NftL1Address)),
		NftL1TokenId:        nftL1TokenId,
		CreatorTreasuryRate: proverNftInfo.NftInfo.CreatorTreasuryRate,
		CollectionId:        proverNftInfo.NftInfo.CollectionId,
	}
	// update liquidity tree
	nBalance, err := chain.ComputeNewBalance(
		proverNftInfo.NftRelatedTxDetail.AssetType,
		proverNftInfo.NftRelatedTxDetail.Balance,
		proverNftInfo.NftRelatedTxDetail.BalanceDelta,
	)
	if err != nil {
		return nftRootBefore, nftBefore, merkleProofsNftBefore, err
	}
	nNftInfo, err := types.ParseNftInfo(nBalance)
	if err != nil {
		return nftRootBefore, nftBefore, merkleProofsNftBefore, err
	}
	nNftHash, err := tree.ComputeNftAssetLeafHash(
		nNftInfo.CreatorAccountIndex,
		nNftInfo.OwnerAccountIndex,
		nNftInfo.NftContentHash,
		nNftInfo.NftL1Address,
		nNftInfo.NftL1TokenId,
		nNftInfo.CreatorTreasuryRate,
		nNftInfo.CollectionId,
	)

	if err != nil {
		return nftRootBefore, nftBefore, merkleProofsNftBefore, err
	}
	err = w.nftTree.Set(uint64(proverNftInfo.NftInfo.NftIndex), nNftHash)
	if err != nil {
		return nftRootBefore, nftBefore, merkleProofsNftBefore, err
	}
	if err != nil {
		return nftRootBefore, nftBefore, merkleProofsNftBefore, err
	}
	return nftRootBefore, nftBefore, merkleProofsNftBefore, nil
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

func (w *WitnessHelper) constructSimpleWitnessInfo(oTx *Tx) (
	accountKeys []int64,
	accountWitnessInfo []*AccountWitnessInfo,
	liquidityWitnessInfo *LiquidityWitnessInfo,
	nftWitnessInfo *NftWitnessInfo,
	err error,
) {
	var (
		// dbinitializer account asset map, because if we have the same asset detail, the before will be the after of the last one
		accountAssetMap  = make(map[int64]map[int64]*AccountAsset)
		accountMap       = make(map[int64]*Account)
		lastAccountOrder = int64(-2)
		accountCount     = -1
	)
	// dbinitializer prover account map
	if oTx.TxType == types.TxTypeRegisterZns {
		accountKeys = append(accountKeys, oTx.AccountIndex)
	}
	for _, txDetail := range oTx.TxDetails {
		switch txDetail.AssetType {
		case types.FungibleAssetType:
			// get account info
			if accountMap[txDetail.AccountIndex] == nil {
				accountInfo, err := w.accountModel.GetConfirmedAccountByIndex(txDetail.AccountIndex)
				if err != nil {
					return nil, nil, nil, nil, err
				}
				// get current nonce
				accountInfo.Nonce = txDetail.Nonce
				accountMap[txDetail.AccountIndex] = accountInfo
			} else {
				if lastAccountOrder != txDetail.AccountOrder {
					if oTx.AccountIndex == txDetail.AccountIndex {
						accountMap[txDetail.AccountIndex].Nonce = oTx.Nonce + 1
					}
				}
			}
			if lastAccountOrder != txDetail.AccountOrder {
				accountKeys = append(accountKeys, txDetail.AccountIndex)
				lastAccountOrder = txDetail.AccountOrder
				accountWitnessInfo = append(accountWitnessInfo, &AccountWitnessInfo{
					AccountInfo: &Account{
						AccountIndex:    accountMap[txDetail.AccountIndex].AccountIndex,
						AccountName:     accountMap[txDetail.AccountIndex].AccountName,
						PublicKey:       accountMap[txDetail.AccountIndex].PublicKey,
						AccountNameHash: accountMap[txDetail.AccountIndex].AccountNameHash,
						L1Address:       accountMap[txDetail.AccountIndex].L1Address,
						Nonce:           accountMap[txDetail.AccountIndex].Nonce,
						CollectionNonce: txDetail.CollectionNonce,
						AssetInfo:       accountMap[txDetail.AccountIndex].AssetInfo,
						AssetRoot:       accountMap[txDetail.AccountIndex].AssetRoot,
						Status:          accountMap[txDetail.AccountIndex].Status,
					},
				})
				accountCount++
			}
			if accountAssetMap[txDetail.AccountIndex] == nil {
				accountAssetMap[txDetail.AccountIndex] = make(map[int64]*AccountAsset)
			}
			if accountAssetMap[txDetail.AccountIndex][txDetail.AssetId] == nil {
				// set account before info
				oAsset, err := types.ParseAccountAsset(txDetail.Balance)
				if err != nil {
					return nil, nil, nil, nil, err
				}
				accountWitnessInfo[accountCount].AccountAssets = append(
					accountWitnessInfo[accountCount].AccountAssets,
					oAsset,
				)
			} else {
				// set account before info
				accountWitnessInfo[accountCount].AccountAssets = append(
					accountWitnessInfo[accountCount].AccountAssets,
					&AccountAsset{
						AssetId:                  accountAssetMap[txDetail.AccountIndex][txDetail.AssetId].AssetId,
						Balance:                  accountAssetMap[txDetail.AccountIndex][txDetail.AssetId].Balance,
						LpAmount:                 accountAssetMap[txDetail.AccountIndex][txDetail.AssetId].LpAmount,
						OfferCanceledOrFinalized: accountAssetMap[txDetail.AccountIndex][txDetail.AssetId].OfferCanceledOrFinalized,
					},
				)
			}
			// set tx detail
			accountWitnessInfo[accountCount].AssetsRelatedTxDetails = append(
				accountWitnessInfo[accountCount].AssetsRelatedTxDetails,
				txDetail,
			)
			// update asset info
			newBalance, err := chain.ComputeNewBalance(txDetail.AssetType, txDetail.Balance, txDetail.BalanceDelta)
			if err != nil {
				return nil, nil, nil, nil, err
			}
			nAsset, err := types.ParseAccountAsset(newBalance)
			if err != nil {
				return nil, nil, nil, nil, err
			}
			accountAssetMap[txDetail.AccountIndex][txDetail.AssetId] = nAsset
		case types.LiquidityAssetType:
			liquidityWitnessInfo = new(LiquidityWitnessInfo)
			liquidityWitnessInfo.LiquidityRelatedTxDetail = txDetail
			poolInfo, err := types.ParseLiquidityInfo(txDetail.Balance)
			if err != nil {
				return nil, nil, nil, nil, err
			}
			liquidityWitnessInfo.LiquidityInfo = poolInfo
		case types.NftAssetType:
			nftWitnessInfo = new(NftWitnessInfo)
			nftWitnessInfo.NftRelatedTxDetail = txDetail
			nftInfo, err := types.ParseNftInfo(txDetail.Balance)
			if err != nil {
				return nil, nil, nil, nil, err
			}
			nftWitnessInfo.NftInfo = nftInfo
		case types.CollectionNonceAssetType:
			// get account info
			if accountMap[txDetail.AccountIndex] == nil {
				accountInfo, err := w.accountModel.GetConfirmedAccountByIndex(txDetail.AccountIndex)
				if err != nil {
					return nil, nil, nil, nil, err
				}
				// get current nonce
				accountInfo.Nonce = txDetail.Nonce
				accountInfo.CollectionNonce = txDetail.CollectionNonce
				accountMap[txDetail.AccountIndex] = accountInfo
				if lastAccountOrder != txDetail.AccountOrder {
					accountKeys = append(accountKeys, txDetail.AccountIndex)
					lastAccountOrder = txDetail.AccountOrder
					accountWitnessInfo = append(accountWitnessInfo, &AccountWitnessInfo{
						AccountInfo: &Account{
							AccountIndex:    accountMap[txDetail.AccountIndex].AccountIndex,
							AccountName:     accountMap[txDetail.AccountIndex].AccountName,
							PublicKey:       accountMap[txDetail.AccountIndex].PublicKey,
							AccountNameHash: accountMap[txDetail.AccountIndex].AccountNameHash,
							L1Address:       accountMap[txDetail.AccountIndex].L1Address,
							Nonce:           accountMap[txDetail.AccountIndex].Nonce,
							CollectionNonce: txDetail.CollectionNonce,
							AssetInfo:       accountMap[txDetail.AccountIndex].AssetInfo,
							AssetRoot:       accountMap[txDetail.AccountIndex].AssetRoot,
							Status:          accountMap[txDetail.AccountIndex].Status,
						},
					})
					accountCount++
				}
			} else {
				accountMap[txDetail.AccountIndex].Nonce = txDetail.Nonce
				accountMap[txDetail.AccountIndex].CollectionNonce = txDetail.CollectionNonce
			}
		default:
			return nil, nil, nil, nil,
				fmt.Errorf("invalid asset type")
		}
	}
	return accountKeys, accountWitnessInfo, liquidityWitnessInfo, nftWitnessInfo, nil
}
