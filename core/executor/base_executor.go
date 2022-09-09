package executor

import (
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb-crypto/wasm/legend/legendTxTypes"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/types"
)

const (
	OfferPerAsset = 128
	TenThousand   = 10000
)

type BaseExecutor struct {
	bc      IBlockchain
	tx      *tx.Tx
	iTxInfo legendTxTypes.TxInfo

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

func (e *BaseExecutor) Prepare() error {
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

		gasAccountIndex, gasFeeAssetId, _ := txInfo.GetGas()
		err = e.bc.VerifyGas(gasAccountIndex, gasFeeAssetId)
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

func (e *BaseExecutor) GenerateTxDetails() ([]*tx.TxDetail, error) {
	return nil, nil
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
