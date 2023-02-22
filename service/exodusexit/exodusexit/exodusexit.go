package exodusexit

import (
	"fmt"
	types2 "github.com/bnb-chain/zkbnb-crypto/circuit/types"
	"github.com/bnb-chain/zkbnb-crypto/util"
	"github.com/bnb-chain/zkbnb-crypto/wasm/txtypes"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/core/executor"
	"github.com/bnb-chain/zkbnb/core/statedb"
	"github.com/bnb-chain/zkbnb/dao/exodusexit"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/common/chain"
	"github.com/bnb-chain/zkbnb/core"
	"github.com/bnb-chain/zkbnb/types"
)

type Config struct {
	core.ChainConfig

	BlockConfig struct {
		OptionalBlockSizes    []int
		SaveBlockDataPoolSize int  `json:",optional"`
		RollbackOnly          bool `json:",optional"`
	}
	LogConf logx.LogConf
}

type ExodusExit struct {
	running bool
	config  *Config
	bc      *core.BlockChain
}

func NewExodusExit(config *Config) (*ExodusExit, error) {
	bc, err := core.NewBlockChainForExodusExit(&config.ChainConfig)
	if err != nil {
		return nil, fmt.Errorf("new blockchain error: %v", err)
	}
	committer := &ExodusExit{
		running: true,
		config:  config,
		bc:      bc,
	}
	return committer, nil
}

