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
	"errors"
	"math/big"

	bsmt "github.com/bnb-chain/bas-smt"
	"github.com/bnb-chain/zkbas-crypto/legend/circuit/bn254/std"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"

	common2 "github.com/bnb-chain/zkbas/common"
	"github.com/bnb-chain/zkbas/common/chain"
	"github.com/bnb-chain/zkbas/tree"
	"github.com/bnb-chain/zkbas/types"
)

type WitnessHelper struct {
	treeCtx *tree.Context

	accountModel AccountModel

	// Trees
	accountTree   bsmt.SparseMerkleTree
	assetTrees    *[]bsmt.SparseMerkleTree
	liquidityTree bsmt.SparseMerkleTree
	nftTree       bsmt.SparseMerkleTree
}

func NewWitnessHelper(treeCtx *tree.Context, accountTree, liquidityTree, nftTree bsmt.SparseMerkleTree,
	assetTrees *[]bsmt.SparseMerkleTree, accountModel AccountModel) *WitnessHelper {
	return &WitnessHelper{
		treeCtx:       treeCtx,
		accountModel:  accountModel,
		accountTree:   accountTree,
		assetTrees:    assetTrees,
		liquidityTree: liquidityTree,
		nftTree:       nftTree,
	}
}

func (w *WitnessHelper) ConstructCryptoTx(oTx *Tx, finalityBlockNr uint64,
) (cryptoTx *CryptoTx, err error) {
	switch oTx.TxType {
	case types.TxTypeEmpty:
		logx.Error("there should be no empty tx")
	default:
		cryptoTx, err = w.constructCryptoTx(oTx, finalityBlockNr)
		if err != nil {
			logx.Errorf("unable to construct crypto tx: %x", err.Error())
			return nil, err
		}
	}
	return cryptoTx, nil
}

func (w *WitnessHelper) constructCryptoTx(oTx *Tx, finalityBlockNr uint64) (cryptoTx *CryptoTx, err error) {
	if oTx == nil || w.accountTree == nil || w.assetTrees == nil || w.liquidityTree == nil || w.nftTree == nil {
		logx.Errorf("failed because of nil tx or tree")
		return nil, errors.New("failed because of nil tx or tree")
	}
	accountKeys, proverAccounts, proverLiquidityInfo, proverNftInfo, err := ConstructProverInfo(oTx, w.accountModel)
	if err != nil {
		logx.Errorf("unable to construct prover info: %s", err.Error())
		return nil, err
	}
	cryptoTx, err = w.constructWitnessInfo(
		oTx,
		finalityBlockNr,
		accountKeys,
		proverAccounts,
		proverLiquidityInfo,
		proverNftInfo,
	)
	if err != nil {
		logx.Errorf("unable to construct witness info: %s", err.Error())
		return nil, err
	}
	cryptoTx.TxType = uint8(oTx.TxType)
	cryptoTx.Nonce = oTx.Nonce
	switch oTx.TxType {
	case types.TxTypeRegisterZns:
		return w.constructRegisterZnsCryptoTx(cryptoTx, oTx)
	case types.TxTypeCreatePair:
		return w.constructCreatePairCryptoTx(cryptoTx, oTx)
	case types.TxTypeUpdatePairRate:
		return w.constructUpdatePairRateCryptoTx(cryptoTx, oTx)
	case types.TxTypeDeposit:
		return w.constructDepositCryptoTx(cryptoTx, oTx)
	case types.TxTypeDepositNft:
		return w.constructDepositNftCryptoTx(cryptoTx, oTx)
	case types.TxTypeTransfer:
		return w.constructTransferCryptoTx(cryptoTx, oTx)
	case types.TxTypeSwap:
		return w.constructSwapCryptoTx(cryptoTx, oTx)
	case types.TxTypeAddLiquidity:
		return w.constructAddLiquidityCryptoTx(cryptoTx, oTx)
	case types.TxTypeRemoveLiquidity:
		return w.constructRemoveLiquidityCryptoTx(cryptoTx, oTx)
	case types.TxTypeWithdraw:
		return w.constructWithdrawCryptoTx(cryptoTx, oTx)
	case types.TxTypeCreateCollection:
		return w.constructCreateCollectionCryptoTx(cryptoTx, oTx)
	case types.TxTypeMintNft:
		return w.constructMintNftCryptoTx(cryptoTx, oTx)
	case types.TxTypeTransferNft:
		return w.constructTransferNftCryptoTx(cryptoTx, oTx)
	case types.TxTypeAtomicMatch:
		return w.constructAtomicMatchCryptoTx(cryptoTx, oTx)
	case types.TxTypeCancelOffer:
		return w.constructCancelOfferCryptoTx(cryptoTx, oTx)
	case types.TxTypeWithdrawNft:
		return w.constructWithdrawNftCryptoTx(cryptoTx, oTx)
	case types.TxTypeFullExit:
		return w.constructFullExitCryptoTx(cryptoTx, oTx)
	case types.TxTypeFullExitNft:
		return w.constructFullExitNftCryptoTx(cryptoTx, oTx)
	default:
		return nil, errors.New("tx type error")
	}
}

