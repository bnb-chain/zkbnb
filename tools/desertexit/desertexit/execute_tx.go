package desertexit

import (
	"fmt"
	"github.com/bnb-chain/zkbnb-crypto/ffmath"
	"github.com/bnb-chain/zkbnb-crypto/wasm/txtypes"
	"github.com/bnb-chain/zkbnb/core/statedb"
	"github.com/bnb-chain/zkbnb/dao/desertexit"
	"github.com/bnb-chain/zkbnb/dao/l1syncedblock"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/common/chain"
	"github.com/bnb-chain/zkbnb/core"
	"github.com/bnb-chain/zkbnb/types"
)

func (c *GenerateProof) Run() (int64, error) {
	limit := 1000
	executedBlock, err := c.bc.DesertExitBlockModel.GetLatestExecutedBlock()
	if err != nil && err != types.DbErrNotFound {
		return 0, fmt.Errorf("get executed tx from desert exit block failed:%s", err.Error())
	}

	var executedTxMaxHeight int64 = 0
	if executedBlock != nil {
		executedTxMaxHeight = executedBlock.BlockHeight
	}

	isDone, err := c.isDone(executedTxMaxHeight)
	if err != nil {
		return 0, err
	}
	if isDone {
		return executedTxMaxHeight, nil
	}

	err = c.bc.LoadAllAccounts(c.pool)
	if err != nil {
		return 0, err
	}

	err = c.bc.LoadAllNfts(c.pool)
	if err != nil {
		return 0, err
	}

	for {
		if !c.running {
			break
		}

		pendingBlocks, err := c.bc.DesertExitBlockModel.GetBlocksByStatusAndMaxHeight(desertexit.StatusVerified, executedTxMaxHeight, int64(limit))
		if err != nil && err != types.DbErrNotFound {
			logx.Errorf("get pending blocks from desert exit block failed:%s", err.Error())
			time.Sleep(100 * time.Millisecond)
			continue
		}
		if err == types.DbErrNotFound {
			l1SyncedBlock, err := c.bc.L1SyncedBlockModel.GetLatestL1SyncedBlockByType(l1syncedblock.TypeDesert)
			if err != nil && err != types.DbErrNotFound {
				return 0, fmt.Errorf("failed to get latest l1 monitor block, err: %v", err)
			}
			if l1SyncedBlock != nil {
				logx.Info("execute all the l2 blocks successfully")
				return executedTxMaxHeight, nil
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}

		for _, pendingBlock := range pendingBlocks {
			if int(pendingBlock.BlockHeight)-int(executedTxMaxHeight) != 1 {
				time.Sleep(50 * time.Millisecond)
				logx.Infof("not equal block height=%d", pendingBlock.BlockHeight)
				break
			}

			err := c.executeBlockFunc(pendingBlock)
			if err != nil {
				return 0, err
			}

			err = c.saveToDb(pendingBlock)
			if err != nil {
				return 0, err
			}

			executedTxMaxHeight = pendingBlock.BlockHeight
		}
	}
	return executedTxMaxHeight, nil
}

func (c *GenerateProof) isDone(executedTxMaxHeight int64) (bool, error) {
	_, err := c.bc.DesertExitBlockModel.GetBlocksByStatusAndMaxHeight(desertexit.StatusVerified, executedTxMaxHeight, 1)
	if err != nil && err != types.DbErrNotFound {
		return false, fmt.Errorf("get pending blocks from desert exit block failed:%s", err.Error())
	}
	if err == types.DbErrNotFound {
		l1SyncedBlock, err := c.bc.L1SyncedBlockModel.GetLatestL1SyncedBlockByType(l1syncedblock.TypeDesert)
		if err != nil && err != types.DbErrNotFound {
			return false, fmt.Errorf("failed to get latest l1 monitor block, err: %v", err)
		}
		if l1SyncedBlock != nil {
			logx.Info("execute all the l2 blocks successfully")
			return true, nil
		}
	}
	return false, nil
}

func (c *GenerateProof) executeBlockFunc(desertExitBlock *desertexit.DesertExitBlock) error {
	c.bc.Statedb.PurgeCache("")
	err := c.bc.Statedb.MarkGasAccountAsPending()
	if err != nil {
		return err
	}

	txInfos, err := chain.ParsePubDataForDesert(desertExitBlock.PubData)
	if err != nil {
		return err
	}

	c.preLoadAccountAndNft(txInfos)

	for _, txInfo := range txInfos {
		err := core.NewDesertProcessor(c.bc).Process(txInfo)
		if err != nil {
			return err
		}
	}

	gasAccount := c.bc.Statedb.StateCache.PendingAccountMap[types.GasAccount]
	if gasAccount != nil {
		if len(c.bc.Statedb.StateCache.PendingGasMap) != 0 {
			for assetId, delta := range c.bc.Statedb.StateCache.PendingGasMap {
				if asset, ok := gasAccount.AssetInfo[assetId]; ok {
					gasAccount.AssetInfo[assetId].Balance = ffmath.Add(asset.Balance, delta)
				} else {
					gasAccount.AssetInfo[assetId] = &types.AccountAsset{
						Balance:                  delta,
						OfferCanceledOrFinalized: types.ZeroBigInt,
					}
				}
				c.bc.Statedb.MarkAccountAssetsDirty(gasAccount.AccountIndex, []int64{assetId})
			}
		} else {
			assetsMap := c.bc.Statedb.GetDirtyAccountsAndAssetsMap()[gasAccount.AccountIndex]
			if assetsMap == nil {
				delete(c.bc.Statedb.StateCache.PendingAccountMap, types.GasAccount)
			}
		}
	}

	for _, formatAccount := range c.bc.Statedb.StateCache.PendingAccountMap {
		assetsMap := c.bc.Statedb.GetDirtyAccountsAndAssetsMap()[formatAccount.AccountIndex]
		if assetsMap == nil {
			return fmt.Errorf("%d exists in PendingAccountMap but not in GetDirtyAccountsAndAssetsMap", formatAccount.AccountIndex)
		}
	}

	for accountIndex, _ := range c.bc.Statedb.GetDirtyAccountsAndAssetsMap() {
		_, exist := c.bc.Statedb.StateCache.GetPendingAccount(accountIndex)
		if !exist {
			accountInfo, err := c.bc.Statedb.GetFormatAccount(accountIndex)
			if err != nil {
				return fmt.Errorf("get account info failed,accountIndex=%d,err=%s ", accountIndex, err.Error())
			}
			c.bc.Statedb.SetPendingAccount(accountIndex, accountInfo)
		}
	}

	for _, nftInfo := range c.bc.Statedb.StateCache.PendingNftMap {
		if c.bc.Statedb.GetDirtyNftMap()[nftInfo.NftIndex] == false {
			return fmt.Errorf("%d exists in PendingNftMap but not in DirtyNftMap", nftInfo.NftIndex)
		}
	}

	for nftIndex, _ := range c.bc.Statedb.StateCache.GetDirtyNftMap() {
		_, exist := c.bc.Statedb.StateCache.GetPendingNft(nftIndex)
		if !exist {
			nftInfo, err := c.bc.Statedb.GetNft(nftIndex)
			if err != nil {
				return fmt.Errorf("get nft info failed,nftIndex=%d,err=%s ", nftIndex, err.Error())
			}
			c.bc.Statedb.SetPendingNft(nftIndex, nftInfo)
		}
	}

	return nil
}

func (c *GenerateProof) preLoadAccountAndNft(txInfos []txtypes.TxInfo) {
	accountIndexMap := make(map[int64]bool, 0)
	nftIndexMap := make(map[int64]bool, 0)
	addressMap := make(map[string]bool, 0)
	for _, txInfo := range txInfos {
		core.NewDesertProcessor(c.bc).PreProcess(txInfo, accountIndexMap, nftIndexMap, addressMap)
	}
	c.bc.Statedb.PreLoadAccountAndNft(accountIndexMap, nftIndexMap, addressMap)
}

func (c *GenerateProof) saveToDb(desertExitBlock *desertexit.DesertExitBlock) error {
	logx.Infof("saveToDb start, blockHeight:%d", desertExitBlock.BlockHeight)
	stateDataCopy := &statedb.StateDataCopy{
		StateCache:   c.bc.Statedb.StateCache,
		CurrentBlock: nil,
	}
	pendingAccounts, _, err := c.bc.Statedb.GetPendingAccount(desertExitBlock.BlockHeight, stateDataCopy)
	if err != nil {
		return err
	}

	pendingNfts, _, err := c.bc.Statedb.GetPendingNft(desertExitBlock.BlockHeight, stateDataCopy)
	if err != nil {
		return err
	}
	// update db
	err = c.bc.DB().DB.Transaction(func(tx *gorm.DB) error {
		err := c.bc.DB().DesertExitBlockModel.UpdateBlockToExecutedInTransact(tx, desertExitBlock)
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
		logx.Errorf("saveToDb failed:%s,blockHeight:%d", err.Error(), desertExitBlock.BlockHeight)
		return err
	}

	for _, accountInfo := range pendingAccounts {
		c.bc.Statedb.PendingAccountMap[accountInfo.AccountIndex].AccountId = int64(accountInfo.ID)
	}

	for _, nftInfo := range pendingNfts {
		c.bc.Statedb.PendingNftMap[nftInfo.NftIndex].ID = nftInfo.ID
	}

	c.bc.Statedb.SyncPendingAccountToMemoryCache(c.bc.Statedb.PendingAccountMap)
	c.bc.Statedb.SyncPendingNftToMemoryCache(c.bc.Statedb.PendingNftMap)
	return nil
}
