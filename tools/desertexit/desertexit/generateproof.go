package desertexit

import (
	"encoding/json"
	"fmt"
	types2 "github.com/bnb-chain/zkbnb-crypto/circuit/types"
	"github.com/bnb-chain/zkbnb-crypto/ffmath"
	"github.com/bnb-chain/zkbnb-crypto/wasm/txtypes"
	bsmt "github.com/bnb-chain/zkbnb-smt"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/common/prove"
	"github.com/bnb-chain/zkbnb/core/statedb"
	"github.com/bnb-chain/zkbnb/dao/desertexit"
	"github.com/bnb-chain/zkbnb/dao/l1syncedblock"
	"github.com/bnb-chain/zkbnb/tools/desertexit/config"
	"github.com/bnb-chain/zkbnb/tree"
	"github.com/ethereum/go-ethereum/common"
	"github.com/panjf2000/ants/v2"
	"io/ioutil"
	"os"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/common/chain"
	"github.com/bnb-chain/zkbnb/core"
	"github.com/bnb-chain/zkbnb/types"
)

const DefaultProofFolder = "./tools/desertexit/proofdata/"

type GenerateProof struct {
	running bool
	config  *config.Config
	bc      *core.BlockChain
	pool    *ants.Pool
}

func NewGenerateProof(config *config.Config) (*GenerateProof, error) {
	bc, err := core.NewBlockChainForDesertExit(config)
	if err != nil {
		return nil, fmt.Errorf("new blockchain error: %v", err)
	}

	pool, err := ants.NewPool(50, ants.WithPanicHandler(func(p interface{}) {
		panic("worker exits from a panic")
	}))

	if config.ProofFolder == "" {
		config.ProofFolder = DefaultProofFolder
	}
	desertExit := &GenerateProof{
		running: true,
		config:  config,
		bc:      bc,
		pool:    pool,
	}
	return desertExit, nil
}

func (c *GenerateProof) Run() error {
	err := c.bc.LoadAllAccounts(c.pool)
	if err != nil {
		return err
	}

	err = c.bc.LoadAllNfts(c.pool)
	if err != nil {
		return err
	}

	limit := 1000
	executedBlock, err := c.bc.DesertExitBlockModel.GetLatestExecutedBlock()
	if err != nil && err != types.DbErrNotFound {
		return fmt.Errorf("get executed tx from desert exit block failed:%s", err.Error())
	}

	var executedTxMaxHeight int64 = 0
	if executedBlock != nil {
		executedTxMaxHeight = executedBlock.BlockHeight
	}
	allBlocksHandled := false
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
				return fmt.Errorf("failed to get latest l1 monitor block, err: %v", err)
			}
			if l1SyncedBlock != nil {
				logx.Info("execute all the l2 blocks successfully")
				allBlocksHandled = true
				break
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
				return err
			}

			err = c.saveToDb(pendingBlock)
			if err != nil {
				return err
			}

			executedTxMaxHeight = pendingBlock.BlockHeight
		}
	}

	if allBlocksHandled {
		logx.Info("execute all the l2 blocks successfully")
		account, err := c.bc.AccountModel.GetAccountByL1Address(c.config.Address)
		if err != nil {
			logx.Errorf("get account by address error L1Address=%s,%v,", c.config.Address, err)
			return err
		}

		err = c.generateProof(executedTxMaxHeight, account.AccountIndex, c.config.NftIndexList, c.config.Token)
		if err != nil {
			return err
		}
	}
	return nil
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
	var accountIndexMap map[int64]bool
	var nftIndexMap map[int64]bool
	var addressMap map[string]bool
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