func (w *WitnessHelper) constructWitnessInfo(
	oTx *Tx,
	finalityBlockNr uint64,
	accountKeys []int64,
	proverAccounts []*ProverAccountInfo,
	proverLiquidityInfo *ProverLiquidityInfo,
	proverNftInfo *ProverNftInfo,
) (
	cryptoTx *CryptoTx,
	err error,
) {
	// construct account witness
	AccountRootBefore, AccountsInfoBefore, MerkleProofsAccountAssetsBefore, MerkleProofsAccountBefore, err :=
		w.constructAccountWitness(oTx, finalityBlockNr, accountKeys, proverAccounts)
	if err != nil {
		logx.Errorf("unable to construct account witness: %s", err.Error())
		return nil, err
	}
	// construct liquidity witness
	LiquidityRootBefore, LiquidityBefore, MerkleProofsLiquidityBefore, err :=
		w.constructLiquidityWitness(proverLiquidityInfo)
	if err != nil {
		logx.Errorf("unable to construct liquidity witness: %s", err.Error())
		return nil, err
	}
	// construct nft witness
	NftRootBefore, NftBefore, MerkleProofsNftBefore, err :=
		w.ConstructNftWitness(proverNftInfo)
	if err != nil {
		logx.Errorf("unable to construct nft witness: %s", err.Error())
		return nil, err
	}
	stateRootBefore := tree.ComputeStateRootHash(AccountRootBefore, LiquidityRootBefore, NftRootBefore)
	stateRootAfter := tree.ComputeStateRootHash(w.accountTree.Root(), w.liquidityTree.Root(), w.nftTree.Root())
	cryptoTx = &CryptoTx{
		AccountRootBefore:               AccountRootBefore,
		AccountsInfoBefore:              AccountsInfoBefore,
		LiquidityRootBefore:             LiquidityRootBefore,
		LiquidityBefore:                 LiquidityBefore,
		NftRootBefore:                   NftRootBefore,
		NftBefore:                       NftBefore,
		StateRootBefore:                 stateRootBefore,
		MerkleProofsAccountAssetsBefore: MerkleProofsAccountAssetsBefore,
		MerkleProofsAccountBefore:       MerkleProofsAccountBefore,
		MerkleProofsLiquidityBefore:     MerkleProofsLiquidityBefore,
		MerkleProofsNftBefore:           MerkleProofsNftBefore,
		StateRootAfter:                  stateRootAfter,
	}
	return cryptoTx, nil
}

