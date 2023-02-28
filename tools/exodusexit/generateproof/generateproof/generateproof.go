package generateproof

import (
	"encoding/json"
	"fmt"
	types2 "github.com/bnb-chain/zkbnb-crypto/circuit/types"
	"github.com/bnb-chain/zkbnb-crypto/util"
	"github.com/bnb-chain/zkbnb-crypto/wasm/txtypes"
	bsmt "github.com/bnb-chain/zkbnb-smt"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/common/prove"
	"github.com/bnb-chain/zkbnb/core/executor"
	"github.com/bnb-chain/zkbnb/core/statedb"
	"github.com/bnb-chain/zkbnb/dao/exodusexit"
	"github.com/bnb-chain/zkbnb/tools/exodusexit/generateproof/config"
	"github.com/bnb-chain/zkbnb/tree"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/panjf2000/ants/v2"
	"math/big"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbnb/common/chain"
	"github.com/bnb-chain/zkbnb/core"
	"github.com/bnb-chain/zkbnb/types"
)

type ExodusExit struct {
	running bool
	config  *config.Config
	bc      *core.BlockChain
	pool    *ants.Pool
}

func NewExodusExit(config *config.Config) (*ExodusExit, error) {
	bc, err := core.NewBlockChainForExodusExit(config)
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
		if c.config.ChainConfig.EndL2BlockHeight == executedTxMaxHeight {
			logx.Info("execute all the l2 blocks successfully")
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
			err = c.saveToDb(pendingBlock)
			if err != nil {
				return err
			}
			executedTxMaxHeight = pendingBlock.BlockHeight
		}
	}
	if c.config.ChainConfig.EndL2BlockHeight == executedTxMaxHeight {
		logx.Info("execute all the l2 blocks successfully")
		account, err := c.bc.AccountModel.GetAccountByName(c.config.AccountName)
		if err != nil {
			logx.Errorf("get account by name error accountName=%s,%v,", c.config.AccountName, err)
			return err
		}
		c.getMerkleProofs(executedTxMaxHeight, account.AccountIndex, c.config.NftIndex, c.config.AssetId)
	}
	return nil
}

func (c *ExodusExit) executeBlockFunc(exodusExitBlock *exodusexit.ExodusExitBlock) error {
	pubData := common.FromHex(exodusExitBlock.PubData)
	sizePerTx := types2.PubDataBitsSizePerTx / 8
	c.bc.Statedb.PurgeCache("")
	for i := 0; i < int(exodusExitBlock.BlockSize); i++ {
		subPubData := pubData[i*sizePerTx : (i+1)*sizePerTx]
		offset := 0
		offset, txType := common2.ReadUint8(subPubData, offset)
		switch txType {
		case types.TxTypeAtomicMatch:
			err := c.executeAtomicMatch(subPubData)
			if err != nil {
				return err
			}
			break
		case types.TxTypeCancelOffer:
			err := c.executeCancelOffer(subPubData)
			if err != nil {
				return err
			}
			break
		case types.TxTypeCreateCollection:
			err := c.executeCollection(subPubData)
			if err != nil {
				return err
			}
			break
		case types.TxTypeDeposit:
			err := c.executeDeposit(subPubData)
			if err != nil {
				return err
			}
			break
		case types.TxTypeDepositNft:
			err := c.executeDepositNft(subPubData)
			if err != nil {
				return err
			}
			break
		case types.TxTypeFullExit:
			err := c.executeFullExit(subPubData)
			if err != nil {
				return err
			}
			break
		case types.TxTypeFullExitNft:
			err := c.executeFullExitNft(subPubData)
			if err != nil {
				return err
			}
			break
		case types.TxTypeMintNft:
			err := c.executeMintNft(subPubData)
			if err != nil {
				return err
			}
			break
		case types.TxTypeRegisterZns:
			err := c.executeRegisterZns(subPubData)
			if err != nil {
				return err
			}
			break
		case types.TxTypeTransfer:
			err := c.executeTransfer(subPubData)
			if err != nil {
				return err
			}
			break
		case types.TxTypeTransferNft:
			err := c.executeTransferNft(subPubData)
			if err != nil {
				return err
			}
			break
		case types.TxTypeWithdraw:
			err := c.executeWithdraw(subPubData)
			if err != nil {
				return err
			}
			break
		case types.TxTypeWithdrawNft:
			err := c.executeWithdrawNft(subPubData)
			if err != nil {
				return err
			}
			break
		case types.TxTypeEmpty:
			return nil
		}
	}
	return nil

}