func (c *ExodusExit) Run() error {
	c.loadAllAccounts()
	c.loadAllNfts()
	limit := 1000
	executedBlock, err := c.bc.ExodusExitBlockModel.GetLatestExecutedBlock()
	if err != nil && err != types.DbErrNotFound {
		logx.Errorf("get executed tx from exodus exit block failed:%s", err.Error())
		panic("get executed tx from exodus exit block failed: " + err.Error())
	}

	var executedTxMaxHeight int64 = 0
	if executedBlock != nil {
		executedTxMaxHeight = executedBlock.BlockHeight
	}
	for {
		if !c.running {
			break
		}
		pendingBlocks, err := c.bc.ExodusExitBlockModel.GetBlocksByStatusAndMaxHeight(exodusexit.StatusVerified, executedTxMaxHeight, int64(limit))
		if err != nil {
			logx.Errorf("get pending blocks from exodus exit block failed:%s", err.Error())
			time.Sleep(100 * time.Millisecond)
			continue
		}
		if len(pendingBlocks) == 0 {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		for _, pendingBlock := range pendingBlocks {
			if int(pendingBlock.BlockHeight)-int(executedTxMaxHeight) != 1 {
				time.Sleep(50 * time.Millisecond)
				logx.Infof("not equal block height=%s", pendingBlock.BlockHeight)
				break
			}
			err := c.executeBlockFunc(pendingBlock)
			if err != nil {
				return err
			}
			err = c.SaveToDb(pendingBlock)
			if err != nil {
				return err
			}
		}
	}
}

func (c *ExodusExit) executeBlockFunc(exodusExitBlock *exodusexit.ExodusExitBlock) error {
	pubData := common.FromHex(exodusExitBlock.PubData)
	sizePerTx := types2.PubDataBitsSizePerTx / 8
	c.bc.Statedb.PurgeCache("")
	for i := 0; i < int(exodusExitBlock.BlockSize); i++ {
		subPubData := pubData[i*sizePerTx : sizePerTx]
		offset := 0
		offset, txType := common2.ReadUint8(subPubData, offset)
		switch txType {
		case types.TxTypeTransfer:
			c.executeTransfer(subPubData)
			break
		}
	}
	return nil

}

func (c *ExodusExit) SaveToDb(exodusExitBlock *exodusexit.ExodusExitBlock) error {
	logx.Infof("SaveToDb start, blockHeight:%d", exodusExitBlock.BlockHeight)
	stateDataCopy := &statedb.StateDataCopy{
		StateCache:   c.bc.Statedb.StateCache,
		CurrentBlock: nil,
	}
	pendingAccounts, _, err := c.bc.Statedb.GetPendingAccount(exodusExitBlock.BlockHeight, stateDataCopy)
	if err != nil {
		return err
	}

	pendingNfts, _, err := c.bc.Statedb.GetPendingNft(exodusExitBlock.BlockHeight, stateDataCopy)
	if err != nil {
		return err
	}
	// update db
	err = c.bc.DB().DB.Transaction(func(tx *gorm.DB) error {
		err := c.bc.DB().ExodusExitBlockModel.UpdateBlockToExecutedInTransact(tx, exodusExitBlock)
		if err != nil {
			return err
		}
		err = c.bc.DB().AccountModel.BatchInsertOrUpdateInTransact(tx, pendingAccounts)
		if err != nil {
			return err
		}
		err = c.bc.DB().L2NftModel.BatchInsertOrUpdateInTransact(tx, pendingNfts)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		logx.Errorf("SaveToDb failed:%s,blockHeight:%d", err.Error(), exodusExitBlock.BlockHeight)
	}
	return nil
}

func (c *ExodusExit) loadAllAccounts() {
	start := time.Now()
	logx.Infof("load all accounts start")
	totalTask := 0
	errChan := make(chan error, 1)
	defer close(errChan)
	batchReloadSize := 1000
	maxAccountIndex, err := c.bc.AccountModel.GetMaxAccountIndex()
	if err != nil && err != types.DbErrNotFound {
		logx.Severef("load all accounts failed:%s", err.Error())
		panic("load all accounts failed: " + err.Error())
	}
	if maxAccountIndex == -1 {
		return
	}
	for i := 0; int64(i) <= maxAccountIndex; i += batchReloadSize {
		toAccountIndex := int64(i+batchReloadSize) - 1
		if toAccountIndex > maxAccountIndex {
			toAccountIndex = maxAccountIndex
		}
		totalTask++
		err := func(fromAccountIndex int64, toAccountIndex int64) error {
			return c.pool.Submit(func() {
				start := time.Now()
				accounts, err := c.bc.AccountModel.GetByAccountIndexRange(fromAccountIndex, toAccountIndex)
				if err != nil && err != types.DbErrNotFound {
					logx.Severef("load all accounts failed:%s", err.Error())
					errChan <- err
					return
				}
				if accounts != nil {
					for _, accountInfo := range accounts {
						formatAccount, err := chain.ToFormatAccountInfo(accountInfo)
						if err != nil {
							logx.Severef("load all accounts failed:%s", err.Error())
							errChan <- err
							return
						}
						c.bc.Statedb.AccountCache.Add(accountInfo.AccountIndex, formatAccount)
					}
				}
				logx.Infof("GetByNftIndexRange cost time %s", float64(time.Since(start).Milliseconds()))
				errChan <- nil
			})
		}(int64(i), toAccountIndex)
		if err != nil {
			logx.Severef("load all accounts failed:%s", err.Error())
			panic("load all accounts failed: " + err.Error())
		}
	}

	for i := 0; i < totalTask; i++ {
		err := <-errChan
		if err != nil {
			logx.Severef("load all accounts failed:%s", err.Error())
			panic("load all accounts failed: " + err.Error())
		}
	}
	logx.Infof("load all accounts end. cost time %s", float64(time.Since(start).Milliseconds()))
}

func (c *ExodusExit) loadAllNfts() {
	start := time.Now()
	logx.Infof("load all nfts start")
	totalTask := 0
	errChan := make(chan error, 1)
	defer close(errChan)
	batchReloadSize := 1000
	maxNftIndex, err := c.bc.L2NftModel.GetMaxNftIndex()
	if err != nil && err != types.DbErrNotFound {
		logx.Severef("load all nfts failed:%s", err.Error())
		panic("load all nfts failed: " + err.Error())
	}
	if maxNftIndex == -1 {
		return
	}
	for i := 0; int64(i) <= maxNftIndex; i += batchReloadSize {
		toNftIndex := int64(i+batchReloadSize) - 1
		if toNftIndex > maxNftIndex {
			toNftIndex = maxNftIndex
		}
		totalTask++
		err := func(fromNftIndex int64, toNftIndex int64) error {
			return c.pool.Submit(func() {
				start := time.Now()
				nfts, err := c.bc.L2NftModel.GetByNftIndexRange(fromNftIndex, toNftIndex)
				if err != nil && err != types.DbErrNotFound {
					logx.Severef("load all nfts failed:%s", err.Error())
					errChan <- err
					return
				}
				if nfts != nil {
					for _, nftInfo := range nfts {
						c.bc.Statedb.NftCache.Add(nftInfo.NftIndex, nftInfo)
					}
				}
				logx.Infof("GetByNftIndexRange cost time %s", float64(time.Since(start).Milliseconds()))
				errChan <- nil
			})
		}(int64(i), toNftIndex)
		if err != nil {
			logx.Severef("load all nfts failed:%s", err.Error())
			panic("load all nfts failed: " + err.Error())
		}
	}

	for i := 0; i < totalTask; i++ {
		err := <-errChan
		if err != nil {
			logx.Severef("load all nfts failed:%s", err.Error())
			panic("load all nfts failed: " + err.Error())
		}
	}
	logx.Infof("load all nfts end. cost time %s", float64(time.Since(start).Milliseconds()))
}

func (c *ExodusExit) Shutdown() {
	c.running = false
	c.bc.Statedb.Close()
	c.bc.ChainDB.Close()
}

func (c *ExodusExit) executeTransfer(pubData []byte) error {
	bc := c.bc
	offset := 1
	offset, fromAccountIndex := common2.ReadUint32(pubData, offset)
	offset, toAccountIndex := common2.ReadUint32(pubData, offset)
	offset, assetId := common2.ReadUint16(pubData, offset)
	offset, packedAmount := common2.ReadUint40(pubData, offset)
	assetAmount, err := util.CleanPackedAmount(big.NewInt(packedAmount))
	if err != nil {
		return err
	}
	offset, gasFeeAssetId := common2.ReadUint16(pubData, offset)
	offset, packedFee := common2.ReadUint16(pubData, offset)
	gasFeeAssetAmount, err := util.CleanPackedFee(big.NewInt(int64(packedFee)))
	if err != nil {
		return err
	}

	txInfo := &txtypes.TransferTxInfo{
		FromAccountIndex:  int64(fromAccountIndex),
		ToAccountIndex:    int64(toAccountIndex),
		AssetId:           int64(assetId),
		AssetAmount:       assetAmount,
		GasFeeAssetId:     int64(gasFeeAssetId),
		GasFeeAssetAmount: gasFeeAssetAmount,
	}
	executor := &executor.TransferExecutor{
		BaseExecutor: executor.NewBaseExecutor(bc, nil, txInfo),
		TxInfo:       txInfo,
	}
	err = executor.ApplyTransaction()
	if err != nil {
		return err
	}
	//fromAccount, err := bc.StateDB().GetFormatAccount(txInfo.FromAccountIndex)
	//if err != nil {
	//	return err
	//}
	//toAccount, err := bc.StateDB().GetFormatAccount(txInfo.ToAccountIndex)
	//if err != nil {
	//	return err
	//}
	//
	//fromAccount.AssetInfo[txInfo.AssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.AssetId].Balance, txInfo.AssetAmount)
	//fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	//toAccount.AssetInfo[txInfo.AssetId].Balance = ffmath.Add(toAccount.AssetInfo[txInfo.AssetId].Balance, txInfo.AssetAmount)
	//fromAccount.Nonce++
	//
	//stateCache := bc.StateDB()
	//stateCache.SetPendingAccount(txInfo.FromAccountIndex, fromAccount)
	//stateCache.SetPendingAccount(txInfo.ToAccountIndex, toAccount)
	//stateCache.SetPendingGas(txInfo.GasFeeAssetId, txInfo.GasFeeAssetAmount)
	return nil
}

func (c *ExodusExit) executeCollection(pubData []byte) error {
	bc := c.bc
	offset := 1
	offset, accountIndex := common2.ReadUint32(pubData, offset)
	offset, collectionId := common2.ReadUint16(pubData, offset)
	offset, gasFeeAssetId := common2.ReadUint16(pubData, offset)
	offset, gasFeeAssetAmount := common2.ReadUint16(pubData, offset)
	gasFeeAssetAmountBigInt, _ := util.CleanPackedFee(big.NewInt(int64(gasFeeAssetAmount)))

	txInfo := &txtypes.CreateCollectionTxInfo{
		AccountIndex:      int64(accountIndex),
		GasFeeAssetId:     int64(gasFeeAssetId),
		GasFeeAssetAmount: gasFeeAssetAmountBigInt,
		CollectionId:      int64(collectionId),
	}

	executor := &executor.CreateCollectionExecutor{
		BaseExecutor: executor.NewBaseExecutor(bc, nil, txInfo),
		TxInfo:       txInfo,
	}
	err := executor.ApplyTransaction()
	if err != nil {
		return err
	}
	//fromAccount, err := bc.StateDB().GetFormatAccount(txInfo.AccountIndex)
	//if err != nil {
	//	return err
	//}
	//
	//// apply changes
	//fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance = ffmath.Sub(fromAccount.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	//fromAccount.Nonce++
	//fromAccount.CollectionNonce++
	//
	//stateCache := c.bc.StateDB()
	//stateCache.SetPendingAccount(fromAccount.AccountIndex, fromAccount)
	//stateCache.SetPendingGas(txInfo.GasFeeAssetId, txInfo.GasFeeAssetAmount)
	return nil
}