func (w *WitnessHelper) constructAccountWitness(
	oTx *Tx,
	finalityBlockNr uint64,
	accountKeys []int64,
	proverAccounts []*ProverAccountInfo,
) (
	AccountRootBefore []byte,
	// account before info, size is 5
	AccountsInfoBefore [NbAccountsPerTx]*CryptoAccount,
	// before account asset merkle proof
	MerkleProofsAccountAssetsBefore [NbAccountsPerTx][NbAccountAssetsPerAccount][AssetMerkleLevels][]byte,
	// before account merkle proof
	MerkleProofsAccountBefore [NbAccountsPerTx][AccountMerkleLevels][]byte,
	err error,
) {
	AccountRootBefore = w.accountTree.Root()
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
			logx.Errorf("unable to build merkle proofs: %s", err.Error())
			return AccountRootBefore, AccountsInfoBefore, MerkleProofsAccountAssetsBefore, MerkleProofsAccountBefore, err
		}
		// it means this is a registerZNS tx
		if proverAccounts == nil {
			if accountKey != int64(len(*w.assetTrees)) {
				logx.Errorf("invalid key")
				return AccountRootBefore, AccountsInfoBefore, MerkleProofsAccountAssetsBefore, MerkleProofsAccountBefore,
					errors.New("invalid key")
			}
			emptyAccountAssetTree, err := tree.NewEmptyAccountAssetTree(w.treeCtx, accountKey, finalityBlockNr)
			if err != nil {
				logx.Errorf("unable to create empty account asset tree: %s", err.Error())
				return AccountRootBefore, AccountsInfoBefore, MerkleProofsAccountAssetsBefore, MerkleProofsAccountBefore, err
			}
			*w.assetTrees = append(*w.assetTrees, emptyAccountAssetTree)
			cryptoAccount = std.EmptyAccount(accountKey, tree.NilAccountAssetRoot)
			// update account info
			accountInfo, err := w.accountModel.GetConfirmedAccountByIndex(accountKey)
			if err != nil {
				logx.Errorf("unable to get confirmed account by account index: %s", err.Error())
				return AccountRootBefore, AccountsInfoBefore, MerkleProofsAccountAssetsBefore, MerkleProofsAccountBefore, err
			}
			proverAccounts = append(proverAccounts, &ProverAccountInfo{
				AccountInfo: &Account{
					AccountIndex:    accountInfo.AccountIndex,
					AccountName:     accountInfo.AccountName,
					PublicKey:       accountInfo.PublicKey,
					AccountNameHash: accountInfo.AccountNameHash,
					L1Address:       accountInfo.L1Address,
					Nonce:           types.NilNonce,
					CollectionNonce: types.NilCollectionId,
					AssetInfo:       types.NilAssetInfo,
					AssetRoot:       common.Bytes2Hex(tree.NilAccountAssetRoot),
					Status:          accountInfo.Status,
				},
			})
		} else {
			proverAccountInfo := proverAccounts[accountCount]
			pk, err := common2.ParsePubKey(proverAccountInfo.AccountInfo.PublicKey)
			if err != nil {
				logx.Errorf("unable to parse pub key: %s", err.Error())
				return AccountRootBefore, AccountsInfoBefore, MerkleProofsAccountAssetsBefore, MerkleProofsAccountBefore, err
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
					logx.Errorf("unable to build merkle proofs: %s", err.Error())
					return AccountRootBefore, AccountsInfoBefore, MerkleProofsAccountAssetsBefore, MerkleProofsAccountBefore, err
				}
				// set crypto account asset
				cryptoAccount.AssetsInfo[assetCount] = &CryptoAccountAsset{
					AssetId:                  accountAsset.AssetId,
					Balance:                  accountAsset.Balance,
					LpAmount:                 accountAsset.LpAmount,
					OfferCanceledOrFinalized: accountAsset.OfferCanceledOrFinalized,
				}

				// set merkle proof
				MerkleProofsAccountAssetsBefore[accountCount][assetCount], err = SetFixedAccountAssetArray(assetMerkleProof)
				if err != nil {
					logx.Errorf("unable to set fixed merkle proof: %s", err.Error())
					return AccountRootBefore, AccountsInfoBefore, MerkleProofsAccountAssetsBefore, MerkleProofsAccountBefore, err
				}
				// update asset merkle tree
				nBalance, err := chain.ComputeNewBalance(
					proverAccountInfo.AssetsRelatedTxDetails[i].AssetType,
					proverAccountInfo.AssetsRelatedTxDetails[i].Balance,
					proverAccountInfo.AssetsRelatedTxDetails[i].BalanceDelta,
				)
				if err != nil {
					logx.Errorf("unable to compute new balance: %s", err.Error())
					return AccountRootBefore, AccountsInfoBefore, MerkleProofsAccountAssetsBefore, MerkleProofsAccountBefore, err
				}
				nAsset, err := types.ParseAccountAsset(nBalance)
				if err != nil {
					logx.Errorf("unable to parse account asset: %s", err.Error())
					return AccountRootBefore, AccountsInfoBefore, MerkleProofsAccountAssetsBefore, MerkleProofsAccountBefore, err
				}
				nAssetHash, err := tree.ComputeAccountAssetLeafHash(nAsset.Balance.String(), nAsset.LpAmount.String(), nAsset.OfferCanceledOrFinalized.String())
				if err != nil {
					logx.Errorf("unable to compute account asset leaf hash: %s", err.Error())
					return AccountRootBefore, AccountsInfoBefore, MerkleProofsAccountAssetsBefore, MerkleProofsAccountBefore, err
				}
				err = (*w.assetTrees)[accountKey].Set(uint64(accountAsset.AssetId), nAssetHash)
				if err != nil {
					logx.Errorf("unable to update asset tree: %s", err.Error())
					return AccountRootBefore, AccountsInfoBefore, MerkleProofsAccountAssetsBefore, MerkleProofsAccountBefore, err
				}

				assetCount++
			}
			if err != nil {
				logx.Errorf("unable to commit asset tree: %s", err.Error())
				return AccountRootBefore, AccountsInfoBefore, MerkleProofsAccountAssetsBefore, MerkleProofsAccountBefore, err
			}
		}
		// padding empty account asset
		for assetCount < NbAccountAssetsPerAccount {
			cryptoAccount.AssetsInfo[assetCount] = std.EmptyAccountAsset(LastAccountAssetId)
			assetMerkleProof, err := (*w.assetTrees)[accountKey].GetProof(LastAccountAssetId)
			if err != nil {
				logx.Errorf("unable to build merkle proofs: %s", err.Error())
				return AccountRootBefore, AccountsInfoBefore, MerkleProofsAccountAssetsBefore, MerkleProofsAccountBefore, err
			}
			MerkleProofsAccountAssetsBefore[accountCount][assetCount], err = SetFixedAccountAssetArray(assetMerkleProof)
			if err != nil {
				logx.Errorf("unable to set fixed merkle proof: %s", err.Error())
				return AccountRootBefore, AccountsInfoBefore, MerkleProofsAccountAssetsBefore, MerkleProofsAccountBefore, err
			}
			assetCount++
		}
		// set account merkle proof
		MerkleProofsAccountBefore[accountCount], err = SetFixedAccountArray(accountMerkleProofs)
		if err != nil {
			logx.Errorf("unable to set fixed merkle proof: %s", err.Error())
			return AccountRootBefore, AccountsInfoBefore, MerkleProofsAccountAssetsBefore, MerkleProofsAccountBefore, err
		}
		// update account merkle tree
		nonce := cryptoAccount.Nonce
		collectionNonce := cryptoAccount.CollectionNonce
		if oTx.AccountIndex == accountKey && oTx.Nonce != types.NilNonce {
			nonce = oTx.Nonce
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
			logx.Errorf("unable to compute account leaf hash: %s", err.Error())
			return AccountRootBefore, AccountsInfoBefore, MerkleProofsAccountAssetsBefore, MerkleProofsAccountBefore, err
		}
		err = w.accountTree.Set(uint64(accountKey), nAccountHash)
		if err != nil {
			logx.Errorf("unable to update account tree: %s", err.Error())
			return AccountRootBefore, AccountsInfoBefore, MerkleProofsAccountAssetsBefore, MerkleProofsAccountBefore, err
		}
		// set account info before
		AccountsInfoBefore[accountCount] = cryptoAccount
		// add count
		accountCount++
	}
	if err != nil {
		logx.Errorf("unable to commit account tree: %s", err.Error())
		return AccountRootBefore, AccountsInfoBefore, MerkleProofsAccountAssetsBefore, MerkleProofsAccountBefore, err
	}
	// padding empty account
	emptyAssetTree, err := tree.NewMemAccountAssetTree()
	if err != nil {
		logx.Errorf("unable to new empty account asset tree: %s", err.Error())
		return AccountRootBefore, AccountsInfoBefore, MerkleProofsAccountAssetsBefore, MerkleProofsAccountBefore, err
	}
	for accountCount < NbAccountsPerTx {
		AccountsInfoBefore[accountCount] = std.EmptyAccount(LastAccountIndex, tree.NilAccountAssetRoot)
		// get account before
		accountMerkleProofs, err := w.accountTree.GetProof(LastAccountIndex)
		if err != nil {
			logx.Errorf("unable to build merkle proofs: %s", err.Error())
			return AccountRootBefore, AccountsInfoBefore, MerkleProofsAccountAssetsBefore, MerkleProofsAccountBefore, err
		}
		// set account merkle proof
		MerkleProofsAccountBefore[accountCount], err = SetFixedAccountArray(accountMerkleProofs)
		if err != nil {
			logx.Errorf("unable to set fixed merkle proof: %s", err.Error())
			return AccountRootBefore, AccountsInfoBefore, MerkleProofsAccountAssetsBefore, MerkleProofsAccountBefore, err
		}
		for i := 0; i < NbAccountAssetsPerAccount; i++ {
			assetMerkleProof, err := emptyAssetTree.GetProof(0)
			if err != nil {
				logx.Errorf("unable to build merkle proofs: %s", err.Error())
				return AccountRootBefore, AccountsInfoBefore, MerkleProofsAccountAssetsBefore, MerkleProofsAccountBefore, err
			}
			MerkleProofsAccountAssetsBefore[accountCount][i], err = SetFixedAccountAssetArray(assetMerkleProof)
			if err != nil {
				logx.Errorf("unable to set fixed merkle proof: %s", err.Error())
				return AccountRootBefore, AccountsInfoBefore, MerkleProofsAccountAssetsBefore, MerkleProofsAccountBefore, err
			}
		}
		accountCount++

	}
	return AccountRootBefore, AccountsInfoBefore, MerkleProofsAccountAssetsBefore, MerkleProofsAccountBefore, nil
}