func (c *GenerateProof) generateProof(blockHeight int64, accountIndex int64, nftIndexList []int64, assetTokenAddress string) error {
	accountTree, accountAssetTrees, nftTree, err := c.initSmtTree(blockHeight)

	accountInfo, err := c.bc.DB().AccountModel.GetAccountByIndex(accountIndex)
	if err != nil {
		logx.Errorf("get account failed: %s", err)
		return err
	}
	formatAccountInfo, err := chain.ToFormatAccountInfo(accountInfo)
	if err != nil {
		return err
	}

	// get account before
	accountMerkleProofs, err := accountTree.GetProof(uint64(accountIndex))
	if err != nil {
		return err
	}

	// set account merkle proof
	merkleProofsAccount, err := prove.SetFixedAccountArray(accountMerkleProofs)
	if err != nil {
		return err
	}

	// Marshal formatted proof.
	merkleProofsAccountBytes, err := json.Marshal(merkleProofsAccount)
	if err != nil {
		return err
	}
	logx.Infof("accountIndex=%d, merkleProofsAccount=%s", accountIndex, string(merkleProofsAccountBytes))

	storedBlockInfo, err := c.getStoredBlockInfo()
	if err != nil {
		logx.Errorf("get stored block info: %s", err.Error())
		return err
	}

	pk, err := common2.ParsePubKey(accountInfo.PublicKey)
	if err != nil {
		logx.Errorf("unable to parse pub key: %s", err.Error())
		return err
	}
	accountExitData := DesertVerifierAccountExitData{
		AccountId:       uint32(accountIndex),
		L1Address:       accountInfo.L1Address,
		PubKeyX:         common.Bytes2Hex(pk.A.X.Marshal()),
		PubKeyY:         common.Bytes2Hex(pk.A.Y.Marshal()),
		Nonce:           accountInfo.Nonce,
		CollectionNonce: accountInfo.CollectionNonce,
	}

	accountMerkleProof := make([]string, len(merkleProofsAccount))
	for i, _ := range merkleProofsAccount {
		accountMerkleProof[i] = common.Bytes2Hex(merkleProofsAccount[i])
	}

	if assetTokenAddress != "" {
		monitor, err := NewDesertExit(c.config)
		if err != nil {
			logx.Severe(err)
			return err
		}

		var assetId uint16
		if assetTokenAddress == types.BNBAddress {
			assetId = 0
		} else {
			assetId, err = monitor.ValidateAssetAddress(common.HexToAddress(assetTokenAddress))
			if err != nil {
				logx.Severe(err)
				return err
			}
		}

		assetMerkleProof, err := accountAssetTrees.Get(accountIndex).GetProof(uint64(assetId))
		if err != nil {
			return err
		}
		merkleProofsAccountAsset, err := prove.SetFixedAccountAssetArray(assetMerkleProof)
		if err != nil {
			return err
		}
		merkleProofsAccountAssetBytes, err := json.Marshal(merkleProofsAccountAsset)
		if err != nil {
			return err
		}
		logx.Infof("accountIndex=%d,assetId=%d, merkleProofsAccountAsset=%s", accountIndex, assetId, string(merkleProofsAccountAssetBytes))

		performDesertData := PerformDesertAssetData{}
		performDesertData.AccountMerkleProof = accountMerkleProof

		assetMerkleProofByte := make([]string, len(merkleProofsAccountAsset))
		for i, _ := range merkleProofsAccountAsset {
			assetMerkleProofByte[i] = common.Bytes2Hex(merkleProofsAccountAsset[i])
		}

		performDesertData.AssetMerkleProof = assetMerkleProofByte
		performDesertData.NftRoot = common.Bytes2Hex(nftTree.Root())

		performDesertData.AccountExitData = accountExitData
		performDesertData.AssetExitData = DesertVerifierAssetExitData{
			AssetId:                  assetId,
			Amount:                   formatAccountInfo.AssetInfo[int64(assetId)].Balance.String(),
			OfferCanceledOrFinalized: formatAccountInfo.AssetInfo[int64(assetId)].OfferCanceledOrFinalized.String(),
		}
		performDesertData.StoredBlockInfo = storedBlockInfo

		data, err := json.Marshal(performDesertData)
		if err != nil {
			return err
		}
		Mkdir(c.config.ProofFolder)
		err = ioutil.WriteFile(c.config.ProofFolder+"performDesertAsset.json", data, 0777)
		if err != nil {
			return err
		}
	}

	if len(nftIndexList) > 0 {
		performDesertNftData := &PerformDesertNftData{}
		var exitNfts []DesertVerifierNftExitData
		var nftMerkleProofsList [][]string
		for _, nftIndex := range nftIndexList {
			nftInfo, err := c.bc.DB().L2NftModel.GetNft(nftIndex)
			if err != nil {
				logx.Errorf("get nft failed: %s", err)
				return err
			}

			nftMerkleProofs, err := nftTree.GetProof(uint64(nftIndex))
			if err != nil {
				return err
			}
			merkleProofsNft, err := prove.SetFixedNftArray(nftMerkleProofs)
			if err != nil {
				return err
			}
			merkleProofsNftBytes, err := json.Marshal(merkleProofsNft)
			if err != nil {
				return err
			}
			logx.Infof("accountIndex=%d,nftIndex=%d, merkleProofsNft=%s", accountIndex, nftIndex, string(merkleProofsNftBytes))

			exitNftData := DesertVerifierNftExitData{}
			contentHash := common.Hex2Bytes(nftInfo.NftContentHash)
			if len(contentHash) >= types2.NftContentHashBytesSize {
				exitNftData.NftContentHash1 = common.Bytes2Hex(contentHash[:types2.NftContentHashBytesSize])
				exitNftData.NftContentHash2 = common.Bytes2Hex(contentHash[types2.NftContentHashBytesSize:])
			} else {
				exitNftData.NftContentHash1 = common.Bytes2Hex(contentHash[:])
			}
			exitNftData.NftIndex = uint64(nftIndex)
			exitNftData.CollectionId = nftInfo.CollectionId
			exitNftData.CreatorAccountIndex = nftInfo.CreatorAccountIndex
			exitNftData.RoyaltyRate = nftInfo.RoyaltyRate
			exitNftData.NftContentType = uint8(nftInfo.NftContentType)
			exitNftData.OwnerAccountIndex = nftInfo.OwnerAccountIndex

			exitNfts = append(exitNfts, exitNftData)
			merkleProofsNftByte := make([]string, len(merkleProofsNft))
			for i, _ := range merkleProofsNft {
				merkleProofsNftByte[i] = common.Bytes2Hex(merkleProofsNft[i])
			}
			nftMerkleProofsList = append(nftMerkleProofsList, merkleProofsNftByte)
		}

		performDesertNftData.ExitNfts = exitNfts
		performDesertNftData.NftMerkleProofs = nftMerkleProofsList
		performDesertNftData.AssetRoot = common.Bytes2Hex(accountAssetTrees.Get(accountIndex).Root())
		performDesertNftData.StoredBlockInfo = storedBlockInfo
		performDesertNftData.AccountExitData = accountExitData
		performDesertNftData.AccountMerkleProof = accountMerkleProof

		data, err := json.Marshal(performDesertNftData)
		Mkdir(c.config.ProofFolder)
		err = ioutil.WriteFile(c.config.ProofFolder+"performDesertNft.json", data, 0777)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *GenerateProof) initSmtTree(blockHeight int64) (accountTree bsmt.SparseMerkleTree, accountAssetTrees *tree.AssetTreeCache, nftTree bsmt.SparseMerkleTree, err error) {
	treeCtx, err := tree.NewContext("desertexit", c.config.TreeDB.Driver, true, true, c.config.TreeDB.RoutinePoolSize, &c.config.TreeDB.LevelDBOption, &c.config.TreeDB.RedisDBOption)
	if err != nil {
		logx.Errorf("init tree database failed: %s", err)
		return nil, nil, nil, err
	}

	treeCtx.SetOptions(bsmt.InitializeVersion(0))
	treeCtx.SetBatchReloadSize(1000)
	err = tree.SetupTreeDB(treeCtx)
	if err != nil {
		logx.Errorf("init tree database failed: %s", err)
		return nil, nil, nil, err
	}

	// dbinitializer accountTree and accountStateTrees
	accountTree, accountAssetTrees, err = tree.InitAccountTree(
		c.bc.AccountModel,
		c.bc.AccountHistoryModel,
		make([]int64, 0),
		blockHeight,
		treeCtx,
		c.config.TreeDB.AssetTreeCacheSize,
		false,
	)
	if err != nil {
		logx.Error("init merkle tree error:", err)
		return nil, nil, nil, err
	}
	accountStateRoot := common.Bytes2Hex(accountTree.Root())
	logx.Infof("account tree accountStateRoot=%s", accountStateRoot)

	// dbinitializer nftTree
	nftTree, err = tree.InitNftTree(
		c.bc.L2NftModel,
		c.bc.L2NftHistoryModel,
		blockHeight,
		treeCtx, false)
	if err != nil {
		logx.Errorf("init nft tree error: %s", err.Error())
		return nil, nil, nil, err
	}
	nftStateRoot := common.Bytes2Hex(nftTree.Root())
	logx.Infof("nft tree nftStateRoot=%s", nftStateRoot)

	stateRoot := tree.ComputeStateRootHash(accountTree.Root(), nftTree.Root())
	logx.Infof("smt tree StateRoot=%s", common.Bytes2Hex(stateRoot))

	return accountTree, accountAssetTrees, nftTree, nil
}

func (c *GenerateProof) getStoredBlockInfo() (*StoredBlockInfo, error) {
	m, err := NewDesertExit(c.config)
	if err != nil {
		return nil, err
	}
	desertExitBlock, err := c.bc.DB().DesertExitBlockModel.GetLatestExecutedBlock()
	if err != nil {
		logx.Errorf("get desert exit block failed: %s", err)
		return nil, err
	}

	lastStoredBlockInfo, err := m.getLastStoredBlockInfo(desertExitBlock.VerifiedTxHash, desertExitBlock.BlockHeight)
	if err != nil {
		logx.Errorf("get last stored block info failed: %s", err)
		return nil, err
	}

	storedBlockInfo := &StoredBlockInfo{
		BlockSize:                    lastStoredBlockInfo.BlockSize,
		BlockNumber:                  lastStoredBlockInfo.BlockNumber,
		PriorityOperations:           lastStoredBlockInfo.PriorityOperations,
		PendingOnchainOperationsHash: common.Bytes2Hex(lastStoredBlockInfo.PendingOnchainOperationsHash[:]),
		Timestamp:                    lastStoredBlockInfo.Timestamp.Int64(),
		StateRoot:                    common.Bytes2Hex(lastStoredBlockInfo.StateRoot[:]),
		Commitment:                   common.Bytes2Hex(lastStoredBlockInfo.Commitment[:]),
	}
	return storedBlockInfo, nil
}

func Mkdir(dir string) {
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		logx.Errorf("make dir error,%s", err)
	}
}

func (c *GenerateProof) Shutdown() {
	c.running = false
	c.bc.Statedb.Close()
	c.bc.ChainDB.Close()
}
