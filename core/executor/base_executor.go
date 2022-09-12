package executor

import (
	"context"
	"fmt"
	"github.com/bnb-chain/zkbnb-crypto/legend/circuit/bn254/block"
	"github.com/bnb-chain/zkbnb-crypto/legend/circuit/bn254/std"
	"github.com/bnb-chain/zkbnb-crypto/wasm/legend/legendTxTypes"
	bsmt "github.com/bnb-chain/zkbnb-smt"
	"github.com/bnb-chain/zkbnb/common/prove"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/tree"
	"github.com/bnb-chain/zkbnb/types"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	OfferPerAsset = 128
	TenThousand   = 10000
)

type BaseExecutor struct {
	bc      IBlockchain
	tx      *tx.Tx
	iTxInfo legendTxTypes.TxInfo

	witnessKeys *legendTxTypes.TxWitnessKeys

	// Affected states.
	dirtyAccountsAndAssetsMap map[int64]map[int64]bool
	dirtyLiquidityMap         map[int64]bool
	dirtyNftMap               map[int64]bool
}

func NewBaseExecutor(bc IBlockchain, tx *tx.Tx, txInfo legendTxTypes.TxInfo) BaseExecutor {
	return BaseExecutor{
		bc:      bc,
		tx:      tx,
		iTxInfo: txInfo,

		dirtyAccountsAndAssetsMap: make(map[int64]map[int64]bool, 0),
		dirtyLiquidityMap:         make(map[int64]bool, 0),
		dirtyNftMap:               make(map[int64]bool, 0),
	}
}

func (e *BaseExecutor) Prepare(ctx context.Context) error {
	// Mark the tree states that would be affected in this executor.
	e.witnessKeys = e.iTxInfo.WitnessKeys(ctx)

	if e.witnessKeys.NftIndex != legendTxTypes.LastNftIndex {
		e.MarkNftDirty(e.witnessKeys.NftIndex)
	}
	if e.witnessKeys.PairIndex != legendTxTypes.LastNftIndex {
		e.MarkLiquidityDirty(e.witnessKeys.PairIndex)
	}
	for _, account := range e.witnessKeys.Accounts {
		e.MarkAccountAssetsDirty(account.Index, account.Assets)
	}
	err := e.bc.StateDB().PrepareAccountsAndAssets(e.dirtyAccountsAndAssetsMap)
	if err != nil {
		logx.Errorf("prepare accounts and assets failed: %s", err.Error())
		return errors.New("internal error")
	}
	return nil
}