func (w *WitnessHelper) constructLiquidityWitness(
	proverLiquidityInfo *ProverLiquidityInfo,
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
			logx.Errorf("unable to build merkle proofs: %s", err.Error())
			return LiquidityRootBefore, LiquidityBefore, MerkleProofsLiquidityBefore, err
		}
		MerkleProofsLiquidityBefore, err = SetFixedLiquidityArray(liquidityMerkleProofs)
		if err != nil {
			logx.Errorf("unable to set fixed liquidity array: %s", err.Error())
			return LiquidityRootBefore, LiquidityBefore, MerkleProofsLiquidityBefore, err
		}
		LiquidityBefore = std.EmptyLiquidity(LastPairIndex)
		return LiquidityRootBefore, LiquidityBefore, MerkleProofsLiquidityBefore, nil
	}
	liquidityMerkleProofs, err := w.liquidityTree.GetProof(uint64(proverLiquidityInfo.LiquidityInfo.PairIndex))
	if err != nil {
		logx.Errorf("unable to build merkle proofs: %s", err.Error())
		return LiquidityRootBefore, LiquidityBefore, MerkleProofsLiquidityBefore, err
	}
	MerkleProofsLiquidityBefore, err = SetFixedLiquidityArray(liquidityMerkleProofs)
	if err != nil {
		logx.Errorf("unable to set fixed liquidity array: %s", err.Error())
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
		logx.Errorf("unable to compute new balance: %s", err.Error())
		return LiquidityRootBefore, LiquidityBefore, MerkleProofsLiquidityBefore, err
	}
	nPoolInfo, err := types.ParseLiquidityInfo(nBalance)
	if err != nil {
		logx.Errorf("unable to parse pool info: %s", err.Error())
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
		logx.Errorf("unable to compute liquidity node hash: %s", err.Error())
		return LiquidityRootBefore, LiquidityBefore, MerkleProofsLiquidityBefore, err
	}
	err = w.liquidityTree.Set(uint64(proverLiquidityInfo.LiquidityInfo.PairIndex), nLiquidityHash)
	if err != nil {
		logx.Errorf("unable to update liquidity tree: %s", err.Error())
		return LiquidityRootBefore, LiquidityBefore, MerkleProofsLiquidityBefore, err
	}
	return LiquidityRootBefore, LiquidityBefore, MerkleProofsLiquidityBefore, nil
}