func (c *ExodusExit) saveToDb(exodusExitBlock *exodusexit.ExodusExitBlock) error {
	logx.Infof("saveToDb start, blockHeight:%d", exodusExitBlock.BlockHeight)
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
		logx.Errorf("saveToDb failed:%s,blockHeight:%d", err.Error(), exodusExitBlock.BlockHeight)
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

func (c *ExodusExit) executeAtomicMatch(pubData []byte) error {
	bc := c.bc
	offset := 1
	offset, accountIndex := common2.ReadUint32(pubData, offset)
	offset, buyOfferAccountIndex := common2.ReadUint32(pubData, offset)
	offset, buyOfferOfferId := common2.ReadUint24(pubData, offset)
	offset, sellOfferAccountIndex := common2.ReadUint32(pubData, offset)
	offset, sellOfferOfferId := common2.ReadUint24(pubData, offset)
	offset, buyOfferNftIndex := common2.ReadUint40(pubData, offset)
	offset, sellOfferAssetId := common2.ReadUint16(pubData, offset)
	offset, buyOfferAssetPackedAmount := common2.ReadUint40(pubData, offset)
	buyOfferAssetAmount, err := util.CleanPackedAmount(big.NewInt(buyOfferAssetPackedAmount))
	if err != nil {
		logx.Errorf("unable to convert amount to packed amount: %s", err.Error())
		return err
	}

	offset, creatorPackedAmount := common2.ReadUint40(pubData, offset)
	creatorAmount, err := util.CleanPackedAmount(big.NewInt(creatorPackedAmount))
	if err != nil {
		logx.Errorf("unable to convert amount to packed amount: %s", err.Error())
		return err
	}

	offset, treasuryPackedAmount := common2.ReadUint40(pubData, offset)
	treasuryAmount, err := util.CleanPackedAmount(big.NewInt(treasuryPackedAmount))
	if err != nil {
		logx.Errorf("unable to convert amount to packed amount: %s", err.Error())
		return err
	}

	offset, gasFeeAssetId := common2.ReadUint16(pubData, offset)

	offset, gasFeeAssetPackedAmount := common2.ReadUint16(pubData, offset)
	gasFeeAssetAmount, err := util.CleanPackedFee(big.NewInt(int64(gasFeeAssetPackedAmount)))
	if err != nil {
		return err
	}

	txInfo := &txtypes.AtomicMatchTxInfo{
		AccountIndex: int64(accountIndex),
		BuyOffer: &txtypes.OfferTxInfo{
			AccountIndex: int64(buyOfferAccountIndex),
			OfferId:      int64(buyOfferOfferId),
			NftIndex:     buyOfferNftIndex,
			AssetAmount:  buyOfferAssetAmount,
		},
		SellOffer: &txtypes.OfferTxInfo{
			AccountIndex: int64(sellOfferAccountIndex),
			OfferId:      int64(sellOfferOfferId),
			AssetId:      int64(sellOfferAssetId),
		},
		CreatorAmount:     creatorAmount,
		TreasuryAmount:    treasuryAmount,
		GasFeeAssetAmount: gasFeeAssetAmount,
		GasFeeAssetId:     int64(gasFeeAssetId),
	}
	executor := &executor.AtomicMatchExecutor{
		BaseExecutor: executor.NewBaseExecutor(bc, nil, txInfo),
		TxInfo:       txInfo,
	}
	err = executor.Prepare()
	if err != nil {
		return err
	}
	err = executor.ApplyTransaction()
	if err != nil {
		return err
	}
	return nil
}

func (c *ExodusExit) executeCancelOffer(pubData []byte) error {
	bc := c.bc
	offset := 1
	offset, accountIndex := common2.ReadUint32(pubData, offset)
	offset, offerId := common2.ReadUint24(pubData, offset)
	offset, gasFeeAssetId := common2.ReadUint16(pubData, offset)
	offset, gasFeeAssetPackedAmount := common2.ReadUint16(pubData, offset)
	gasFeeAssetAmount, _ := util.CleanPackedFee(big.NewInt(int64(gasFeeAssetPackedAmount)))

	txInfo := &txtypes.CancelOfferTxInfo{
		AccountIndex:      int64(accountIndex),
		GasFeeAssetId:     int64(gasFeeAssetId),
		GasFeeAssetAmount: gasFeeAssetAmount,
		OfferId:           int64(offerId),
	}

	executor := &executor.CancelOfferExecutor{
		BaseExecutor: executor.NewBaseExecutor(bc, nil, txInfo),
		TxInfo:       txInfo,
	}
	err := executor.Prepare()
	if err != nil {
		return err
	}
	err = executor.ApplyTransaction()
	if err != nil {
		return err
	}
	return nil
}

func (c *ExodusExit) executeCollection(pubData []byte) error {
	bc := c.bc
	offset := 1
	offset, accountIndex := common2.ReadUint32(pubData, offset)
	offset, collectionId := common2.ReadUint16(pubData, offset)
	offset, gasFeeAssetId := common2.ReadUint16(pubData, offset)
	offset, gasFeeAssetPackedAmount := common2.ReadUint16(pubData, offset)
	gasFeeAssetAmount, _ := util.CleanPackedFee(big.NewInt(int64(gasFeeAssetPackedAmount)))

	txInfo := &txtypes.CreateCollectionTxInfo{
		AccountIndex:      int64(accountIndex),
		GasFeeAssetId:     int64(gasFeeAssetId),
		GasFeeAssetAmount: gasFeeAssetAmount,
		CollectionId:      int64(collectionId),
	}

	executor := &executor.CreateCollectionExecutor{
		BaseExecutor: executor.NewBaseExecutor(bc, nil, txInfo),
		TxInfo:       txInfo,
	}
	err := executor.Prepare()
	if err != nil {
		return err
	}
	err = executor.ApplyTransaction()
	if err != nil {
		return err
	}
	return nil
}

func (c *ExodusExit) executeDeposit(pubData []byte) error {
	bc := c.bc
	offset := 1
	offset, accountIndex := common2.ReadUint32(pubData, offset)
	offset, assetId := common2.ReadUint16(pubData, offset)
	offset, assetPackedAmount := common2.ReadUint128(pubData, offset)
	assetAmount, _ := util.CleanPackedFee(assetPackedAmount)
	offset, accountNameHash := common2.ReadPrefixPaddingBufToChunkSize(pubData, offset)

	txInfo := &txtypes.DepositTxInfo{
		AccountIndex:    int64(accountIndex),
		AssetId:         int64(assetId),
		AssetAmount:     assetAmount,
		AccountNameHash: accountNameHash,
	}

	executor := &executor.DepositExecutor{
		BaseExecutor: executor.NewBaseExecutor(bc, nil, txInfo),
		TxInfo:       txInfo,
	}
	err := executor.Prepare()
	if err != nil {
		return err
	}
	err = executor.ApplyTransaction()
	if err != nil {
		return err
	}
	return nil
}

func (c *ExodusExit) executeDepositNft(pubData []byte) error {
	bc := c.bc
	offset := 1
	offset, accountIndex := common2.ReadUint32(pubData, offset)
	offset, nftIndex := common2.ReadUint40(pubData, offset)
	offset, creatorAccountIndex := common2.ReadUint32(pubData, offset)
	offset, creatorTreasuryRate := common2.ReadUint16(pubData, offset)
	offset, collectionId := common2.ReadUint16(pubData, offset)
	offset, nftContentHash := common2.ReadPrefixPaddingBufToChunkSize(pubData, offset)
	offset, accountNameHash := common2.ReadPrefixPaddingBufToChunkSize(pubData, offset)
	txInfo := &txtypes.DepositNftTxInfo{
		AccountIndex:        int64(accountIndex),
		NftIndex:            nftIndex,
		CreatorAccountIndex: int64(creatorAccountIndex),
		CollectionId:        int64(collectionId),
		CreatorTreasuryRate: int64(creatorTreasuryRate),
		NftContentHash:      nftContentHash,
		AccountNameHash:     accountNameHash,
	}
	executor := &executor.DepositNftExecutor{
		BaseExecutor: executor.NewBaseExecutor(bc, nil, txInfo),
		TxInfo:       txInfo,
	}
	err := executor.Prepare()
	if err != nil {
		return err
	}
	err = executor.ApplyTransaction()
	if err != nil {
		return err
	}
	return nil
}

func (c *ExodusExit) executeFullExit(pubData []byte) error {
	bc := c.bc
	offset := 1
	offset, accountIndex := common2.ReadUint32(pubData, offset)
	offset, assetId := common2.ReadUint16(pubData, offset)
	offset, assetAmount := common2.ReadUint128(pubData, offset)
	offset, accountNameHash := common2.ReadPrefixPaddingBufToChunkSize(pubData, offset)
	var txInfo = &txtypes.FullExitTxInfo{
		AccountIndex:    int64(accountIndex),
		AssetId:         int64(assetId),
		AssetAmount:     assetAmount,
		AccountNameHash: accountNameHash,
	}
	executor := &executor.FullExitExecutor{
		BaseExecutor: executor.NewBaseExecutor(bc, nil, txInfo),
		TxInfo:       txInfo,
	}
	err := executor.Prepare()
	if err != nil {
		return err
	}
	err = executor.ApplyTransaction()
	if err != nil {
		return err
	}
	return nil
}

func (c *ExodusExit) executeFullExitNft(pubData []byte) error {
	bc := c.bc
	offset := 1
	offset, accountIndex := common2.ReadUint32(pubData, offset)
	offset, creatorAccountIndex := common2.ReadUint32(pubData, offset)
	offset, creatorTreasuryRate := common2.ReadUint16(pubData, offset)
	offset, nftIndex := common2.ReadUint40(pubData, offset)
	offset, collectionId := common2.ReadUint16(pubData, offset)
	offset, accountNameHash := common2.ReadPrefixPaddingBufToChunkSize(pubData, offset)
	offset, creatorAccountNameHash := common2.ReadPrefixPaddingBufToChunkSize(pubData, offset)
	offset, nftContentHash := common2.ReadPrefixPaddingBufToChunkSize(pubData, offset)

	var txInfo = &txtypes.FullExitNftTxInfo{
		AccountIndex:           int64(accountIndex),
		CreatorAccountIndex:    int64(creatorAccountIndex),
		CreatorTreasuryRate:    int64(creatorTreasuryRate),
		NftIndex:               nftIndex,
		CollectionId:           int64(collectionId),
		AccountNameHash:        accountNameHash,
		CreatorAccountNameHash: creatorAccountNameHash,
		NftContentHash:         nftContentHash,
	}
	executor := &executor.FullExitNftExecutor{
		BaseExecutor: executor.NewBaseExecutor(bc, nil, txInfo),
		TxInfo:       txInfo,
	}
	err := executor.Prepare()
	if err != nil {
		return err
	}
	err = executor.ApplyTransaction()
	if err != nil {
		return err
	}
	return nil
}

func (c *ExodusExit) executeMintNft(pubData []byte) error {
	bc := c.bc
	offset := 1
	offset, creatorAccountIndex := common2.ReadUint32(pubData, offset)
	offset, toAccountIndex := common2.ReadUint32(pubData, offset)
	offset, nftIndex := common2.ReadUint40(pubData, offset)
	offset, gasFeeAssetId := common2.ReadUint16(pubData, offset)

	offset, gasFeeAssetPackedAmount := common2.ReadUint16(pubData, offset)
	gasFeeAssetAmount, err := util.CleanPackedFee(big.NewInt(int64(gasFeeAssetPackedAmount)))
	if err != nil {
		return err
	}
	offset, creatorTreasuryRate := common2.ReadUint16(pubData, offset)
	offset, collectionId := common2.ReadUint16(pubData, offset)
	offset, nftContentHash := common2.ReadPrefixPaddingBufToChunkSize(pubData, offset)

	var txInfo = &txtypes.MintNftTxInfo{
		CreatorAccountIndex: int64(creatorAccountIndex),
		ToAccountIndex:      int64(toAccountIndex),
		NftIndex:            nftIndex,
		GasFeeAssetId:       int64(gasFeeAssetId),
		GasFeeAssetAmount:   gasFeeAssetAmount,
		NftCollectionId:     int64(collectionId),
		NftContentHash:      string(nftContentHash),
		CreatorTreasuryRate: int64(creatorTreasuryRate),
	}
	executor := &executor.MintNftExecutor{
		BaseExecutor: executor.NewBaseExecutor(bc, nil, txInfo),
		TxInfo:       txInfo,
	}
	err = executor.Prepare()
	if err != nil {
		return err
	}
	err = executor.ApplyTransaction()
	if err != nil {
		return err
	}
	return nil
}

func (c *ExodusExit) executeRegisterZns(pubData []byte) error {
	bc := c.bc
	offset := 1
	offset, accountIndex := common2.ReadUint32(pubData, offset)
	offset, accountName := common2.ReadAccountNameFromBytes20(pubData, offset)
	offset, accountNameHash := common2.ReadPrefixPaddingBufToChunkSize(pubData, offset)
	offset, pubKeyX := common2.ReadBytes32(pubData, offset)
	_, pubKeyY := common2.ReadBytes32(pubData, offset)
	pk := new(eddsa.PublicKey)
	pk.A.X.SetBytes(pubKeyX)
	pk.A.Y.SetBytes(pubKeyY)

	var txInfo = &txtypes.RegisterZnsTxInfo{
		AccountIndex:    int64(accountIndex),
		AccountName:     common2.CleanAccountName(common2.SerializeAccountName(accountName)),
		AccountNameHash: accountNameHash,
		PubKey:          common.Bytes2Hex(pk.Bytes()),
	}
	executor := &executor.RegisterZnsExecutor{
		BaseExecutor: executor.NewBaseExecutor(bc, nil, txInfo),
		TxInfo:       txInfo,
	}
	err := executor.Prepare()
	if err != nil {
		return err
	}
	err = executor.ApplyTransaction()
	if err != nil {
		return err
	}
	return nil

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
	err = executor.Prepare()
	if err != nil {
		return err
	}
	err = executor.ApplyTransaction()
	if err != nil {
		return err
	}
	return nil
}

func (c *ExodusExit) executeTransferNft(pubData []byte) error {
	bc := c.bc
	offset := 1
	offset, fromAccountIndex := common2.ReadUint32(pubData, offset)
	offset, toAccountIndex := common2.ReadUint32(pubData, offset)
	offset, nftIndex := common2.ReadUint40(pubData, offset)
	offset, gasFeeAssetId := common2.ReadUint16(pubData, offset)
	offset, packedFee := common2.ReadUint16(pubData, offset)
	gasFeeAssetAmount, err := util.CleanPackedFee(big.NewInt(int64(packedFee)))
	if err != nil {
		return err
	}
	offset, callDataHash := common2.ReadPrefixPaddingBufToChunkSize(pubData, offset)
	txInfo := &txtypes.TransferNftTxInfo{
		FromAccountIndex:  int64(fromAccountIndex),
		ToAccountIndex:    int64(toAccountIndex),
		NftIndex:          int64(nftIndex),
		GasFeeAssetId:     int64(gasFeeAssetId),
		GasFeeAssetAmount: gasFeeAssetAmount,
		CallDataHash:      callDataHash,
	}
	executor := &executor.TransferNftExecutor{
		BaseExecutor: executor.NewBaseExecutor(bc, nil, txInfo),
		TxInfo:       txInfo,
	}
	err = executor.Prepare()
	if err != nil {
		return err
	}
	err = executor.ApplyTransaction()
	if err != nil {
		return err
	}
	return nil
}

func (c *ExodusExit) executeWithdraw(pubData []byte) error {
	bc := c.bc
	offset := 1
	offset, fromAccountIndex := common2.ReadUint32(pubData, offset)
	offset, toAddress := common2.ReadAddress(pubData, offset)
	offset, assetId := common2.ReadUint16(pubData, offset)
	offset, assetAmount := common2.ReadUint128(pubData, offset)
	offset, gasFeeAssetId := common2.ReadUint16(pubData, offset)
	offset, gasFeeAssetPackedAmount := common2.ReadUint16(pubData, offset)
	gasFeeAssetAmount, err := util.CleanPackedFee(big.NewInt(int64(gasFeeAssetPackedAmount)))
	if err != nil {
		return err
	}
	txInfo := &txtypes.WithdrawTxInfo{
		FromAccountIndex:  int64(fromAccountIndex),
		ToAddress:         toAddress,
		AssetId:           int64(assetId),
		AssetAmount:       assetAmount,
		GasFeeAssetId:     int64(gasFeeAssetId),
		GasFeeAssetAmount: gasFeeAssetAmount,
	}
	executor := &executor.WithdrawExecutor{
		BaseExecutor: executor.NewBaseExecutor(bc, nil, txInfo),
		TxInfo:       txInfo,
	}
	err = executor.Prepare()
	if err != nil {
		return err
	}
	err = executor.ApplyTransaction()
	if err != nil {
		return err
	}
	return nil
}

func (c *ExodusExit) executeWithdrawNft(pubData []byte) error {
	bc := c.bc
	offset := 1
	offset, fromAccountIndex := common2.ReadUint32(pubData, offset)
	offset, creatorAccountIndex := common2.ReadUint32(pubData, offset)
	offset, creatorTreasuryRate := common2.ReadUint16(pubData, offset)
	offset, nftIndex := common2.ReadUint40(pubData, offset)
	offset, collectionId := common2.ReadUint16(pubData, offset)
	offset, toAddress := common2.ReadAddress(pubData, offset)
	offset, gasFeeAssetId := common2.ReadUint16(pubData, offset)
	offset, gasFeeAssetPackedAmount := common2.ReadUint16(pubData, offset)
	gasFeeAssetAmount, err := util.CleanPackedFee(big.NewInt(int64(gasFeeAssetPackedAmount)))
	if err != nil {
		return err
	}
	offset, nftContentHash := common2.ReadPrefixPaddingBufToChunkSize(pubData, offset)
	offset, creatorAccountNameHash := common2.ReadPrefixPaddingBufToChunkSize(pubData, offset)
	txInfo := &txtypes.WithdrawNftTxInfo{
		AccountIndex:           int64(fromAccountIndex),
		CreatorAccountIndex:    int64(creatorAccountIndex),
		CreatorTreasuryRate:    int64(creatorTreasuryRate),
		NftIndex:               nftIndex,
		ToAddress:              toAddress,
		CollectionId:           int64(collectionId),
		NftContentHash:         nftContentHash,
		CreatorAccountNameHash: creatorAccountNameHash,
		GasFeeAssetId:          int64(gasFeeAssetId),
		GasFeeAssetAmount:      gasFeeAssetAmount,
	}
	executor := &executor.WithdrawNftExecutor{
		BaseExecutor: executor.NewBaseExecutor(bc, nil, txInfo),
		TxInfo:       txInfo,
	}
	err = executor.Prepare()
	if err != nil {
		return err
	}
	err = executor.ApplyTransaction()
	if err != nil {
		return err
	}
	return nil
}

func (c *ExodusExit) getMerkleProofs(blockHeight int64, accountIndex int64, nftIndex int64, accountAssetId int64) error {
	treeCtx, err := tree.NewContext("generateproof", c.config.TreeDB.Driver, true, true, c.config.TreeDB.RoutinePoolSize, &c.config.TreeDB.LevelDBOption, &c.config.TreeDB.RedisDBOption)
	if err != nil {
		logx.Errorf("init tree database failed: %s", err)
		return err
	}

	treeCtx.SetOptions(bsmt.InitializeVersion(0))
	treeCtx.SetBatchReloadSize(1000)
	err = tree.SetupTreeDB(treeCtx)
	if err != nil {
		logx.Errorf("init tree database failed: %s", err)
		return err
	}

	// dbinitializer accountTree and accountStateTrees
	accountTree, accountAssetTrees, err := tree.InitAccountTree(
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
		return err
	}
	accountStateRoot := common.Bytes2Hex(accountTree.Root())
	logx.Infof("account tree accountStateRoot=%s", accountStateRoot)
	// dbinitializer nftTree
	nftTree, err := tree.InitNftTree(
		c.bc.L2NftModel,
		c.bc.L2NftHistoryModel,
		blockHeight,
		treeCtx, false)
	if err != nil {
		logx.Errorf("init nft tree error: %s", err.Error())
		return err
	}
	nftStateRoot := common.Bytes2Hex(nftTree.Root())
	logx.Infof("nft tree nftStateRoot=%s", nftStateRoot)
	stateRoot := tree.ComputeStateRootHash(accountTree.Root(), nftTree.Root())
	logx.Infof("nft tree nftStateRoot=%s", common.Bytes2Hex(stateRoot))
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

	if accountAssetId != -1 {
		assetMerkleProof, err := accountAssetTrees.Get(accountIndex).GetProof(uint64(accountAssetId))
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
		logx.Infof("accountIndex=%d,accountAssetId=%d, merkleProofsAccountAsset=%s", accountIndex, accountAssetId, string(merkleProofsAccountAssetBytes))
	}

	if nftIndex != -1 {
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
	}
	return nil
}