func (e *BaseExecutor) VerifyInputs() error {
	txInfo := e.iTxInfo

	err := txInfo.Validate()
	if err != nil {
		return err
	}
	err = e.bc.VerifyExpiredAt(txInfo.GetExpiredAt())
	if err != nil {
		return err
	}

	from := txInfo.GetFromAccountIndex()
	if from != types.NilAccountIndex {
		err = e.bc.VerifyNonce(from, txInfo.GetNonce())
		if err != nil {
			return err
		}

		err = txInfo.VerifySignature(e.bc.StateDB().AccountMap[from].PublicKey)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *BaseExecutor) ApplyTransaction() error {
	e.SyncDirtyToStateCache()
	return nil
}

func (e *BaseExecutor) GeneratePubData() error {
	return nil
}

func (e *BaseExecutor) GetExecutedTx() (*tx.Tx, error) {
	e.tx.BlockHeight = e.bc.CurrentBlock().BlockHeight
	e.tx.TxStatus = tx.StatusSuccess
	e.tx.TxIndex = int64(len(e.bc.StateDB().Txs))
	return e.tx, nil
}

func (e *BaseExecutor) GenerateWitness() (witness *prove.TxWitness, err error) {
	keys := e.witnessKeys
	witness = &prove.TxWitness{}

	// Assign NFT released
	witness.NftBefore = std.EmptyNft(legendTxTypes.LastNftIndex)
	if keys.NftIndex != legendTxTypes.LastNftIndex {
		if nft, exist := e.bc.StateDB().NftMap[keys.NftIndex]; exist {
			witness.NftBefore, err = nft.ToStdNFT()
			if err != nil {
				return nil, err
			}
		} else {
			witness.NftBefore = std.EmptyNft(e.tx.NftIndex)
		}
	}
	nftMerkleProofs, err := e.bc.StateDB().NftTree.GetProof(uint64(witness.NftBefore.NftIndex))
	if err != nil {
		return nil, err
	}
	witness.MerkleProofsNftBefore, err = prove.SetFixedNftArray(nftMerkleProofs)
	if err != nil {
		return nil, err
	}
	witness.NftRootBefore = e.bc.StateDB().NftTree.Root()

	// Assign liquidity related
	witness.LiquidityBefore = std.EmptyLiquidity(legendTxTypes.LastNftIndex)
	if keys.PairIndex != legendTxTypes.LastPairIndex {
		if pair, exist := e.bc.StateDB().LiquidityMap[keys.PairIndex]; exist {
			witness.LiquidityBefore = pair.ToStdLiquidity()
		} else {
			witness.LiquidityBefore = std.EmptyLiquidity(e.tx.PairIndex)
		}
	}
	liquidityMerkleProofs, err := e.bc.StateDB().LiquidityTree.GetProof(uint64(witness.LiquidityBefore.PairIndex))
	if err != nil {
		return nil, err
	}
	witness.MerkleProofsLiquidityBefore, err = prove.SetFixedLiquidityArray(liquidityMerkleProofs)
	if err != nil {
		return nil, err
	}
	witness.LiquidityRootBefore = e.bc.StateDB().LiquidityTree.Root()

	// Assign account related
	for index, ak := range keys.Accounts {
		account := e.bc.StateDB().AccountMap[ak.Index]
		if account == nil {
			// register zks
			if ak.Index != int64(len(e.bc.StateDB().AccountAssetTrees)) {
				return nil, fmt.Errorf("invalid key")
			}
			witness.AccountsInfoBefore[index] = std.EmptyAccount(ak.Index, tree.NilAccountAssetRoot)
		} else {
			ac, err := account.ToStdAccount()
			if err != nil {
				return nil, err
			}
			for idx, assetKey := range ak.Assets {
				if as, exist := account.AssetInfo[assetKey]; exist {
					ac.AssetsInfo[idx] = &std.AccountAsset{
						AssetId:                  as.AssetId,
						Balance:                  as.Balance,
						LpAmount:                 as.LpAmount,
						OfferCanceledOrFinalized: as.OfferCanceledOrFinalized,
					}
				} else {
					ac.AssetsInfo[idx] = std.EmptyAccountAsset(assetKey)
				}
			}
			for idx := len(ak.Assets); idx < std.NbAccountAssetsPerAccount; idx++ {
				ac.AssetsInfo[idx] = std.EmptyAccountAsset(block.LastAccountAssetId)
			}
			witness.AccountsInfoBefore[index] = ac
		}
	}

	for index := len(keys.Accounts); index < std.NbAccountsPerTx; index++ {
		witness.AccountsInfoBefore[index] = std.EmptyAccount(block.LastAccountIndex, tree.NilAccountAssetRoot)
	}

	for _, ac := range witness.AccountsInfoBefore {
		accountMerkleProofs, err := e.bc.StateDB().AccountTree.GetProof(uint64(ac.AccountIndex))
		if err != nil {
			return nil, err
		}
		witness.MerkleProofsAccountBefore[ac.AccountIndex], err = prove.SetFixedAccountArray(accountMerkleProofs)
		if err != nil {
			return nil, err
		}
		var assetTree bsmt.SparseMerkleTree
		if ac.AccountIndex >= int64(len(e.bc.StateDB().AccountAssetTrees)) {
			assetTree, err = tree.NewEmptyAccountAssetTree(e.bc.StateDB().TreeCtx, ac.AccountIndex)
			if err != nil {
				return nil, err
			}
		} else {
			assetTree = (e.bc.StateDB().AccountAssetTrees)[ac.AccountIndex]
		}

		for idx, asset := range ac.AssetsInfo {
			assetMerkleProof, err := assetTree.GetProof(uint64(asset.AssetId))
			if err != nil {
				return nil, err
			}
			witness.MerkleProofsAccountAssetsBefore[ac.AccountIndex][idx], err = prove.SetFixedAccountAssetArray(assetMerkleProof)
		}
		if err != nil {
			return nil, err
		}
	}

	witness.AccountRootBefore = e.bc.StateDB().AccountTree.Root()
	witness.StateRootBefore = tree.ComputeStateRootHash(witness.AccountRootBefore, witness.LiquidityRootBefore, witness.NftRootBefore)
	witness.TxType = uint8(e.tx.TxType)
	witness.Nonce = e.tx.Nonce
	witness.ExpiredAt = e.tx.ExpiredAt
	return witness, nil
}

func (e *BaseExecutor) MarkAccountAssetsDirty(accountIndex int64, assets []int64) {
	if accountIndex < 0 {
		return
	}

	_, ok := e.dirtyAccountsAndAssetsMap[accountIndex]
	if !ok {
		e.dirtyAccountsAndAssetsMap[accountIndex] = make(map[int64]bool, 0)
	}

	for _, assetIndex := range assets {
		// Should never happen, but protect here.
		if assetIndex < 0 {
			continue
		}
		e.dirtyAccountsAndAssetsMap[accountIndex][assetIndex] = true
	}
}

func (e *BaseExecutor) MarkLiquidityDirty(pairIndex int64) {
	e.dirtyLiquidityMap[pairIndex] = true
}

func (e *BaseExecutor) MarkNftDirty(nftIndex int64) {
	e.dirtyNftMap[nftIndex] = true
}

func (e *BaseExecutor) SyncDirtyToStateCache() {
	for accountIndex, assetsMap := range e.dirtyAccountsAndAssetsMap {
		assets := make([]int64, 0, len(assetsMap))
		for assetIndex := range assetsMap {
			assets = append(assets, assetIndex)
		}
		e.bc.StateDB().MarkAccountAssetsDirty(accountIndex, assets)
	}

	for pairIndex := range e.dirtyLiquidityMap {
		e.bc.StateDB().MarkLiquidityDirty(pairIndex)
	}

	for nftIndex := range e.dirtyNftMap {
		e.bc.StateDB().MarkNftDirty(nftIndex)
	}
}