func (w *WitnessHelper) ConstructNftWitness(
	proverNftInfo *ProverNftInfo,
) (
	// nft root before
	NftRootBefore []byte,
	// nft before
	NftBefore *CryptoNft,
	// before nft tree merkle proof
	MerkleProofsNftBefore [NftMerkleLevels][]byte,
	err error,
) {
	NftRootBefore = w.nftTree.Root()
	if proverNftInfo == nil {
		liquidityMerkleProofs, err := w.nftTree.GetProof(LastNftIndex)
		if err != nil {
			logx.Errorf("unable to build merkle proofs: %s", err.Error())
			return NftRootBefore, NftBefore, MerkleProofsNftBefore, err
		}
		MerkleProofsNftBefore, err = SetFixedNftArray(liquidityMerkleProofs)
		if err != nil {
			logx.Errorf("unable to set fixed nft array: %s", err.Error())
			return NftRootBefore, NftBefore, MerkleProofsNftBefore, err
		}
		NftBefore = std.EmptyNft(LastNftIndex)
		return NftRootBefore, NftBefore, MerkleProofsNftBefore, nil
	}
	nftMerkleProofs, err := w.nftTree.GetProof(uint64(proverNftInfo.NftInfo.NftIndex))
	if err != nil {
		logx.Errorf("unable to build merkle proofs: %s", err.Error())
		return NftRootBefore, NftBefore, MerkleProofsNftBefore, err
	}
	MerkleProofsNftBefore, err = SetFixedNftArray(nftMerkleProofs)
	if err != nil {
		logx.Errorf("unable to set fixed liquidity array: %s", err.Error())
		return NftRootBefore, NftBefore, MerkleProofsNftBefore, err
	}
	nftL1TokenId, isValid := new(big.Int).SetString(proverNftInfo.NftInfo.NftL1TokenId, 10)
	if !isValid {
		logx.Errorf("unable to parse big int")
		return NftRootBefore, NftBefore, MerkleProofsNftBefore, errors.New("unable to parse big int")
	}
	NftBefore = &CryptoNft{
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
		logx.Errorf("unable to compute new balance: %s", err.Error())
		return NftRootBefore, NftBefore, MerkleProofsNftBefore, err
	}
	nNftInfo, err := types.ParseNftInfo(nBalance)
	if err != nil {
		logx.Errorf("unable to parse pool info: %s", err.Error())
		return NftRootBefore, NftBefore, MerkleProofsNftBefore, err
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
		logx.Errorf("unable to compute liquidity node hash: %s", err.Error())
		return NftRootBefore, NftBefore, MerkleProofsNftBefore, err
	}
	err = w.nftTree.Set(uint64(proverNftInfo.NftInfo.NftIndex), nNftHash)
	if err != nil {
		logx.Errorf("unable to update liquidity tree: %s", err.Error())
		return NftRootBefore, NftBefore, MerkleProofsNftBefore, err
	}
	if err != nil {
		logx.Errorf("unable to commit liquidity tree: %s", err.Error())
		return NftRootBefore, NftBefore, MerkleProofsNftBefore, err
	}
	return NftRootBefore, NftBefore, MerkleProofsNftBefore, nil
}

func SetFixedAccountArray(proof [][]byte) (res [AccountMerkleLevels][]byte, err error) {
	if len(proof) != AccountMerkleLevels {
		logx.Errorf("invalid size")
		return res, errors.New("invalid size")
	}
	copy(res[:], proof[:])
	return res, nil
}

func SetFixedAccountAssetArray(proof [][]byte) (res [AssetMerkleLevels][]byte, err error) {
	if len(proof) != AssetMerkleLevels {
		logx.Errorf("invalid size")
		return res, errors.New("invalid size")
	}
	copy(res[:], proof[:])
	return res, nil
}

func SetFixedLiquidityArray(proof [][]byte) (res [LiquidityMerkleLevels][]byte, err error) {
	if len(proof) != LiquidityMerkleLevels {
		logx.Errorf("invalid size")
		return res, errors.New("invalid size")
	}
	copy(res[:], proof[:])
	return res, nil
}

func SetFixedNftArray(proof [][]byte) (res [NftMerkleLevels][]byte, err error) {
	if len(proof) != NftMerkleLevels {
		logx.Errorf("invalid size")
		return res, errors.New("invalid size")
	}
	copy(res[:], proof[:])
	return res, nil
}
