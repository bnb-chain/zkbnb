package restorer

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"gorm.io/gorm"

	"github.com/bnb-chain/zkbas-crypto/ffmath"
	zkbas "github.com/bnb-chain/zkbas-eth-rpc/zkbas/core/legend"
	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/model/account"
	"github.com/bnb-chain/zkbas/common/model/assetInfo"
	"github.com/bnb-chain/zkbas/common/model/block"
	"github.com/bnb-chain/zkbas/common/model/liquidity"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/common/model/sysconfig"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/common/sysconfigName"
	"github.com/bnb-chain/zkbas/common/tree"
	"github.com/bnb-chain/zkbas/common/util"
)

const (
	Layer1FilterStep  = 100
	CommitChannelSize = 16
	OfferPerAsset     = 128
	TxPubDataLength   = 192
	ChunkBytesSize    = 32
)

var (
	BlockCommitTopic = crypto.Keccak256Hash([]byte("BlockCommit(uint32)"))
	BlockVerifyTopic = crypto.Keccak256Hash([]byte("BlockVerification(uint32)"))
	BlockRevertTopic = crypto.Keccak256Hash([]byte("BlocksRevert(uint32,uint32)"))

	ErrNotFound      = sqlx.ErrNotFound
	ZeroBigInt       = big.NewInt(0)
	ZeroBigIntString = "0"
)

type (
	RegisterZnsTxInfo      = commonTx.RegisterZnsTxInfo
	CreatePairTxInfo       = commonTx.CreatePairTxInfo
	UpdatePairRateTxInfo   = commonTx.UpdatePairRateTxInfo
	DepositTxInfo          = commonTx.DepositTxInfo
	DepositNftTxInfo       = commonTx.DepositNftTxInfo
	FullExitTxInfo         = commonTx.FullExitTxInfo
	FullExitNftTxInfo      = commonTx.FullExitNftTxInfo
	TransferTxInfo         = commonTx.TransferTxInfo
	SwapTxInfo             = commonTx.SwapTxInfo
	AddLiquidityTxInfo     = commonTx.AddLiquidityTxInfo
	RemoveLiquidityTxInfo  = commonTx.RemoveLiquidityTxInfo
	WithdrawTxInfo         = commonTx.WithdrawTxInfo
	CreateCollectionTxInfo = commonTx.CreateCollectionTxInfo
	MintNftTxInfo          = commonTx.MintNftTxInfo
	TransferNftTxInfo      = commonTx.TransferNftTxInfo
	OfferTxInfo            = commonTx.OfferTxInfo
	AtomicMatchTxInfo      = commonTx.AtomicMatchTxInfo
	CancelOfferTxInfo      = commonTx.CancelOfferTxInfo
	WithdrawNftTxInfo      = commonTx.WithdrawNftTxInfo
)

type Config struct {
	Postgres struct {
		DataSource string
	}
	CacheRedis cache.CacheConf
}

type ChainStorage struct {
	SysconfigModel        sysconfig.SysconfigModel
	BlockModel            block.BlockModel
	L2AssetInfoModel      assetInfo.AssetInfoModel
	AccountModel          account.AccountModel
	AccountHistoryModel   account.AccountHistoryModel
	LiquidityModel        liquidity.LiquidityModel
	LiquidityHistoryModel liquidity.LiquidityHistoryModel
	L2NftModel            nft.L2NftModel
	L2NftHistoryModel     nft.L2NftHistoryModel
}

type BlockReplayState struct {
	blockNumber int64
	txs         []*tx.Tx
	stateRoot   string

	pendingNewAccountIndexMap       map[int64]bool
	pendingNewLiquidityInfoIndexMap map[int64]bool
	pendingNewNftIndexMap           map[int64]bool

	pendingUpdateAccountIndexMap   map[int64]bool
	pendingUpdateLiquidityIndexMap map[int64]bool
	pendingUpdateNftIndexMap       map[int64]bool

	pendingNewNftWithdrawHistory []*nft.L2NftWithdrawHistory

	priorityOperations              int64
	pubDataOffset                   []uint32
	pendingOnChainOperationsPubData [][]byte
	pendingOnChainOperationsHash    []byte
}

type RestoreManager struct {
	Blockchain *ChainStorage

	accountMap   map[int64]*commonAsset.AccountInfo
	liquidityMap map[int64]*liquidity.Liquidity
	nftMap       map[int64]*nft.L2Nft

	accountTree       *tree.Tree
	liquidityTree     *tree.Tree
	nftTree           *tree.Tree
	accountAssetTrees []*tree.Tree

	l1Provider         *ethclient.Client
	l1Contract         common.Address
	l1ContractABI      abi.ABI
	l1ContractInstance *zkbas.Zkbas
	l1genesisNumber    *big.Int

	commitCh chan *zkbas.OldZkbasCommitBlockInfo
	verifyCh chan struct{}
	revertCh chan struct{}
	quitCh   chan struct{}
}

func NewRestoreManager(db *gorm.DB, conn sqlx.SqlConn, redisConn *redis.Redis, cacheConf cache.CacheConf) (*RestoreManager, error) {
	chainStorage := &ChainStorage{
		SysconfigModel:        sysconfig.NewSysconfigModel(conn, cacheConf, db),
		BlockModel:            block.NewBlockModel(conn, cacheConf, db, redisConn),
		L2AssetInfoModel:      assetInfo.NewAssetInfoModel(conn, cacheConf, db),
		AccountModel:          account.NewAccountModel(conn, cacheConf, db),
		AccountHistoryModel:   account.NewAccountHistoryModel(conn, cacheConf, db),
		LiquidityModel:        liquidity.NewLiquidityModel(conn, cacheConf, db),
		LiquidityHistoryModel: liquidity.NewLiquidityHistoryModel(conn, cacheConf, db),
		L2NftModel:            nft.NewL2NftModel(conn, cacheConf, db),
		L2NftHistoryModel:     nft.NewL2NftHistoryModel(conn, cacheConf, db),
	}

	l1ProviderConfig, err := chainStorage.SysconfigModel.GetSysconfigByName(sysconfigName.BscTestNetworkRpc)
	if err != nil {
		return nil, fmt.Errorf("get %s failed: %v", sysconfigName.BscTestNetworkRpc, err)
	}
	l1Provider, err := ethclient.Dial(l1ProviderConfig.Value)
	if err != nil {
		return nil, fmt.Errorf("dial failed, l1 provider: %s, error: %v", l1ProviderConfig.Value, err)
	}
	l1ContractConfig, err := chainStorage.SysconfigModel.GetSysconfigByName(sysconfigName.ZkbasContract)
	if err != nil {
		return nil, fmt.Errorf("get %s failed: %v", sysconfigName.ZkbasContract, err)
	}
	l1Contract := common.HexToAddress(l1ContractConfig.Value)
	l1ContractInstance, err := zkbas.NewZkbas(l1Contract, l1Provider)
	if err != nil {
		return nil, fmt.Errorf("new zecrey legend instance failed: %v", err)
	}
	l1ContractABI, err := abi.JSON(strings.NewReader(string(zkbas.ZkbasABI)))
	if err != nil {
		return nil, fmt.Errorf("parse layer1 contract abi failed: %v", err)
	}
	l1GenesisNumberConfig, err := chainStorage.SysconfigModel.GetSysconfigByName(sysconfigName.ZkbasGenesisNumber)
	if err != nil {
		return nil, fmt.Errorf("get %s failed: %v", sysconfigName.ZkbasGenesisNumber, err)
	}
	l1GenesisNumber, err := strconv.ParseInt(l1GenesisNumberConfig.Value, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse l1 genesis number failed: %v", err)
	}

	height, err := chainStorage.BlockModel.GetCurrentBlockHeight()
	if err != nil {
		return nil, fmt.Errorf("get current block height failed: %v", err)
	}
	accountTree, accountAssetTrees, err := tree.InitAccountTree(chainStorage.AccountModel, chainStorage.AccountHistoryModel, height)
	if err != nil {
		return nil, fmt.Errorf("init account tree failed: %v", err)
	}
	nftTree, err := tree.InitNftTree(chainStorage.L2NftHistoryModel, height)
	if err != nil {
		return nil, fmt.Errorf("init nft tree failed: %v", err)
	}
	liquidityTree, err := tree.InitLiquidityTree(chainStorage.LiquidityHistoryModel, height)
	if err != nil {
		return nil, fmt.Errorf("init liquidity tree failed: %v", err)
	}

	manager := &RestoreManager{
		Blockchain: chainStorage,

		accountMap:   make(map[int64]*commonAsset.AccountInfo),
		liquidityMap: make(map[int64]*liquidity.Liquidity),
		nftMap:       make(map[int64]*nft.L2Nft),

		accountTree:       accountTree,
		liquidityTree:     liquidityTree,
		nftTree:           nftTree,
		accountAssetTrees: accountAssetTrees,

		l1Provider:         l1Provider,
		l1Contract:         l1Contract,
		l1ContractABI:      l1ContractABI,
		l1ContractInstance: l1ContractInstance,
		l1genesisNumber:    big.NewInt(l1GenesisNumber),

		commitCh: make(chan *zkbas.OldZkbasCommitBlockInfo, CommitChannelSize),
		quitCh:   make(chan struct{}),
	}
	go manager.loop()
	return manager, nil
}

func (m *RestoreManager) loop() {
	defer func() {
		if m.quitCh != nil {
			close(m.quitCh)
		}
	}()

	blockNumber, err := m.Blockchain.BlockModel.GetCurrentBlockHeight()
	if err != nil && err != block.ErrNotFound {
		logx.Errorf("Get current block height failed :%s", err.Error())
		return
	}
	block, err := m.Blockchain.BlockModel.GetBlockByBlockHeight(blockNumber)
	if err != nil {
		logx.Errorf("Get block failed: %s, block number: %d", err.Error(), blockNumber)
	}

	for {
		select {
		case commitData := <-m.commitCh:
			if commitData.BlockNumber != uint32(blockNumber)+1 {
				logx.Errorf("unexpected block, expect: %d, real: %d", blockNumber+1, commitData.BlockNumber)
				return
			}
			blockNumber += 1
			block, err = m.replayBlockPubData(block, commitData)
			if err != nil {
				logx.Errorf("replay block public data failed: %s", err.Error())
				return
			}

		case <-m.quitCh:
			logx.Info("restorer manager loop exit")
			return
		}
	}
}

func (m *RestoreManager) replayBlockPubData(lastBlock *block.Block, commitData *zkbas.OldZkbasCommitBlockInfo) (*block.Block, error) {
	pubData := commitData.PublicData
	expectRoot := common.Bytes2Hex(commitData.NewStateRoot[:])
	createTime := commitData.Timestamp.Int64()
	if len(pubData)%TxPubDataLength != 0 {
		return nil, fmt.Errorf("wrong block public data length: %d", len(pubData))
	}

	txCount := len(pubData) / TxPubDataLength
	state := &BlockReplayState{
		blockNumber: lastBlock.BlockHeight + 1,
		txs:         make([]*tx.Tx, 0, txCount),

		pendingNewAccountIndexMap:       make(map[int64]bool),
		pendingNewLiquidityInfoIndexMap: make(map[int64]bool),
		pendingNewNftIndexMap:           make(map[int64]bool),

		pendingUpdateAccountIndexMap:   make(map[int64]bool),
		pendingUpdateLiquidityIndexMap: make(map[int64]bool),
		pendingUpdateNftIndexMap:       make(map[int64]bool),

		pendingNewNftWithdrawHistory: make([]*nft.L2NftWithdrawHistory, 0),

		priorityOperations:              0,
		pubDataOffset:                   make([]uint32, 0),
		pendingOnChainOperationsPubData: make([][]byte, 0),
		pendingOnChainOperationsHash:    common.FromHex(util.EmptyStringKeccak),
	}
	for i := 0; i < txCount; i++ {
		err := m.replayTxPubData(state, pubData[i*TxPubDataLength:(i+1)*TxPubDataLength])
		if err != nil {
			return nil, fmt.Errorf("replay tx pubdata failed: %s, tx index: %d", err.Error(), i)
		}
	}
	if state.stateRoot != expectRoot {
		logx.Errorf("unexpected state root, block number: %d, expect: %s, real: %s", state.blockNumber, expectRoot, state.stateRoot)
		return nil, fmt.Errorf("state root not expect")
	}

	var (
		pendingNewAccounts         []*account.Account
		pendingUpdateAccounts      []*account.Account
		pendingNewAccountHistory   []*account.AccountHistory
		pendingNewLiquidity        []*liquidity.Liquidity
		pendingUpdateLiquidity     []*liquidity.Liquidity
		pendingNewLiquidityHistory []*liquidity.LiquidityHistory
		pendingNewNft              []*nft.L2Nft
		pendingUpdateNft           []*nft.L2Nft
		pendingNewNftHistory       []*nft.L2NftHistory
	)

	for accountIndex, flag := range state.pendingNewAccountIndexMap {
		if !flag {
			continue
		}
		accountInfo, err := commonAsset.FromFormatAccountInfo(m.accountMap[accountIndex])
		if err != nil {
			return nil, fmt.Errorf("convert format account info failed: %v", err)
		}
		pendingNewAccounts = append(pendingNewAccounts, accountInfo)
		pendingNewAccountHistory = append(pendingNewAccountHistory, &account.AccountHistory{
			AccountIndex:    accountInfo.AccountIndex,
			Nonce:           accountInfo.Nonce,
			CollectionNonce: accountInfo.CollectionNonce,
			AssetInfo:       accountInfo.AssetInfo,
			AssetRoot:       accountInfo.AssetRoot,
			L2BlockHeight:   state.blockNumber,
		})
	}
	for accountIndex, flag := range state.pendingUpdateAccountIndexMap {
		if !flag {
			continue
		}
		accountInfo, err := commonAsset.FromFormatAccountInfo(m.accountMap[accountIndex])
		if err != nil {
			return nil, fmt.Errorf("convert format account info failed: %v", err)
		}
		pendingUpdateAccounts = append(pendingUpdateAccounts, accountInfo)
		pendingNewAccountHistory = append(pendingNewAccountHistory, &account.AccountHistory{
			AccountIndex:    accountInfo.AccountIndex,
			Nonce:           accountInfo.Nonce,
			CollectionNonce: accountInfo.CollectionNonce,
			AssetInfo:       accountInfo.AssetInfo,
			AssetRoot:       accountInfo.AssetRoot,
			L2BlockHeight:   state.blockNumber,
		})
	}

	for pairIndex, flag := range state.pendingNewLiquidityInfoIndexMap {
		if !flag {
			continue
		}
		pendingNewLiquidity = append(pendingNewLiquidity, m.liquidityMap[pairIndex])
		pendingNewLiquidityHistory = append(pendingNewLiquidityHistory, &liquidity.LiquidityHistory{
			PairIndex:            m.liquidityMap[pairIndex].PairIndex,
			AssetAId:             m.liquidityMap[pairIndex].AssetAId,
			AssetA:               m.liquidityMap[pairIndex].AssetA,
			AssetBId:             m.liquidityMap[pairIndex].AssetBId,
			AssetB:               m.liquidityMap[pairIndex].AssetB,
			LpAmount:             m.liquidityMap[pairIndex].LpAmount,
			KLast:                m.liquidityMap[pairIndex].KLast,
			FeeRate:              m.liquidityMap[pairIndex].FeeRate,
			TreasuryAccountIndex: m.liquidityMap[pairIndex].TreasuryAccountIndex,
			TreasuryRate:         m.liquidityMap[pairIndex].TreasuryRate,
			L2BlockHeight:        state.blockNumber,
		})
	}
	for pairIndex, flag := range state.pendingUpdateLiquidityIndexMap {
		if !flag {
			continue
		}
		pendingUpdateLiquidity = append(pendingUpdateLiquidity, m.liquidityMap[pairIndex])
		pendingNewLiquidityHistory = append(pendingNewLiquidityHistory, &liquidity.LiquidityHistory{
			PairIndex:            m.liquidityMap[pairIndex].PairIndex,
			AssetAId:             m.liquidityMap[pairIndex].AssetAId,
			AssetA:               m.liquidityMap[pairIndex].AssetA,
			AssetBId:             m.liquidityMap[pairIndex].AssetBId,
			AssetB:               m.liquidityMap[pairIndex].AssetB,
			LpAmount:             m.liquidityMap[pairIndex].LpAmount,
			KLast:                m.liquidityMap[pairIndex].KLast,
			FeeRate:              m.liquidityMap[pairIndex].FeeRate,
			TreasuryAccountIndex: m.liquidityMap[pairIndex].TreasuryAccountIndex,
			TreasuryRate:         m.liquidityMap[pairIndex].TreasuryRate,
			L2BlockHeight:        state.blockNumber,
		})
	}

	for nftIndex, flag := range state.pendingNewNftIndexMap {
		if !flag {
			continue
		}
		pendingNewNft = append(pendingNewNft, m.nftMap[nftIndex])
		pendingNewNftHistory = append(pendingNewNftHistory, &nft.L2NftHistory{
			NftIndex:            m.nftMap[nftIndex].NftIndex,
			CreatorAccountIndex: m.nftMap[nftIndex].CreatorAccountIndex,
			OwnerAccountIndex:   m.nftMap[nftIndex].OwnerAccountIndex,
			NftContentHash:      m.nftMap[nftIndex].NftContentHash,
			NftL1Address:        m.nftMap[nftIndex].NftL1Address,
			NftL1TokenId:        m.nftMap[nftIndex].NftL1TokenId,
			CreatorTreasuryRate: m.nftMap[nftIndex].CreatorTreasuryRate,
			CollectionId:        m.nftMap[nftIndex].CollectionId,
			L2BlockHeight:       state.blockNumber,
		})
	}
	for nftIndex, flag := range state.pendingUpdateNftIndexMap {
		if !flag {
			continue
		}
		pendingUpdateNft = append(pendingUpdateNft, m.nftMap[nftIndex])
		pendingNewNftHistory = append(pendingNewNftHistory, &nft.L2NftHistory{
			NftIndex:            m.nftMap[nftIndex].NftIndex,
			CreatorAccountIndex: m.nftMap[nftIndex].CreatorAccountIndex,
			OwnerAccountIndex:   m.nftMap[nftIndex].OwnerAccountIndex,
			NftContentHash:      m.nftMap[nftIndex].NftContentHash,
			NftL1Address:        m.nftMap[nftIndex].NftL1Address,
			NftL1TokenId:        m.nftMap[nftIndex].NftL1TokenId,
			CreatorTreasuryRate: m.nftMap[nftIndex].CreatorTreasuryRate,
			CollectionId:        m.nftMap[nftIndex].CollectionId,
			L2BlockHeight:       state.blockNumber,
		})
	}

	commitment := util.CreateBlockCommitment(
		state.blockNumber,
		createTime,
		common.FromHex(lastBlock.StateRoot),
		common.FromHex(state.stateRoot),
		pubData,
		int64(len(state.pubDataOffset)),
	)
	block := &block.Block{
		Model: gorm.Model{
			CreatedAt: time.UnixMilli(createTime),
		},
		BlockSize:                    commitData.BlockSize,
		BlockCommitment:              commitment,
		BlockHeight:                  state.blockNumber,
		StateRoot:                    state.stateRoot,
		PriorityOperations:           state.priorityOperations,
		PendingOnChainOperationsHash: common.Bytes2Hex(state.pendingOnChainOperationsHash),
		Txs:                          state.txs,
		BlockStatus:                  block.StatusCommitted,
	}
	if len(state.pendingOnChainOperationsPubData) != 0 {
		onChainOperationsPubDataBytes, err := json.Marshal(state.pendingOnChainOperationsPubData)
		if err != nil {
			return nil, fmt.Errorf("marshal on chain operations pub data failed: %v", err)
		}
		block.PendingOnChainOperationsPubData = string(onChainOperationsPubDataBytes)
	}
	err := m.Blockchain.BlockModel.CreateBlockForRestorer(
		block,
		pendingNewAccounts,
		pendingUpdateAccounts,
		pendingNewAccountHistory,
		pendingNewLiquidity,
		pendingUpdateLiquidity,
		pendingNewLiquidityHistory,
		pendingNewNft,
		pendingUpdateNft,
		pendingNewNftHistory,
		state.pendingNewNftWithdrawHistory,
	)
	if err != nil {
		return nil, fmt.Errorf("create block for restorer failed: %v", err)
	}
	logx.Infof("Inserted block: %d, state root: %s\n", block.BlockHeight, block.StateRoot)

	return block, nil
}

func (m *RestoreManager) replayTxPubData(state *BlockReplayState, pubData []byte) error {
	txType := pubData[0]
	switch txType {
	case commonTx.TxTypeEmpty:
		// Do nothing
	case commonTx.TxTypeRegisterZns:
		err := m.replayRegisterZns(state, pubData)
		if err != nil {
			return fmt.Errorf("replay register zns failed: %v", err)
		}
	case commonTx.TxTypeCreatePair:
		err := m.replayCreatePair(state, pubData)
		if err != nil {
			return fmt.Errorf("replay create pair failed: %v", err)
		}
	case commonTx.TxTypeUpdatePairRate:
		err := m.replayUpdatePairRate(state, pubData)
		if err != nil {
			return fmt.Errorf("replay update pair rate failed: %v", err)
		}
	case commonTx.TxTypeDeposit:
		err := m.replayDeposit(state, pubData)
		if err != nil {
			return fmt.Errorf("replay deposit failed: %v", err)
		}
	case commonTx.TxTypeDepositNft:
		err := m.replayDepositNft(state, pubData)
		if err != nil {
			return fmt.Errorf("replay deposit nft failed: %v", err)
		}
	case commonTx.TxTypeTransfer:
		err := m.replayTransfer(state, pubData)
		if err != nil {
			return fmt.Errorf("replay transfer failed: %v", err)
		}
	case commonTx.TxTypeSwap:
		err := m.replaySwap(state, pubData)
		if err != nil {
			return fmt.Errorf("replay swap failed: %v", err)
		}
	case commonTx.TxTypeAddLiquidity:
		err := m.replayAddLiquidity(state, pubData)
		if err != nil {
			return fmt.Errorf("replay add liquidity failed: %v", err)
		}
	case commonTx.TxTypeRemoveLiquidity:
		err := m.replayRemoveLiquidity(state, pubData)
		if err != nil {
			return fmt.Errorf("replay remove liquidity failed: %v", err)
		}
	case commonTx.TxTypeWithdraw:
		err := m.replayWithdraw(state, pubData)
		if err != nil {
			return fmt.Errorf("replay withdraw failed: %v", err)
		}
	case commonTx.TxTypeCreateCollection:
		err := m.replayCreateCollection(state, pubData)
		if err != nil {
			return fmt.Errorf("replay create collection failed: %v", err)
		}
	case commonTx.TxTypeMintNft:
		err := m.replayMintNft(state, pubData)
		if err != nil {
			return fmt.Errorf("replay mint nft failed: %v", err)
		}
	case commonTx.TxTypeTransferNft:
		err := m.replayTransferNft(state, pubData)
		if err != nil {
			return fmt.Errorf("replay transfer nft failed: %v", err)
		}
	case commonTx.TxTypeAtomicMatch:
		err := m.replayAtomicMatch(state, pubData)
		if err != nil {
			return fmt.Errorf("replay atomic match failed: %v", err)
		}
	case commonTx.TxTypeCancelOffer:
		err := m.replayCancelOffer(state, pubData)
		if err != nil {
			return fmt.Errorf("replay cancel offer failed: %v", err)
		}
	case commonTx.TxTypeWithdrawNft:
		err := m.replayWithdrawNft(state, pubData)
		if err != nil {
			return fmt.Errorf("replay withdraw nft failed: %v", err)
		}
	case commonTx.TxTypeFullExit:
		err := m.replayFullExit(state, pubData)
		if err != nil {
			return fmt.Errorf("replay full exit failed: %v", err)
		}
	case commonTx.TxTypeFullExitNft:
		err := m.replayFullExitNft(state, pubData)
		if err != nil {
			return fmt.Errorf("replay full exit nft failed: %v", err)
		}
	case commonTx.TxTypeOffer:
		break
	default:
		return fmt.Errorf("unknown tx type, pubData: %s, tx type: %d", pubData, txType)
	}

	return nil
}

func (m *RestoreManager) prepareAccountsAndAssets(accounts []int64, assets []int64) error {
	for _, accountIndex := range accounts {
		if m.accountMap[accountIndex] == nil {
			accountInfo, err := m.Blockchain.AccountModel.GetAccountByAccountIndex(accountIndex)
			if err != nil {
				return err
			}
			m.accountMap[accountIndex], err = commonAsset.ToFormatAccountInfo(accountInfo)
			if err != nil {
				return fmt.Errorf("convert to format account info failed: %v", err)
			}
		}
		if m.accountMap[accountIndex].AssetInfo == nil {
			m.accountMap[accountIndex].AssetInfo = make(map[int64]*commonAsset.AccountAsset)
		}
		for _, assetId := range assets {
			if m.accountMap[accountIndex].AssetInfo[assetId] == nil {
				m.accountMap[accountIndex].AssetInfo[assetId] = &commonAsset.AccountAsset{
					AssetId:                  assetId,
					Balance:                  ZeroBigInt,
					LpAmount:                 ZeroBigInt,
					OfferCanceledOrFinalized: ZeroBigInt,
				}
			}
		}
	}

	return nil
}

func (m *RestoreManager) prepareLiquidity(pairIndex int64) error {
	if m.liquidityMap[pairIndex] == nil {
		liquidityInfo, err := m.Blockchain.LiquidityModel.GetLiquidityByPairIndex(pairIndex)
		if err != nil {
			return err
		}
		m.liquidityMap[pairIndex] = liquidityInfo
	}
	return nil
}

func (m *RestoreManager) prepareNft(nftIndex int64) error {
	if m.nftMap[nftIndex] == nil {
		nftAsset, err := m.Blockchain.L2NftModel.GetNftAsset(nftIndex)
		if err != nil {
			return err
		}
		m.nftMap[nftIndex] = nftAsset
	}
	return nil
}

func (m *RestoreManager) updateAccountTree(accounts []int64, assets []int64) error {
	for _, accountIndex := range accounts {
		for _, assetId := range assets {
			assetLeaf, err := tree.ComputeAccountAssetLeafHash(
				m.accountMap[accountIndex].AssetInfo[assetId].Balance.String(),
				m.accountMap[accountIndex].AssetInfo[assetId].LpAmount.String(),
				m.accountMap[accountIndex].AssetInfo[assetId].OfferCanceledOrFinalized.String(),
			)
			if err != nil {
				return fmt.Errorf("compute new account asset leaf failed: %v", err)
			}
			err = m.accountAssetTrees[accountIndex].Update(assetId, assetLeaf)
			if err != nil {
				return fmt.Errorf("update asset tree failed: %v", err)
			}
		}

		m.accountMap[accountIndex].AssetRoot = common.Bytes2Hex(m.accountAssetTrees[accountIndex].RootNode.Value)
		nAccountLeafHash, err := tree.ComputeAccountLeafHash(
			m.accountMap[accountIndex].AccountNameHash,
			m.accountMap[accountIndex].PublicKey,
			m.accountMap[accountIndex].Nonce,
			m.accountMap[accountIndex].CollectionNonce,
			m.accountAssetTrees[accountIndex].RootNode.Value,
		)
		if err != nil {
			return fmt.Errorf("unable to compute account leaf: %v", err)
		}
		err = m.accountTree.Update(accountIndex, nAccountLeafHash)
		if err != nil {
			return fmt.Errorf("unable to update account tree: %v", err)
		}
	}

	return nil
}

func (m *RestoreManager) updateLiquidityTree(pairIndex int64) error {
	nLiquidityAssetLeaf, err := tree.ComputeLiquidityAssetLeafHash(
		m.liquidityMap[pairIndex].AssetAId,
		m.liquidityMap[pairIndex].AssetA,
		m.liquidityMap[pairIndex].AssetBId,
		m.liquidityMap[pairIndex].AssetB,
		m.liquidityMap[pairIndex].LpAmount,
		m.liquidityMap[pairIndex].KLast,
		m.liquidityMap[pairIndex].FeeRate,
		m.liquidityMap[pairIndex].TreasuryAccountIndex,
		m.liquidityMap[pairIndex].TreasuryRate,
	)
	if err != nil {
		return fmt.Errorf("unable to compute liquidity leaf: %v", err)
	}
	err = m.liquidityTree.Update(pairIndex, nLiquidityAssetLeaf)
	if err != nil {
		return fmt.Errorf("unable to update liquidity tree: %v", err)
	}

	return nil
}

func (m *RestoreManager) updateNftTree(nftIndex int64) error {
	nftAssetLeaf, err := tree.ComputeNftAssetLeafHash(
		m.nftMap[nftIndex].CreatorAccountIndex,
		m.nftMap[nftIndex].OwnerAccountIndex,
		m.nftMap[nftIndex].NftContentHash,
		m.nftMap[nftIndex].NftL1Address,
		m.nftMap[nftIndex].NftL1TokenId,
		m.nftMap[nftIndex].CreatorTreasuryRate,
		m.nftMap[nftIndex].CollectionId,
	)
	if err != nil {
		return fmt.Errorf("unable to compute nft leaf: %v", err)
	}
	err = m.nftTree.Update(nftIndex, nftAssetLeaf)
	if err != nil {
		return fmt.Errorf("unable to update nft tree: %v", err)
	}

	return nil
}

func (m *RestoreManager) getStateRoot() string {
	hFunc := mimc.NewMiMC()
	hFunc.Write(m.accountTree.RootNode.Value)
	hFunc.Write(m.liquidityTree.RootNode.Value)
	hFunc.Write(m.nftTree.RootNode.Value)
	return common.Bytes2Hex(hFunc.Sum(nil))
}

func (m *RestoreManager) getL1Address(accountIndex int64, accountNameHash []byte) (string, error) {
	var nameHash [32]byte
	copy(nameHash[:], accountNameHash)
	l1Address, err := m.l1ContractInstance.GetAddressByAccountNameHash(&bind.CallOpts{}, nameHash)
	if err != nil {
		logx.Errorf("get address by account name hash failed: %v", err)
		return "", fmt.Errorf("get address by account name hash failed: %v", err)
	}

	return l1Address.String(), nil
}

func (m *RestoreManager) replayRegisterZns(state *BlockReplayState, pubData []byte) error {
	offset := 0
	offset, txType := util.ReadUint8(pubData, offset)
	offset, accountIndex := util.ReadUint32(pubData, offset)
	offset = 1 * ChunkBytesSize
	offset, accountName := util.ReadBytes32(pubData, offset)
	offset, accountNameHash := util.ReadBytes32(pubData, offset)
	offset, pubKeyX := util.ReadBytes32(pubData, offset)
	offset, pubKeyY := util.ReadBytes32(pubData, offset)
	pk := new(eddsa.PublicKey)
	pk.A.X.SetBytes(pubKeyX)
	pk.A.Y.SetBytes(pubKeyY)
	txInfo := RegisterZnsTxInfo{
		TxType:          txType,
		AccountIndex:    int64(accountIndex),
		AccountName:     util.CleanAccountName(util.SerializeAccountName(accountName)),
		AccountNameHash: accountNameHash,
		PubKey:          common.Bytes2Hex(pk.Bytes()),
	}
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return fmt.Errorf("unable to serialize tx info : %v", err)
	}

	// Check if the account name has been registered.
	_, err = m.Blockchain.AccountModel.GetAccountByAccountName(txInfo.AccountName)
	if err != ErrNotFound {
		return fmt.Errorf("account name has been registered")
	}
	if txInfo.AccountIndex != int64(len(m.accountAssetTrees)) {
		return fmt.Errorf("unepxect account index, expect: %d, real: %d", len(m.accountAssetTrees), txInfo.AccountIndex)
	}

	// Register new account.
	accountInfo := &account.Account{
		AccountIndex:    txInfo.AccountIndex,
		AccountName:     txInfo.AccountName,
		PublicKey:       txInfo.PubKey,
		AccountNameHash: common.Bytes2Hex(txInfo.AccountNameHash),
		// L1Address: get address by account name hash from L1 contract.
		Nonce:           commonConstant.NilNonce,
		CollectionNonce: commonConstant.NilNonce,
		AssetInfo:       commonConstant.NilAssetInfo,
		AssetRoot:       common.Bytes2Hex(tree.NilAccountAssetRoot),
		Status:          account.AccountStatusConfirmed,
	}
	accountInfo.L1Address, err = m.getL1Address(txInfo.AccountIndex, txInfo.AccountNameHash)
	if err != nil {
		return fmt.Errorf("get layer1 address failed: %v", err)
	}
	m.accountMap[txInfo.AccountIndex], err = commonAsset.ToFormatAccountInfo(accountInfo)
	if err != nil {
		return fmt.Errorf("convert to format account info failed: %v", err)
	}
	if int64(len(m.accountAssetTrees)) != txInfo.AccountIndex {
		return fmt.Errorf("unepxect account index from account asset tree, expect: %d, real: %d",
			int64(len(m.accountAssetTrees)), txInfo.AccountIndex)
	}
	emptyAssetTree, err := tree.NewEmptyAccountAssetTree()
	if err != nil {
		return fmt.Errorf("unable to new empty account state tree: %v", err)
	}
	m.accountAssetTrees = append(m.accountAssetTrees, emptyAssetTree)

	// Update trees and compute state root.
	err = m.updateAccountTree([]int64{txInfo.AccountIndex}, nil)
	if err != nil {
		return fmt.Errorf("update account tree failed: %v", err)
	}
	stateRoot := m.getStateRoot()

	tx := &tx.Tx{
		TxHash:        common.Hash{}.String(), // FIXME: need change tx hash
		TxType:        int64(txInfo.TxType),
		GasFee:        commonConstant.NilAssetAmountStr,
		GasFeeAssetId: commonConstant.NilAssetId,
		TxStatus:      mempool.SuccessTxStatus,
		BlockHeight:   state.blockNumber,
		StateRoot:     stateRoot,
		NftIndex:      commonConstant.NilTxNftIndex,
		PairIndex:     commonConstant.NilPairIndex,
		AssetId:       commonConstant.NilAssetId,
		TxAmount:      commonConstant.NilAssetAmountStr,
		NativeAddress: m.accountMap[txInfo.AccountIndex].L1Address, // TODO: think about update time
		TxInfo:        string(txInfoBytes),
		AccountIndex:  txInfo.AccountIndex,
		Nonce:         commonConstant.NilNonce,
		ExpiredAt:     commonConstant.NilExpiredAt,
	}
	state.txs = append(state.txs, tx)
	state.stateRoot = stateRoot
	state.pendingNewAccountIndexMap[txInfo.AccountIndex] = true
	state.priorityOperations += 1
	state.pubDataOffset = append(state.pubDataOffset, uint32(len(state.pubDataOffset)))
	return nil
}

func (m *RestoreManager) replayCreatePair(state *BlockReplayState, pubData []byte) error {
	offset := 0
	offset, txType := util.ReadUint8(pubData, offset)
	offset, pairIndex := util.ReadUint16(pubData, offset)
	offset, assetAId := util.ReadUint16(pubData, offset)
	offset, assetBId := util.ReadUint16(pubData, offset)
	offset, feeRate := util.ReadUint16(pubData, offset)
	offset, treasuryAccountIndex := util.ReadUint32(pubData, offset)
	offset, treasuryRate := util.ReadUint16(pubData, offset)
	txInfo := &CreatePairTxInfo{
		TxType:               txType,
		PairIndex:            int64(pairIndex),
		AssetAId:             int64(assetAId),
		AssetBId:             int64(assetBId),
		FeeRate:              int64(feeRate),
		TreasuryAccountIndex: int64(treasuryAccountIndex),
		TreasuryRate:         int64(treasuryRate),
	}
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return fmt.Errorf("unable to serialize tx info : %v", err)
	}

	_, err = m.Blockchain.LiquidityModel.GetLiquidityByPairIndex(txInfo.PairIndex)
	if err != ErrNotFound {
		return fmt.Errorf("unexpected get liquidity by pair index: %d err: %v", txInfo.PairIndex, err)
	}
	liquidityInfo := &liquidity.Liquidity{
		PairIndex:            txInfo.PairIndex,
		AssetAId:             txInfo.AssetAId,
		AssetA:               ZeroBigIntString,
		AssetBId:             txInfo.AssetBId,
		AssetB:               ZeroBigIntString,
		LpAmount:             ZeroBigIntString,
		KLast:                ZeroBigIntString,
		TreasuryAccountIndex: txInfo.TreasuryAccountIndex,
		FeeRate:              txInfo.FeeRate,
		TreasuryRate:         txInfo.TreasuryRate,
	}
	m.liquidityMap[txInfo.PairIndex] = liquidityInfo

	// Update trees and compute state root.
	err = m.updateLiquidityTree(txInfo.PairIndex)
	if err != nil {
		return fmt.Errorf("update liquidity tree failed: %v", err)
	}
	stateRoot := m.getStateRoot()

	poolInfo := &commonAsset.LiquidityInfo{
		PairIndex:            txInfo.PairIndex,
		AssetAId:             txInfo.AssetAId,
		AssetA:               big.NewInt(0),
		AssetBId:             txInfo.AssetBId,
		AssetB:               big.NewInt(0),
		LpAmount:             big.NewInt(0),
		KLast:                big.NewInt(0),
		FeeRate:              txInfo.FeeRate,
		TreasuryAccountIndex: txInfo.TreasuryAccountIndex,
		TreasuryRate:         txInfo.TreasuryRate,
	}
	txDetail := &tx.TxDetail{
		AssetId:      txInfo.PairIndex,
		AssetType:    commonAsset.LiquidityAssetType,
		AccountIndex: commonConstant.NilTxAccountIndex,
		AccountName:  commonConstant.NilAccountName,
		BalanceDelta: poolInfo.String(),
		Order:        0,
		AccountOrder: commonConstant.NilAccountOrder,
	}
	tx := &tx.Tx{
		TxHash:        common.Hash{}.String(), // FIXME: need change tx hash
		TxType:        int64(txInfo.TxType),
		GasFee:        commonConstant.NilAssetAmountStr,
		GasFeeAssetId: commonConstant.NilAssetId,
		TxStatus:      mempool.SuccessTxStatus,
		BlockHeight:   state.blockNumber,
		StateRoot:     stateRoot,
		NftIndex:      commonConstant.NilTxNftIndex,
		PairIndex:     txInfo.PairIndex,
		AssetId:       commonConstant.NilAssetId,
		TxAmount:      commonConstant.NilAssetAmountStr,
		NativeAddress: commonConstant.NilL1Address,
		TxInfo:        string(txInfoBytes),
		TxDetails:     []*tx.TxDetail{txDetail},
		AccountIndex:  commonConstant.NilTxAccountIndex,
		Nonce:         commonConstant.NilNonce,
		ExpiredAt:     commonConstant.NilExpiredAt,
	}
	state.txs = append(state.txs, tx)
	state.stateRoot = stateRoot
	state.pendingNewLiquidityInfoIndexMap[txInfo.PairIndex] = true
	state.priorityOperations += 1
	state.pubDataOffset = append(state.pubDataOffset, uint32(len(state.pubDataOffset)))
	return nil
}

func (m *RestoreManager) replayUpdatePairRate(state *BlockReplayState, pubData []byte) error {
	offset := 0
	offset, txType := util.ReadUint8(pubData, offset)
	offset, pairIndex := util.ReadUint16(pubData, offset)
	offset, feeRate := util.ReadUint16(pubData, offset)
	offset, treasuryAccountIndex := util.ReadUint32(pubData, offset)
	offset, treasuryRate := util.ReadUint16(pubData, offset)
	txInfo := &UpdatePairRateTxInfo{
		TxType:               txType,
		PairIndex:            int64(pairIndex),
		FeeRate:              int64(feeRate),
		TreasuryAccountIndex: int64(treasuryAccountIndex),
		TreasuryRate:         int64(treasuryRate),
	}
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return fmt.Errorf("unable to serialize tx info : %v", err)
	}

	err = m.prepareLiquidity(txInfo.PairIndex)
	if err != nil {
		return fmt.Errorf("prepare liquidity failed: %v", err)
	}
	liquidityInfo := m.liquidityMap[txInfo.PairIndex]
	liquidityInfo.FeeRate = txInfo.FeeRate
	liquidityInfo.TreasuryAccountIndex = txInfo.TreasuryAccountIndex
	liquidityInfo.TreasuryRate = txInfo.TreasuryRate

	// Update trees and compute state root.
	err = m.updateLiquidityTree(txInfo.PairIndex)
	if err != nil {
		return fmt.Errorf("update liquidity tree failed: %v", err)
	}
	stateRoot := m.getStateRoot()

	poolInfo, err := commonAsset.ConstructLiquidityInfo(
		liquidityInfo.PairIndex,
		liquidityInfo.AssetAId,
		liquidityInfo.AssetA,
		liquidityInfo.AssetBId,
		liquidityInfo.AssetB,
		liquidityInfo.LpAmount,
		liquidityInfo.KLast,
		liquidityInfo.FeeRate,
		liquidityInfo.TreasuryAccountIndex,
		liquidityInfo.TreasuryRate,
	)
	if err != nil {
		return fmt.Errorf("unable to construct liquidity info: %v", err)
	}
	txDetail := &tx.TxDetail{
		AssetId:      txInfo.PairIndex,
		AssetType:    commonAsset.LiquidityAssetType,
		AccountIndex: commonConstant.NilTxAccountIndex,
		AccountName:  commonConstant.NilAccountName,
		BalanceDelta: poolInfo.String(),
		Order:        0,
		AccountOrder: commonConstant.NilAccountOrder,
	}
	tx := &tx.Tx{
		TxHash:        common.Hash{}.String(), // FIXME: need change tx hash
		TxType:        int64(txInfo.TxType),
		GasFee:        commonConstant.NilAssetAmountStr,
		GasFeeAssetId: commonConstant.NilAssetId,
		TxStatus:      mempool.SuccessTxStatus,
		BlockHeight:   state.blockNumber,
		StateRoot:     stateRoot,
		NftIndex:      commonConstant.NilTxNftIndex,
		PairIndex:     txInfo.PairIndex,
		AssetId:       commonConstant.NilAssetId,
		TxAmount:      commonConstant.NilAssetAmountStr,
		NativeAddress: commonConstant.NilL1Address,
		TxInfo:        string(txInfoBytes),
		TxDetails:     []*tx.TxDetail{txDetail},
		AccountIndex:  commonConstant.NilTxAccountIndex,
		Nonce:         commonConstant.NilNonce,
		ExpiredAt:     commonConstant.NilExpiredAt,
	}
	state.txs = append(state.txs, tx)
	state.stateRoot = stateRoot
	state.pendingUpdateLiquidityIndexMap[txInfo.PairIndex] = true
	state.priorityOperations += 1
	state.pubDataOffset = append(state.pubDataOffset, uint32(len(state.pubDataOffset)))
	return nil
}

func (m *RestoreManager) replayDeposit(state *BlockReplayState, pubData []byte) error {
	offset := 0
	offset, txType := util.ReadUint8(pubData, offset)
	offset, accountIndex := util.ReadUint32(pubData, offset)
	offset, assetId := util.ReadUint16(pubData, offset)
	offset, amount := util.ReadUint128(pubData, offset)
	offset = 1 * ChunkBytesSize
	offset, accountNameHash := util.ReadBytes32(pubData, offset)
	txInfo := &DepositTxInfo{
		TxType:          txType,
		AccountIndex:    int64(accountIndex),
		AccountNameHash: accountNameHash,
		AssetId:         int64(assetId),
		AssetAmount:     amount,
	}
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return fmt.Errorf("unable to serialize tx info : %v", err)
	}

	err = m.prepareAccountsAndAssets([]int64{txInfo.AccountIndex}, []int64{txInfo.AssetId})
	if err != nil {
		return fmt.Errorf("prepare accounts failed: %v", err)
	}
	accountInfo := m.accountMap[txInfo.AccountIndex]
	accountInfo.AssetInfo[txInfo.AssetId].Balance = ffmath.Add(accountInfo.AssetInfo[txInfo.AssetId].Balance, txInfo.AssetAmount)

	// Update trees and compute state root.
	err = m.updateAccountTree([]int64{txInfo.AccountIndex}, []int64{txInfo.AssetId})
	if err != nil {
		return fmt.Errorf("update account tree failed: %v", err)
	}
	stateRoot := m.getStateRoot()

	tx := &tx.Tx{
		TxHash:        common.Hash{}.String(), // FIXME: need change tx hash
		TxType:        int64(txInfo.TxType),
		GasFee:        commonConstant.NilAssetAmountStr,
		GasFeeAssetId: commonConstant.NilAssetId,
		TxStatus:      mempool.SuccessTxStatus,
		BlockHeight:   state.blockNumber,
		StateRoot:     stateRoot,
		NftIndex:      commonConstant.NilTxNftIndex,
		PairIndex:     commonConstant.NilPairIndex,
		AssetId:       txInfo.AssetId,
		TxAmount:      txInfo.AssetAmount.String(),
		NativeAddress: m.accountMap[txInfo.AccountIndex].L1Address,
		TxInfo:        string(txInfoBytes),
		//TxDetails:     []*tx.TxDetail{txDetail},
		AccountIndex: txInfo.AccountIndex,
		Nonce:        commonConstant.NilNonce,
		ExpiredAt:    commonConstant.NilExpiredAt,
	}
	state.txs = append(state.txs, tx)
	state.stateRoot = stateRoot
	state.pendingUpdateAccountIndexMap[txInfo.AccountIndex] = true
	state.priorityOperations++
	state.pubDataOffset = append(state.pubDataOffset, uint32(len(state.pubDataOffset)))
	return nil
}

func (m *RestoreManager) replayDepositNft(state *BlockReplayState, pubData []byte) error {
	offset := 0
	offset, txType := util.ReadUint8(pubData, offset)
	offset, accountIndex := util.ReadUint32(pubData, offset)
	offset, nftIndex := util.ReadUint40(pubData, offset)
	offset, nftL1Address := util.ReadAddress(pubData, offset)
	offset = 1 * ChunkBytesSize
	offset, creatorAccountIndex := util.ReadUint32(pubData, offset) // TODO: Prefix Padding, maybe wrong here.
	offset, creatorTreasuryRate := util.ReadUint16(pubData, offset)
	offset, collectionId := util.ReadUint16(pubData, offset)
	offset = 2 * ChunkBytesSize
	offset, nftContentHash := util.ReadBytes32(pubData, offset)
	offset, nftL1TokenId := util.ReadUint256(pubData, offset)
	offset, accountNameHash := util.ReadBytes32(pubData, offset)
	txInfo := &DepositNftTxInfo{
		TxType:              txType,
		AccountIndex:        int64(accountIndex),
		NftIndex:            nftIndex,
		NftL1Address:        nftL1Address,
		CreatorAccountIndex: int64(creatorAccountIndex),
		CreatorTreasuryRate: int64(creatorTreasuryRate),
		NftContentHash:      nftContentHash,
		NftL1TokenId:        nftL1TokenId,
		AccountNameHash:     accountNameHash,
		CollectionId:        int64(collectionId),
	}
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return fmt.Errorf("unable to serialize tx info : %v", err)
	}

	_, err = m.Blockchain.L2NftModel.GetNftAsset(txInfo.NftIndex)
	if err != ErrNotFound {
		return fmt.Errorf("unexpected get nft asset result, index: %d, result: %v", txInfo.NftIndex, err)
	}
	err = m.prepareAccountsAndAssets([]int64{txInfo.AccountIndex}, nil)
	if err != nil {
		return fmt.Errorf("prepare accounts failed: %v", err)
	}
	accountInfo := m.accountMap[txInfo.AccountIndex]
	nftAsset := &nft.L2Nft{
		NftIndex:            txInfo.NftIndex,
		CreatorAccountIndex: txInfo.CreatorAccountIndex,
		OwnerAccountIndex:   accountInfo.AccountIndex,
		NftContentHash:      common.Bytes2Hex(txInfo.NftContentHash),
		NftL1Address:        txInfo.NftL1Address,
		NftL1TokenId:        txInfo.NftL1TokenId.String(),
		CreatorTreasuryRate: txInfo.CreatorTreasuryRate,
		CollectionId:        txInfo.CollectionId,
	}
	m.nftMap[txInfo.NftIndex] = nftAsset

	// Update trees and compute state root.
	err = m.updateNftTree(txInfo.NftIndex)
	if err != nil {
		return fmt.Errorf("update nft tree failed: %v", err)
	}
	stateRoot := m.getStateRoot()

	// Construct tx.
	var txDetails []*tx.TxDetail
	nftInfo := commonAsset.ConstructNftInfo(
		txInfo.NftIndex,
		txInfo.CreatorAccountIndex,
		accountInfo.AccountIndex,
		common.Bytes2Hex(txInfo.NftContentHash),
		txInfo.NftL1TokenId.String(),
		txInfo.NftL1Address,
		txInfo.CreatorTreasuryRate,
		txInfo.CollectionId,
	)
	emptyDeltaAsset := &commonAsset.AccountAsset{
		AssetId:                  0,
		Balance:                  big.NewInt(0),
		LpAmount:                 big.NewInt(0),
		OfferCanceledOrFinalized: big.NewInt(0),
	}
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      0,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.AccountIndex,
		AccountName:  accountInfo.AccountName,
		BalanceDelta: emptyDeltaAsset.String(),
		AccountOrder: 0,
	})
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.NftIndex,
		AssetType:    commonAsset.NftAssetType,
		AccountIndex: txInfo.AccountIndex,
		AccountName:  accountInfo.AccountName,
		BalanceDelta: nftInfo.String(),
		AccountOrder: commonConstant.NilAccountOrder,
	})
	tx := &tx.Tx{
		TxHash:        common.Hash{}.String(), // FIXME: need change tx hash
		TxType:        int64(txInfo.TxType),
		GasFee:        commonConstant.NilAssetAmountStr,
		GasFeeAssetId: commonConstant.NilAssetId,
		TxStatus:      mempool.SuccessTxStatus,
		BlockHeight:   state.blockNumber,
		StateRoot:     stateRoot,
		NftIndex:      txInfo.NftIndex,
		PairIndex:     commonConstant.NilPairIndex,
		AssetId:       commonConstant.NilAssetId,
		TxAmount:      commonConstant.NilAssetAmountStr,
		NativeAddress: m.accountMap[txInfo.AccountIndex].L1Address,
		TxInfo:        string(txInfoBytes),
		TxDetails:     txDetails,
		AccountIndex:  txInfo.AccountIndex,
		Nonce:         commonConstant.NilNonce,
		ExpiredAt:     commonConstant.NilExpiredAt,
	}
	state.txs = append(state.txs, tx)
	state.stateRoot = stateRoot
	state.pendingNewNftIndexMap[txInfo.NftIndex] = true
	state.priorityOperations += 1
	state.pubDataOffset = append(state.pubDataOffset, uint32(len(state.pubDataOffset)))
	return nil
}

func (m *RestoreManager) replayTransfer(state *BlockReplayState, pubData []byte) error {
	offset := 0
	offset, txType := util.ReadUint8(pubData, offset)
	offset, fromAccountIndex := util.ReadUint32(pubData, offset)
	offset, toAccountIndex := util.ReadUint32(pubData, offset)
	offset, assetId := util.ReadUint16(pubData, offset)
	offset, packedAssetAmount := util.ReadUint40(pubData, offset)
	offset, gasAccountIndex := util.ReadUint32(pubData, offset)
	offset, gasFeeAssetId := util.ReadUint16(pubData, offset)
	offset, packedGasFeeAssetAmount := util.ReadUint16(pubData, offset)
	offset = 1 * ChunkBytesSize
	offset, callDataHash := util.ReadBytes32(pubData, offset)
	assetAmount, err := util.FromPackedAmount(packedAssetAmount)
	if err != nil {
		return fmt.Errorf("parse packed asset amount failed: %v", err)
	}
	gasFeeAssetAmount, err := util.FromPackedFee(int64(packedGasFeeAssetAmount))
	if err != nil {
		return fmt.Errorf("parse packed fee amount failed: %v", err)
	}
	txInfo := &TransferTxInfo{
		FromAccountIndex:  int64(fromAccountIndex),
		ToAccountIndex:    int64(toAccountIndex),
		AssetId:           int64(assetId),
		AssetAmount:       assetAmount,
		GasAccountIndex:   int64(gasAccountIndex),
		GasFeeAssetId:     int64(gasFeeAssetId),
		GasFeeAssetAmount: gasFeeAssetAmount,
		CallDataHash:      callDataHash,
	}

	// Prepare account map and asset info.
	err = m.prepareAccountsAndAssets([]int64{txInfo.FromAccountIndex, txInfo.ToAccountIndex, txInfo.GasAccountIndex}, []int64{txInfo.AssetId, txInfo.GasFeeAssetId})
	if err != nil {
		return fmt.Errorf("prepare accounts failed: %v", err)
	}

	// Update asset info and asset tree.
	fromAccountInfo := m.accountMap[txInfo.FromAccountIndex]
	toAccountInfo := m.accountMap[txInfo.ToAccountIndex]
	gasAccountInfo := m.accountMap[txInfo.GasAccountIndex]
	txInfo.Nonce = fromAccountInfo.Nonce
	fromAccountInfo.Nonce++
	fromAccountInfo.AssetInfo[txInfo.AssetId].Balance = new(big.Int).Sub(fromAccountInfo.AssetInfo[txInfo.AssetId].Balance, txInfo.AssetAmount)
	fromAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance = new(big.Int).Sub(fromAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	toAccountInfo.AssetInfo[txInfo.AssetId].Balance = new(big.Int).Add(toAccountInfo.AssetInfo[txInfo.AssetId].Balance, txInfo.AssetAmount)
	gasAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance = new(big.Int).Add(gasAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)

	// Update trees and compute state root.
	err = m.updateAccountTree([]int64{txInfo.FromAccountIndex, txInfo.ToAccountIndex, txInfo.GasAccountIndex}, []int64{txInfo.AssetId, txInfo.GasFeeAssetId})
	if err != nil {
		return fmt.Errorf("update account tree failed: %v", err)
	}
	stateRoot := m.getStateRoot()

	tx := &tx.Tx{
		TxHash:        common.Hash{}.String(), // FIXME: need change tx hash
		TxType:        int64(txType),
		GasFee:        txInfo.GasFeeAssetAmount.String(),
		GasFeeAssetId: txInfo.GasFeeAssetId,
		TxStatus:      mempool.SuccessTxStatus,
		BlockHeight:   state.blockNumber,
		StateRoot:     stateRoot,
		NftIndex:      commonConstant.NilTxNftIndex,
		PairIndex:     commonConstant.NilPairIndex,
		AssetId:       txInfo.AssetId,
		TxAmount:      txInfo.AssetAmount.String(),
		//TxInfo:        string(txInfoBytes),
		//TxDetails:     []*tx.TxDetail{txDetail}, // TODO: tx details.
		AccountIndex: txInfo.FromAccountIndex,
		Nonce:        txInfo.Nonce,
		ExpiredAt:    commonConstant.NilExpiredAt,
	}
	state.txs = append(state.txs, tx)
	state.stateRoot = stateRoot
	state.pendingUpdateAccountIndexMap[txInfo.FromAccountIndex] = true
	state.pendingUpdateAccountIndexMap[txInfo.ToAccountIndex] = true
	state.pendingUpdateAccountIndexMap[txInfo.GasAccountIndex] = true
	return nil
}

func (m *RestoreManager) replaySwap(state *BlockReplayState, pubData []byte) error {
	return nil
}

func (m *RestoreManager) replayAddLiquidity(state *BlockReplayState, pubData []byte) error {
	return nil
}

func (m *RestoreManager) replayRemoveLiquidity(state *BlockReplayState, pubData []byte) error {
	return nil
}

func (m *RestoreManager) replayWithdraw(state *BlockReplayState, pubData []byte) error {
	offset := 0
	offset, txType := util.ReadUint8(pubData, offset)
	offset, fromAccountIndex := util.ReadUint32(pubData, offset)
	offset, toAddress := util.ReadAddress(pubData, offset)
	offset, assetId := util.ReadUint16(pubData, offset)
	offset = 1 * ChunkBytesSize
	offset, packedAssetAmount := util.ReadUint40(pubData, offset)
	offset, gasAccountIndex := util.ReadUint32(pubData, offset)
	offset, gasFeeAssetId := util.ReadUint16(pubData, offset)
	offset, packedGasFeeAssetAmount := util.ReadUint16(pubData, offset)
	assetAmount, err := util.FromPackedAmount(packedAssetAmount)
	if err != nil {
		return fmt.Errorf("parse packed asset amount failed: %v", err)
	}
	gasFeeAssetAmount, err := util.FromPackedFee(int64(packedGasFeeAssetAmount))
	if err != nil {
		return fmt.Errorf("parse packed fee amount failed: %v", err)
	}
	txInfo := &WithdrawTxInfo{
		FromAccountIndex:  int64(fromAccountIndex),
		ToAddress:         toAddress,
		AssetId:           int64(assetId),
		AssetAmount:       assetAmount,
		GasAccountIndex:   int64(gasAccountIndex),
		GasFeeAssetId:     int64(gasFeeAssetId),
		GasFeeAssetAmount: gasFeeAssetAmount,
	}

	// Prepare account map and asset info.
	err = m.prepareAccountsAndAssets([]int64{txInfo.FromAccountIndex, txInfo.GasAccountIndex}, []int64{txInfo.AssetId, txInfo.GasFeeAssetId})
	if err != nil {
		return fmt.Errorf("prepare accounts failed: %v", err)
	}

	// Update asset info and asset tree.
	fromAccountInfo := m.accountMap[txInfo.FromAccountIndex]
	gasAccountInfo := m.accountMap[txInfo.GasAccountIndex]
	txInfo.Nonce = fromAccountInfo.Nonce
	fromAccountInfo.Nonce++
	fromAccountInfo.AssetInfo[txInfo.AssetId].Balance = new(big.Int).Sub(fromAccountInfo.AssetInfo[txInfo.AssetId].Balance, txInfo.AssetAmount)
	fromAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance = new(big.Int).Sub(fromAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	gasAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance = new(big.Int).Add(gasAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)

	// Update trees and compute state root.
	err = m.updateAccountTree([]int64{txInfo.FromAccountIndex, txInfo.GasAccountIndex}, []int64{txInfo.AssetId, txInfo.GasFeeAssetId})
	if err != nil {
		return fmt.Errorf("update account tree failed: %v", err)
	}
	stateRoot := m.getStateRoot()

	tx := &tx.Tx{
		TxHash:        common.Hash{}.String(), // FIXME: need change tx hash
		TxType:        int64(txType),
		GasFee:        txInfo.GasFeeAssetAmount.String(),
		GasFeeAssetId: txInfo.GasFeeAssetId,
		TxStatus:      mempool.SuccessTxStatus,
		BlockHeight:   state.blockNumber,
		StateRoot:     stateRoot,
		NftIndex:      commonConstant.NilTxNftIndex,
		PairIndex:     commonConstant.NilPairIndex,
		AssetId:       txInfo.AssetId,
		TxAmount:      txInfo.AssetAmount.String(),
		NativeAddress: txInfo.ToAddress,
		//TxInfo:        string(txInfoBytes),
		//TxDetails:     []*tx.TxDetail{txDetail}, // TODO: tx details.
		AccountIndex: txInfo.FromAccountIndex,
		Nonce:        txInfo.Nonce,
		ExpiredAt:    commonConstant.NilExpiredAt,
	}
	state.txs = append(state.txs, tx)
	state.stateRoot = stateRoot
	state.priorityOperations++
	state.pendingUpdateAccountIndexMap[txInfo.FromAccountIndex] = true
	state.pendingUpdateAccountIndexMap[txInfo.GasAccountIndex] = true
	state.pubDataOffset = append(state.pubDataOffset, uint32(len(state.pubDataOffset)))
	state.pendingOnChainOperationsPubData = append(state.pendingOnChainOperationsPubData, pubData)
	state.pendingOnChainOperationsHash = util.ConcatKeccakHash(state.pendingOnChainOperationsHash, pubData)
	return nil
}

func (m *RestoreManager) replayCreateCollection(state *BlockReplayState, pubData []byte) error {
	offset := 0
	offset, txType := util.ReadUint8(pubData, offset)
	offset, accountIndex := util.ReadUint32(pubData, offset)
	offset, collectionId := util.ReadUint16(pubData, offset)
	offset, gasAccountIndex := util.ReadUint32(pubData, offset)
	offset, gasFeeAssetId := util.ReadUint16(pubData, offset)
	offset, packedGasFeeAssetAmount := util.ReadUint16(pubData, offset)
	gasFeeAssetAmount, err := util.FromPackedFee(int64(packedGasFeeAssetAmount))
	if err != nil {
		return fmt.Errorf("parse packed fee amount failed: %v", err)
	}
	txInfo := &CreateCollectionTxInfo{
		AccountIndex:      int64(accountIndex),
		CollectionId:      int64(collectionId),
		GasAccountIndex:   int64(gasAccountIndex),
		GasFeeAssetId:     int64(gasFeeAssetId),
		GasFeeAssetAmount: gasFeeAssetAmount,
	}

	// Prepare account map and asset info.
	err = m.prepareAccountsAndAssets([]int64{txInfo.AccountIndex, txInfo.GasAccountIndex}, []int64{txInfo.GasFeeAssetId})
	if err != nil {
		return fmt.Errorf("prepare accounts failed: %v", err)
	}

	// Update asset info and asset tree.
	fromAccountInfo := m.accountMap[txInfo.AccountIndex]
	gasAccountInfo := m.accountMap[txInfo.GasAccountIndex]
	txInfo.Nonce = fromAccountInfo.Nonce
	fromAccountInfo.Nonce++
	fromAccountInfo.CollectionNonce = txInfo.CollectionId
	fromAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance = new(big.Int).Sub(fromAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	gasAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance = new(big.Int).Add(gasAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)

	// Update account tree.
	// Update trees and compute state root.
	err = m.updateAccountTree([]int64{txInfo.AccountIndex, txInfo.GasAccountIndex}, []int64{txInfo.GasFeeAssetId})
	if err != nil {
		return fmt.Errorf("update account tree failed: %v", err)
	}
	stateRoot := m.getStateRoot()

	tx := &tx.Tx{
		TxHash:        common.Hash{}.String(), // FIXME: need change tx hash
		TxType:        int64(txType),
		GasFee:        txInfo.GasFeeAssetAmount.String(),
		GasFeeAssetId: txInfo.GasFeeAssetId,
		TxStatus:      mempool.SuccessTxStatus,
		BlockHeight:   state.blockNumber,
		StateRoot:     stateRoot,
		NftIndex:      commonConstant.NilTxNftIndex,
		PairIndex:     commonConstant.NilPairIndex,
		NativeAddress: commonConstant.NilL1Address,
		//TxInfo:        string(txInfoBytes),
		//TxDetails:     []*tx.TxDetail{txDetail}, // TODO: tx details.
		AccountIndex: txInfo.AccountIndex,
		Nonce:        txInfo.Nonce,
		ExpiredAt:    commonConstant.NilExpiredAt,
	}
	state.txs = append(state.txs, tx)
	state.stateRoot = stateRoot
	state.pendingUpdateAccountIndexMap[txInfo.AccountIndex] = true
	state.pendingUpdateAccountIndexMap[txInfo.GasAccountIndex] = true
	return nil
}

func (m *RestoreManager) replayMintNft(state *BlockReplayState, pubData []byte) error {
	offset := 0
	offset, txType := util.ReadUint8(pubData, offset)
	offset, creatorAccountIndex := util.ReadUint32(pubData, offset)
	offset, toAccountIndex := util.ReadUint32(pubData, offset)
	offset, nftIndex := util.ReadUint40(pubData, offset)
	offset, gasAccountIndex := util.ReadUint32(pubData, offset)
	offset, gasFeeAssetId := util.ReadUint16(pubData, offset)
	offset, packedGasFeeAssetAmount := util.ReadUint16(pubData, offset)
	offset, creatorTreasuryRate := util.ReadUint16(pubData, offset)
	offset, nftCollectionId := util.ReadUint16(pubData, offset)
	offset = 1 * ChunkBytesSize
	offset, nftContentHash := util.ReadBytes32(pubData, offset)
	gasFeeAssetAmount, err := util.FromPackedFee(int64(packedGasFeeAssetAmount))
	if err != nil {
		return fmt.Errorf("parse packed fee amount failed: %v", err)
	}
	txInfo := &MintNftTxInfo{
		CreatorAccountIndex: int64(creatorAccountIndex),
		ToAccountIndex:      int64(toAccountIndex),
		NftIndex:            nftIndex,
		GasAccountIndex:     int64(gasAccountIndex),
		GasFeeAssetId:       int64(gasFeeAssetId),
		GasFeeAssetAmount:   gasFeeAssetAmount,
		CreatorTreasuryRate: int64(creatorTreasuryRate),
		NftCollectionId:     int64(nftCollectionId),
		NftContentHash:      common.Bytes2Hex(nftContentHash),
	}

	// Prepare account map and asset info.
	err = m.prepareAccountsAndAssets([]int64{txInfo.CreatorAccountIndex, txInfo.GasAccountIndex}, []int64{txInfo.GasFeeAssetId})
	if err != nil {
		return fmt.Errorf("prepare accounts failed: %v", err)
	}

	// Update asset info and asset tree.
	fromAccountInfo := m.accountMap[txInfo.CreatorAccountIndex]
	gasAccountInfo := m.accountMap[txInfo.GasAccountIndex]
	txInfo.Nonce = fromAccountInfo.Nonce
	fromAccountInfo.Nonce++
	fromAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance = new(big.Int).Sub(fromAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	gasAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance = new(big.Int).Add(gasAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	m.nftMap[txInfo.NftIndex] = &nft.L2Nft{
		NftIndex:            txInfo.NftIndex,
		CreatorAccountIndex: txInfo.CreatorAccountIndex,
		OwnerAccountIndex:   txInfo.ToAccountIndex,
		NftContentHash:      txInfo.NftContentHash,
		NftL1Address:        commonConstant.NilL1Address,
		NftL1TokenId:        commonConstant.NilL1TokenId,
		CreatorTreasuryRate: txInfo.CreatorTreasuryRate,
		CollectionId:        txInfo.NftCollectionId,
	}

	// Update trees and compute state root.
	err = m.updateAccountTree([]int64{txInfo.CreatorAccountIndex, txInfo.GasAccountIndex}, []int64{txInfo.GasFeeAssetId})
	if err != nil {
		return fmt.Errorf("update account tree failed: %v", err)
	}
	err = m.updateNftTree(txInfo.NftIndex)
	if err != nil {
		return fmt.Errorf("update nft tree failed: %v", err)
	}
	stateRoot := m.getStateRoot()

	tx := &tx.Tx{
		TxHash:        common.Hash{}.String(), // FIXME: need change tx hash
		TxType:        int64(txType),
		GasFee:        txInfo.GasFeeAssetAmount.String(),
		GasFeeAssetId: txInfo.GasFeeAssetId,
		TxStatus:      mempool.SuccessTxStatus,
		BlockHeight:   state.blockNumber,
		StateRoot:     stateRoot,
		NftIndex:      txInfo.NftIndex,
		PairIndex:     commonConstant.NilPairIndex,
		AssetId:       commonConstant.NilAssetId,
		TxAmount:      commonConstant.NilAssetAmountStr,
		NativeAddress: commonConstant.NilL1Address,
		//TxInfo:        string(txInfoBytes),
		//TxDetails:     []*tx.TxDetail{txDetail}, // TODO: tx details.
		AccountIndex: txInfo.CreatorAccountIndex,
		Nonce:        txInfo.Nonce,
		ExpiredAt:    commonConstant.NilExpiredAt,
	}
	state.txs = append(state.txs, tx)
	state.stateRoot = stateRoot
	state.pendingUpdateAccountIndexMap[txInfo.CreatorAccountIndex] = true
	state.pendingUpdateAccountIndexMap[txInfo.GasAccountIndex] = true
	state.pendingNewNftIndexMap[txInfo.NftIndex] = true
	return nil
}

func (m *RestoreManager) replayTransferNft(state *BlockReplayState, pubData []byte) error {
	offset := 0
	offset, txType := util.ReadUint8(pubData, offset)
	offset, fromAccountIndex := util.ReadUint32(pubData, offset)
	offset, toAccountIndex := util.ReadUint32(pubData, offset)
	offset, nftIndex := util.ReadUint40(pubData, offset)
	offset, gasAccountIndex := util.ReadUint32(pubData, offset)
	offset, gasFeeAssetId := util.ReadUint16(pubData, offset)
	offset, packedGasFeeAssetAmount := util.ReadUint16(pubData, offset)
	offset = 1 * ChunkBytesSize
	offset, callDataHash := util.ReadBytes32(pubData, offset)
	gasFeeAssetAmount, err := util.FromPackedFee(int64(packedGasFeeAssetAmount))
	if err != nil {
		return fmt.Errorf("parse packed fee amount failed: %v", err)
	}
	txInfo := &TransferNftTxInfo{
		FromAccountIndex:  int64(fromAccountIndex),
		ToAccountIndex:    int64(toAccountIndex),
		NftIndex:          nftIndex,
		GasAccountIndex:   int64(gasAccountIndex),
		GasFeeAssetId:     int64(gasFeeAssetId),
		GasFeeAssetAmount: gasFeeAssetAmount,
		CallDataHash:      callDataHash,
	}

	// Prepare account map and asset info.
	err = m.prepareAccountsAndAssets([]int64{txInfo.FromAccountIndex, txInfo.GasAccountIndex}, []int64{txInfo.GasFeeAssetId})
	if err != nil {
		return fmt.Errorf("prepare accounts failed: %v", err)
	}
	err = m.prepareNft(txInfo.NftIndex)
	if err != nil {
		return fmt.Errorf("prepare nft failed: %v", err)
	}

	// Update asset info and asset tree.
	fromAccountInfo := m.accountMap[txInfo.FromAccountIndex]
	gasAccountInfo := m.accountMap[txInfo.GasAccountIndex]
	txInfo.Nonce = fromAccountInfo.Nonce
	fromAccountInfo.Nonce++
	fromAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance = new(big.Int).Sub(fromAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	gasAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance = new(big.Int).Add(gasAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	m.nftMap[txInfo.NftIndex].OwnerAccountIndex = txInfo.ToAccountIndex

	// Update trees and compute state root.
	err = m.updateAccountTree([]int64{txInfo.FromAccountIndex, txInfo.GasAccountIndex}, []int64{txInfo.GasFeeAssetId})
	if err != nil {
		return fmt.Errorf("update account tree failed: %v", err)
	}
	err = m.updateNftTree(txInfo.NftIndex)
	if err != nil {
		return fmt.Errorf("update nft tree failed: %v", err)
	}
	stateRoot := m.getStateRoot()

	tx := &tx.Tx{
		TxHash:        common.Hash{}.String(), // FIXME: need change tx hash
		TxType:        int64(txType),
		GasFee:        txInfo.GasFeeAssetAmount.String(),
		GasFeeAssetId: txInfo.GasFeeAssetId,
		TxStatus:      mempool.SuccessTxStatus,
		BlockHeight:   state.blockNumber,
		StateRoot:     stateRoot,
		NftIndex:      txInfo.NftIndex,
		PairIndex:     commonConstant.NilPairIndex,
		AssetId:       commonConstant.NilAssetId,
		TxAmount:      commonConstant.NilAssetAmountStr,
		NativeAddress: "",
		//TxInfo:        string(txInfoBytes),
		//TxDetails:     []*tx.TxDetail{txDetail}, // TODO: tx details.
		AccountIndex: txInfo.FromAccountIndex,
		Nonce:        txInfo.Nonce,
		ExpiredAt:    commonConstant.NilExpiredAt,
	}
	state.txs = append(state.txs, tx)
	state.stateRoot = stateRoot
	state.pendingUpdateAccountIndexMap[txInfo.FromAccountIndex] = true
	state.pendingUpdateAccountIndexMap[txInfo.GasAccountIndex] = true
	state.pendingUpdateNftIndexMap[txInfo.NftIndex] = true
	return nil
}

func (m *RestoreManager) replayAtomicMatch(state *BlockReplayState, pubData []byte) error {
	offset := 0
	offset, txType := util.ReadUint8(pubData, offset)
	offset, accountIndex := util.ReadUint32(pubData, offset)
	offset, buyOfferAccountIndex := util.ReadUint32(pubData, offset)
	offset, buyOfferOfferId := util.ReadUint24(pubData, offset)
	offset, sellOfferAccountIndex := util.ReadUint32(pubData, offset)
	offset, sellOfferOfferId := util.ReadUint24(pubData, offset)
	offset, buyOfferNftIndex := util.ReadUint40(pubData, offset)
	offset, sellOfferAssetId := util.ReadUint16(pubData, offset)
	offset = 1 * ChunkBytesSize
	offset, buyOfferAssetAmount := util.ReadUint16(pubData, offset) // TODO: maybe wrong
	offset, creatorAmount := util.ReadUint16(pubData, offset)       // TODO: maybe wrong
	offset, treasuryAmount := util.ReadUint16(pubData, offset)      // TODO: maybe wrong
	offset, gasAccountIndex := util.ReadUint32(pubData, offset)
	offset, gasFeeAssetId := util.ReadUint16(pubData, offset)
	offset, gasFeeAssetAmount := util.ReadUint16(pubData, offset) // TODO: maybe wrong
	txInfo := &AtomicMatchTxInfo{
		AccountIndex: int64(accountIndex),
		BuyOffer: &OfferTxInfo{
			AccountIndex: int64(buyOfferAccountIndex),
			OfferId:      int64(buyOfferOfferId),
			NftIndex:     buyOfferNftIndex,
			AssetAmount:  big.NewInt(int64(buyOfferAssetAmount)),
		},
		SellOffer: &OfferTxInfo{
			AccountIndex: int64(sellOfferAccountIndex),
			OfferId:      int64(sellOfferOfferId),
			AssetId:      int64(sellOfferAssetId),
		},
		CreatorAmount:     big.NewInt(int64(creatorAmount)),
		TreasuryAmount:    big.NewInt(int64(treasuryAmount)),
		GasAccountIndex:   int64(gasAccountIndex),
		GasFeeAssetId:     int64(gasFeeAssetId),
		GasFeeAssetAmount: big.NewInt(int64(gasFeeAssetAmount)),
	}

	// Prepare account map and asset info.
	err := m.prepareNft(txInfo.BuyOffer.NftIndex)
	if err != nil {
		return fmt.Errorf("prepare nft failed: %v", err)
	}
	nftAsset := m.nftMap[txInfo.BuyOffer.NftIndex]
	offerAssetId := txInfo.BuyOffer.OfferId / OfferPerAsset
	err = m.prepareAccountsAndAssets([]int64{txInfo.AccountIndex, txInfo.BuyOffer.AccountIndex, txInfo.SellOffer.AccountIndex, nftAsset.CreatorAccountIndex, txInfo.GasAccountIndex},
		[]int64{offerAssetId, txInfo.SellOffer.AssetId, txInfo.GasFeeAssetId})
	if err != nil {
		return fmt.Errorf("prepare accounts failed: %v", err)
	}

	// Update asset info and asset tree.
	fromAccountInfo := m.accountMap[txInfo.AccountIndex]
	gasAccountInfo := m.accountMap[txInfo.GasAccountIndex]
	buyerAccountInfo := m.accountMap[txInfo.BuyOffer.AccountIndex]
	sellerAccountInfo := m.accountMap[txInfo.SellOffer.AccountIndex]
	creatorAccountInfo := m.accountMap[nftAsset.CreatorAccountIndex]
	txInfo.Nonce = fromAccountInfo.Nonce
	fromAccountInfo.Nonce++
	fromAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance = new(big.Int).Sub(fromAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	gasAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance = new(big.Int).Add(gasAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	// Update buyer.
	buyerAccountInfo.AssetInfo[txInfo.SellOffer.AssetId].Balance = new(big.Int).Sub(buyerAccountInfo.AssetInfo[txInfo.SellOffer.AssetId].Balance, txInfo.BuyOffer.AssetAmount)
	oOffer := buyerAccountInfo.AssetInfo[offerAssetId].OfferCanceledOrFinalized
	offerIndex := txInfo.BuyOffer.OfferId % OfferPerAsset
	nOffer := new(big.Int).SetBit(oOffer, int(offerIndex), 1)
	buyerAccountInfo.AssetInfo[offerAssetId].OfferCanceledOrFinalized = nOffer
	// Update seller.
	sellerIncBalance := ffmath.Sub(txInfo.BuyOffer.AssetAmount, ffmath.Add(txInfo.TreasuryAmount, txInfo.CreatorAmount))
	sellerAccountInfo.AssetInfo[txInfo.SellOffer.AssetId].Balance = ffmath.Add(sellerAccountInfo.AssetInfo[txInfo.SellOffer.AssetId].Balance, sellerIncBalance)
	oOffer = sellerAccountInfo.AssetInfo[offerAssetId].OfferCanceledOrFinalized
	nOffer = new(big.Int).SetBit(oOffer, int(offerIndex), 1)
	sellerAccountInfo.AssetInfo[offerAssetId].OfferCanceledOrFinalized = nOffer
	// Update creator.
	creatorAccountInfo.AssetInfo[txInfo.SellOffer.AssetId].Balance = ffmath.Add(creatorAccountInfo.AssetInfo[txInfo.SellOffer.AssetId].Balance, txInfo.CreatorAmount)
	// Update treasury.
	gasAccountInfo.AssetInfo[txInfo.SellOffer.AssetId].Balance = ffmath.Add(gasAccountInfo.AssetInfo[txInfo.SellOffer.AssetId].Balance, txInfo.TreasuryAmount)
	// TODO: update nft info.

	// Update trees and compute state root.
	err = m.updateAccountTree([]int64{txInfo.AccountIndex, txInfo.BuyOffer.AccountIndex, txInfo.SellOffer.AccountIndex, nftAsset.CreatorAccountIndex, txInfo.GasAccountIndex},
		[]int64{offerAssetId, txInfo.SellOffer.AssetId, txInfo.GasFeeAssetId})
	if err != nil {
		return fmt.Errorf("update account tree failed: %v", err)
	}
	err = m.updateNftTree(txInfo.BuyOffer.NftIndex)
	if err != nil {
		return fmt.Errorf("update nft tree failed: %v", err)
	}
	stateRoot := m.getStateRoot()

	tx := &tx.Tx{
		TxHash:        common.Hash{}.String(), // FIXME: need change tx hash
		TxType:        int64(txType),
		GasFee:        txInfo.GasFeeAssetAmount.String(),
		GasFeeAssetId: txInfo.GasFeeAssetId,
		TxStatus:      mempool.SuccessTxStatus,
		BlockHeight:   state.blockNumber,
		StateRoot:     stateRoot,
		NftIndex:      txInfo.BuyOffer.NftIndex,
		PairIndex:     commonConstant.NilPairIndex,
		AssetId:       txInfo.SellOffer.AssetId,
		TxAmount:      txInfo.BuyOffer.AssetAmount.String(),
		NativeAddress: commonConstant.NilL1Address,
		//TxInfo:        string(txInfoBytes),
		//TxDetails:     []*tx.TxDetail{txDetail}, // TODO: tx details.
		AccountIndex: txInfo.AccountIndex,
		Nonce:        txInfo.Nonce,
		ExpiredAt:    commonConstant.NilExpiredAt,
	}
	state.txs = append(state.txs, tx)
	state.stateRoot = stateRoot
	state.pendingUpdateAccountIndexMap[txInfo.AccountIndex] = true
	state.pendingUpdateAccountIndexMap[txInfo.BuyOffer.AccountIndex] = true
	state.pendingUpdateAccountIndexMap[txInfo.SellOffer.AccountIndex] = true
	state.pendingUpdateAccountIndexMap[nftAsset.CreatorAccountIndex] = true
	state.pendingUpdateAccountIndexMap[txInfo.GasAccountIndex] = true
	state.pendingUpdateNftIndexMap[txInfo.BuyOffer.NftIndex] = true
	return nil
}

func (m *RestoreManager) replayCancelOffer(state *BlockReplayState, pubData []byte) error {
	offset := 0
	offset, txType := util.ReadUint8(pubData, offset)
	offset, accountIndex := util.ReadUint32(pubData, offset)
	offset, offerId := util.ReadUint24(pubData, offset)
	offset, gasAccountIndex := util.ReadUint32(pubData, offset)
	offset, gasFeeAssetId := util.ReadUint16(pubData, offset)
	offset, packedGasFeeAssetAmount := util.ReadUint16(pubData, offset) // TODO: maybe wrong
	gasFeeAssetAmount, err := util.FromPackedFee(int64(packedGasFeeAssetAmount))
	if err != nil {
		return fmt.Errorf("parse packed fee amount failed: %v", err)
	}
	txInfo := &CancelOfferTxInfo{
		AccountIndex:      int64(accountIndex),
		OfferId:           int64(offerId),
		GasAccountIndex:   int64(gasAccountIndex),
		GasFeeAssetId:     int64(gasFeeAssetId),
		GasFeeAssetAmount: gasFeeAssetAmount,
	}

	// Prepare account map and asset info.
	offerAssetId := txInfo.OfferId / OfferPerAsset
	offerIndex := txInfo.OfferId % OfferPerAsset
	// Prepare account map and asset info.
	err = m.prepareAccountsAndAssets([]int64{txInfo.AccountIndex, txInfo.GasAccountIndex}, []int64{offerAssetId, txInfo.GasFeeAssetId})
	if err != nil {
		return fmt.Errorf("prepare accounts failed: %v", err)
	}

	// Update asset info and asset tree.
	fromAccountInfo := m.accountMap[txInfo.AccountIndex]
	gasAccountInfo := m.accountMap[txInfo.GasAccountIndex]
	txInfo.Nonce = fromAccountInfo.Nonce
	fromAccountInfo.Nonce++
	fromAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance = new(big.Int).Sub(fromAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	gasAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance = new(big.Int).Add(gasAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	oOffer := fromAccountInfo.AssetInfo[offerAssetId].OfferCanceledOrFinalized
	nOffer := new(big.Int).SetBit(oOffer, int(offerIndex), 1)
	fromAccountInfo.AssetInfo[offerAssetId].OfferCanceledOrFinalized = nOffer

	// Update trees and compute state root.
	err = m.updateAccountTree([]int64{txInfo.AccountIndex, txInfo.GasAccountIndex}, []int64{offerAssetId, txInfo.GasFeeAssetId})
	if err != nil {
		return fmt.Errorf("update account tree failed: %v", err)
	}
	stateRoot := m.getStateRoot()

	tx := &tx.Tx{
		TxHash:        common.Hash{}.String(), // FIXME: need change tx hash
		TxType:        int64(txType),
		GasFee:        txInfo.GasFeeAssetAmount.String(),
		GasFeeAssetId: txInfo.GasFeeAssetId,
		TxStatus:      mempool.SuccessTxStatus,
		BlockHeight:   state.blockNumber,
		StateRoot:     stateRoot,
		NftIndex:      commonConstant.NilTxNftIndex,
		PairIndex:     commonConstant.NilPairIndex,
		AssetId:       commonConstant.NilAssetId,
		TxAmount:      commonConstant.NilAssetAmountStr,
		NativeAddress: commonConstant.NilL1Address,
		//TxInfo:        string(txInfoBytes),
		//TxDetails:     []*tx.TxDetail{txDetail}, // TODO: tx details.
		AccountIndex: txInfo.AccountIndex,
		Nonce:        txInfo.Nonce,
		ExpiredAt:    commonConstant.NilExpiredAt,
	}
	state.txs = append(state.txs, tx)
	state.stateRoot = stateRoot
	state.pendingUpdateAccountIndexMap[txInfo.AccountIndex] = true
	state.pendingUpdateAccountIndexMap[txInfo.GasAccountIndex] = true
	return nil
}

func (m *RestoreManager) replayWithdrawNft(state *BlockReplayState, pubData []byte) error {
	offset := 0
	offset, txType := util.ReadUint8(pubData, offset)
	offset, accountIndex := util.ReadUint32(pubData, offset)
	offset, creatorAccountIndex := util.ReadUint32(pubData, offset)
	offset, creatorTreasuryRate := util.ReadUint16(pubData, offset)
	offset, nftIndex := util.ReadUint40(pubData, offset)
	offset, collectionId := util.ReadUint16(pubData, offset)
	offset = 1 * ChunkBytesSize
	offset, nftL1Address := util.ReadAddress(pubData, offset)
	offset = 2 * ChunkBytesSize
	offset, toAddress := util.ReadAddress(pubData, offset)
	offset, gasAccountIndex := util.ReadUint32(pubData, offset)
	offset, gasFeeAssetId := util.ReadUint16(pubData, offset)
	offset, packedGasFeeAssetAmount := util.ReadUint16(pubData, offset)
	offset = 3 * ChunkBytesSize
	offset, nftContentHash := util.ReadBytes32(pubData, offset)
	offset, nftL1TokenId := util.ReadUint256(pubData, offset)
	offset, creatorAccountNameHash := util.ReadBytes32(pubData, offset)
	gasFeeAssetAmount, err := util.FromPackedFee(int64(packedGasFeeAssetAmount))
	if err != nil {
		return fmt.Errorf("parse packed fee amount failed: %v", err)
	}
	txInfo := &WithdrawNftTxInfo{
		AccountIndex:           int64(accountIndex),
		CreatorAccountIndex:    int64(creatorAccountIndex),
		CreatorTreasuryRate:    int64(creatorTreasuryRate),
		NftIndex:               nftIndex,
		CollectionId:           int64(collectionId),
		NftL1Address:           nftL1Address,
		ToAddress:              toAddress,
		GasAccountIndex:        int64(gasAccountIndex),
		GasFeeAssetId:          int64(gasFeeAssetId),
		GasFeeAssetAmount:      gasFeeAssetAmount,
		NftContentHash:         nftContentHash,
		NftL1TokenId:           nftL1TokenId,
		CreatorAccountNameHash: creatorAccountNameHash,
	}

	// Prepare account map and asset info.
	err = m.prepareAccountsAndAssets([]int64{txInfo.AccountIndex, txInfo.GasAccountIndex}, []int64{txInfo.GasFeeAssetId})
	if err != nil {
		return fmt.Errorf("prepare accounts failed: %v", err)
	}
	err = m.prepareNft(txInfo.NftIndex)
	if err != nil {
		return fmt.Errorf("prepare nft failed: %v", err)
	}

	// Update asset info and asset tree.
	fromAccountInfo := m.accountMap[txInfo.AccountIndex]
	gasAccountInfo := m.accountMap[txInfo.GasAccountIndex]
	txInfo.Nonce = fromAccountInfo.Nonce
	fromAccountInfo.Nonce++
	fromAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance = new(big.Int).Sub(fromAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	gasAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance = new(big.Int).Add(gasAccountInfo.AssetInfo[txInfo.GasFeeAssetId].Balance, txInfo.GasFeeAssetAmount)
	nftAsset := m.nftMap[txInfo.NftIndex]
	state.pendingNewNftWithdrawHistory = append(state.pendingNewNftWithdrawHistory, &nft.L2NftWithdrawHistory{
		NftIndex:            nftAsset.NftIndex,
		CreatorAccountIndex: nftAsset.CreatorAccountIndex,
		OwnerAccountIndex:   nftAsset.OwnerAccountIndex,
		NftContentHash:      nftAsset.NftContentHash,
		NftL1Address:        nftAsset.NftL1Address,
		NftL1TokenId:        nftAsset.NftL1TokenId,
		CreatorTreasuryRate: nftAsset.CreatorTreasuryRate,
		CollectionId:        nftAsset.CollectionId,
	})
	newNftInfo := commonAsset.EmptyNftInfo(txInfo.NftIndex)
	m.nftMap[txInfo.NftIndex] = &nft.L2Nft{
		Model:               nftAsset.Model,
		NftIndex:            newNftInfo.NftIndex,
		CreatorAccountIndex: newNftInfo.CreatorAccountIndex,
		OwnerAccountIndex:   newNftInfo.OwnerAccountIndex,
		NftContentHash:      newNftInfo.NftContentHash,
		NftL1Address:        newNftInfo.NftL1Address,
		NftL1TokenId:        newNftInfo.NftL1TokenId,
		CreatorTreasuryRate: newNftInfo.CreatorTreasuryRate,
		CollectionId:        newNftInfo.CollectionId,
	}

	// Update trees and compute state root.
	err = m.updateAccountTree([]int64{txInfo.AccountIndex, txInfo.GasAccountIndex}, []int64{txInfo.GasFeeAssetId})
	if err != nil {
		return fmt.Errorf("update account tree failed: %v", err)
	}
	err = m.updateNftTree(txInfo.NftIndex)
	if err != nil {
		return fmt.Errorf("update nft tree failed: %v", err)
	}
	stateRoot := m.getStateRoot()

	tx := &tx.Tx{
		TxHash:        common.Hash{}.String(), // FIXME: need change tx hash
		TxType:        int64(txType),
		GasFee:        txInfo.GasFeeAssetAmount.String(),
		GasFeeAssetId: txInfo.GasFeeAssetId,
		TxStatus:      mempool.SuccessTxStatus,
		BlockHeight:   state.blockNumber,
		StateRoot:     stateRoot,
		NftIndex:      txInfo.NftIndex,
		PairIndex:     commonConstant.NilPairIndex,
		AssetId:       commonConstant.NilAssetId,
		TxAmount:      commonConstant.NilAssetAmountStr,
		NativeAddress: "",
		//TxInfo:        string(txInfoBytes),
		//TxDetails:     []*tx.TxDetail{txDetail}, // TODO: tx details.
		AccountIndex: txInfo.AccountIndex,
		Nonce:        txInfo.Nonce,
		ExpiredAt:    commonConstant.NilExpiredAt,
	}
	state.txs = append(state.txs, tx)
	state.stateRoot = stateRoot
	state.pendingUpdateAccountIndexMap[txInfo.AccountIndex] = true
	state.pendingUpdateAccountIndexMap[txInfo.GasAccountIndex] = true
	state.pendingUpdateNftIndexMap[txInfo.NftIndex] = true
	state.pubDataOffset = append(state.pubDataOffset, uint32(len(state.pubDataOffset)))
	state.pendingOnChainOperationsPubData = append(state.pendingOnChainOperationsPubData, pubData)
	state.pendingOnChainOperationsHash = util.ConcatKeccakHash(state.pendingOnChainOperationsHash, pubData)
	return nil
}

func (m *RestoreManager) replayFullExit(state *BlockReplayState, pubData []byte) error {
	offset := 0
	offset, txType := util.ReadUint8(pubData, offset)
	offset, accountIndex := util.ReadUint32(pubData, offset)
	offset, assetId := util.ReadUint16(pubData, offset)
	offset, assetAmount := util.ReadUint128(pubData, offset)
	offset = 1 * ChunkBytesSize
	offset, accountNameHash := util.ReadBytes32(pubData, offset)
	txInfo := &FullExitTxInfo{
		TxType:          txType,
		AccountIndex:    int64(accountIndex),
		AccountNameHash: accountNameHash,
		AssetId:         int64(assetId),
		AssetAmount:     assetAmount,
	}
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return fmt.Errorf("unable to serialize tx info : %v", err)
	}

	// Prepare account map and asset info.
	err = m.prepareAccountsAndAssets([]int64{txInfo.AccountIndex}, []int64{txInfo.AssetId})
	if err != nil {
		return fmt.Errorf("prepare accounts failed: %v", err)
	}

	// Update asset balance.
	accountInfo := m.accountMap[txInfo.AccountIndex]
	if (accountInfo.AssetInfo == nil || accountInfo.AssetInfo[txInfo.AssetId] == nil) && txInfo.AssetAmount.Int64() != 0 {
		return fmt.Errorf("unexpect asset amount, expect: 0, real: %d", txInfo.AssetAmount.Int64())
	}
	if accountInfo.AssetInfo != nil && accountInfo.AssetInfo[txInfo.AssetId] != nil &&
		accountInfo.AssetInfo[txInfo.AssetId].Balance.Cmp(txInfo.AssetAmount) != 0 {
		return fmt.Errorf("unexpect asset amount, expect: %d, real: %d",
			accountInfo.AssetInfo[txInfo.AssetId].Balance.Int64(), txInfo.AssetAmount.Int64())
	}
	accountInfo.AssetInfo[txInfo.AssetId].Balance = ZeroBigInt

	// Update trees and compute state root.
	err = m.updateAccountTree([]int64{txInfo.AccountIndex}, []int64{txInfo.AssetId})
	if err != nil {
		return fmt.Errorf("update account tree failed: %v", err)
	}
	stateRoot := m.getStateRoot()

	balanceDelta := &commonAsset.AccountAsset{
		AssetId:                  txInfo.AssetId,
		Balance:                  ffmath.Neg(txInfo.AssetAmount),
		LpAmount:                 big.NewInt(0),
		OfferCanceledOrFinalized: big.NewInt(0),
	}
	txDetail := &tx.TxDetail{
		AssetId:      txInfo.AssetId,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.AccountIndex,
		AccountName:  accountInfo.AccountName,
		BalanceDelta: balanceDelta.String(),
		Order:        0,
		AccountOrder: 0,
	}
	tx := &tx.Tx{
		TxHash:        common.Hash{}.String(), // FIXME: need change tx hash
		TxType:        int64(txInfo.TxType),
		GasFee:        commonConstant.NilAssetAmountStr,
		GasFeeAssetId: commonConstant.NilAssetId,
		TxStatus:      mempool.SuccessTxStatus,
		BlockHeight:   state.blockNumber,
		StateRoot:     stateRoot,
		NftIndex:      commonConstant.NilTxNftIndex,
		PairIndex:     commonConstant.NilPairIndex,
		AssetId:       txInfo.AssetId,
		TxAmount:      txInfo.AssetAmount.String(),
		NativeAddress: m.accountMap[txInfo.AccountIndex].L1Address,
		TxInfo:        string(txInfoBytes),
		TxDetails:     []*tx.TxDetail{txDetail},
		AccountIndex:  txInfo.AccountIndex,
		Nonce:         commonConstant.NilNonce,
		ExpiredAt:     commonConstant.NilExpiredAt,
	}
	state.txs = append(state.txs, tx)
	state.stateRoot = stateRoot
	state.pendingUpdateAccountIndexMap[txInfo.AccountIndex] = true
	state.priorityOperations++
	state.pubDataOffset = append(state.pubDataOffset, uint32(len(state.pubDataOffset)))
	state.pendingOnChainOperationsPubData = append(state.pendingOnChainOperationsPubData, pubData)
	state.pendingOnChainOperationsHash = util.ConcatKeccakHash(state.pendingOnChainOperationsHash, pubData)
	return nil
}

func (m *RestoreManager) replayFullExitNft(state *BlockReplayState, pubData []byte) error {
	offset := 0
	offset, txType := util.ReadUint8(pubData, offset)
	offset, accountIndex := util.ReadUint32(pubData, offset)
	offset, creatorAccountIndex := util.ReadUint32(pubData, offset)
	offset, creatorTreasuryRate := util.ReadUint16(pubData, offset)
	offset, nftIndex := util.ReadUint40(pubData, offset)
	offset, collectionId := util.ReadUint16(pubData, offset)
	offset = 1 * ChunkBytesSize
	offset, nftL1Address := util.ReadAddress(pubData, offset) // TODO: Prefix padding, maybe wrong
	offset = 2 * ChunkBytesSize
	offset, accountNameHash := util.ReadBytes32(pubData, offset)
	offset, creatorAccountNameHash := util.ReadBytes32(pubData, offset)
	offset, nftContentHash := util.ReadBytes32(pubData, offset)
	offset, nftL1TokenId := util.ReadUint256(pubData, offset)
	txInfo := &FullExitNftTxInfo{
		TxType:                 txType,
		AccountIndex:           int64(accountIndex),
		CreatorAccountIndex:    int64(creatorAccountIndex),
		CreatorTreasuryRate:    int64(creatorTreasuryRate),
		NftIndex:               nftIndex,
		CollectionId:           int64(collectionId),
		NftL1Address:           nftL1Address,
		AccountNameHash:        accountNameHash,
		CreatorAccountNameHash: creatorAccountNameHash,
		NftContentHash:         nftContentHash,
		NftL1TokenId:           nftL1TokenId,
	}
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		return fmt.Errorf("unable to serialize tx info : %v", err)
	}

	// Prepare account map and asset info.
	err = m.prepareAccountsAndAssets([]int64{txInfo.AccountIndex}, nil)
	if err != nil {
		return fmt.Errorf("prepare accounts failed: %v", err)
	}
	err = m.prepareNft(txInfo.NftIndex)
	if err != nil {
		return fmt.Errorf("prepare nft failed: %v", err)
	}

	nftAsset := m.nftMap[txInfo.NftIndex]
	if nftAsset.OwnerAccountIndex != txInfo.AccountIndex {
		return fmt.Errorf("owner mismatch, expect: %d, real: %d", nftAsset.OwnerAccountIndex, txInfo.AccountIndex)
	}
	state.pendingNewNftWithdrawHistory = append(state.pendingNewNftWithdrawHistory, &nft.L2NftWithdrawHistory{
		NftIndex:            nftAsset.NftIndex,
		CreatorAccountIndex: nftAsset.CreatorAccountIndex,
		OwnerAccountIndex:   nftAsset.OwnerAccountIndex,
		NftContentHash:      nftAsset.NftContentHash,
		NftL1Address:        nftAsset.NftL1Address,
		NftL1TokenId:        nftAsset.NftL1TokenId,
		CreatorTreasuryRate: nftAsset.CreatorTreasuryRate,
		CollectionId:        nftAsset.CollectionId,
	})
	nftInfo := commonAsset.EmptyNftInfo(txInfo.NftIndex)
	m.nftMap[txInfo.NftIndex] = &nft.L2Nft{
		Model:               m.nftMap[txInfo.NftIndex].Model,
		NftIndex:            nftInfo.NftIndex,
		CreatorAccountIndex: nftInfo.CreatorAccountIndex,
		OwnerAccountIndex:   nftInfo.OwnerAccountIndex,
		NftContentHash:      nftInfo.NftContentHash,
		NftL1Address:        nftInfo.NftL1Address,
		NftL1TokenId:        nftInfo.NftL1TokenId,
		CreatorTreasuryRate: nftInfo.CreatorTreasuryRate,
		CollectionId:        nftInfo.CollectionId,
	}

	// Update trees and compute state root.
	err = m.updateNftTree(txInfo.NftIndex)
	if err != nil {
		return fmt.Errorf("update nft tree failed: %v", err)
	}
	stateRoot := m.getStateRoot()

	// Construct tx.
	var txDetails []*tx.TxDetail
	accountInfo := m.accountMap[txInfo.AccountIndex]
	emptyDeltaAsset := &commonAsset.AccountAsset{
		AssetId:                  0,
		Balance:                  big.NewInt(0),
		LpAmount:                 big.NewInt(0),
		OfferCanceledOrFinalized: big.NewInt(0),
	}
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      0,
		AssetType:    commonAsset.GeneralAssetType,
		AccountIndex: txInfo.AccountIndex,
		AccountName:  accountInfo.AccountName,
		BalanceDelta: emptyDeltaAsset.String(),
		Order:        0,
		AccountOrder: 0,
	})
	txDetails = append(txDetails, &tx.TxDetail{
		AssetId:      txInfo.NftIndex,
		AssetType:    commonAsset.NftAssetType,
		AccountIndex: txInfo.AccountIndex,
		AccountName:  accountInfo.AccountName,
		BalanceDelta: nftInfo.String(),
		Order:        1,
		AccountOrder: commonConstant.NilAccountOrder,
	})
	tx := &tx.Tx{
		TxHash:        common.Hash{}.String(), // FIXME: need change tx hash
		TxType:        int64(txInfo.TxType),
		GasFee:        commonConstant.NilAssetAmountStr,
		GasFeeAssetId: commonConstant.NilAssetId,
		TxStatus:      mempool.SuccessTxStatus,
		BlockHeight:   state.blockNumber,
		StateRoot:     state.stateRoot,
		NftIndex:      txInfo.NftIndex,
		PairIndex:     commonConstant.NilPairIndex,
		AssetId:       commonConstant.NilAssetId,
		TxAmount:      commonConstant.NilAssetAmountStr,
		NativeAddress: m.accountMap[txInfo.AccountIndex].L1Address,
		TxInfo:        string(txInfoBytes),
		TxDetails:     txDetails,
		AccountIndex:  txInfo.AccountIndex,
		Nonce:         commonConstant.NilNonce,
		ExpiredAt:     commonConstant.NilExpiredAt,
	}
	state.txs = append(state.txs, tx)
	state.stateRoot = stateRoot
	state.pendingUpdateAccountIndexMap[txInfo.AccountIndex] = true
	state.pendingUpdateNftIndexMap[txInfo.NftIndex] = true
	state.priorityOperations++
	state.pubDataOffset = append(state.pubDataOffset, uint32(len(state.pubDataOffset)))
	state.pendingOnChainOperationsPubData = append(state.pendingOnChainOperationsPubData, pubData)
	state.pendingOnChainOperationsHash = util.ConcatKeccakHash(state.pendingOnChainOperationsHash, pubData)
	return nil
}

func (m *RestoreManager) handleBlockCommitTopic(log types.Log) ([]zkbas.OldZkbasCommitBlockInfo, error) {
	l1Tx, _, err := m.l1Provider.TransactionByHash(context.Background(), log.TxHash)
	if err != nil {
		return nil, fmt.Errorf("get layer1 tx failed: %v, tx hash: %s", err, log.TxHash)
	}

	if len(l1Tx.Data()) < 4 {
		return nil, fmt.Errorf("layer1 tx data length is too small, tx hash: %s", l1Tx.Hash().String())
	}
	method, err := m.l1ContractABI.MethodById(l1Tx.Data()[:4])
	if err != nil {
		return nil, fmt.Errorf("layer1 tx get method failed, tx hash: %s", l1Tx.Hash().String())
	}
	inputs, err := method.Inputs.Unpack(l1Tx.Data()[4:])
	if err != nil {
		return nil, fmt.Errorf("layer1 tx unpack inputs failed, tx hash: %s", l1Tx.Hash().String())
	}
	if len(inputs) < 2 {
		return nil, fmt.Errorf("layer1 tx unexpected inputs length, tx hash: %s, expect: 2, real: %d", l1Tx.Hash().String(), len(inputs))
	}

	var commitBlocks []zkbas.OldZkbasCommitBlockInfo
	abi.ConvertType(inputs[1], &commitBlocks)
	if len(commitBlocks) == 0 {
		return nil, fmt.Errorf("convert block data failed")
	}
	return commitBlocks, nil
}

func (m *RestoreManager) handleBlockVerifyTopic(log types.Log) (uint32, error) {
	event, err := m.l1ContractABI.EventByID(BlockVerifyTopic)
	if err != nil {
		return 0, fmt.Errorf("layer1 event get event id failed: %v", err)
	}
	inputs, err := event.Inputs.Unpack(log.Data)
	if err != nil {
		return 0, fmt.Errorf("layer1 event input unpack failed: %v", err)
	}
	if len(inputs) < 1 {
		return 0, fmt.Errorf("layer1 event unexpected input length, expect: 1, real: 0")
	}
	var l2BlockNumber uint32
	abi.ConvertType(inputs[0], &l2BlockNumber)
	if l2BlockNumber == 0 {
		return 0, fmt.Errorf("unexpect layer2 block number: 0")
	}

	return l2BlockNumber, nil
}

func (m *RestoreManager) handleBlockRevertTopic(log types.Log) {

}

func (m *RestoreManager) RestoreHistoryData(startL1Number, endL1Number *big.Int) error {
	// Handle error inputs.
	if endL1Number == nil {
		return fmt.Errorf("nil end layer1 block number")
	}
	if startL1Number == nil {
		startL1Number = m.l1genesisNumber
	}
	if startL1Number.Cmp(endL1Number) > 0 {
		return fmt.Errorf("layer1 block number end is little than start")
	}

	topics := []common.Hash{BlockCommitTopic, BlockVerifyTopic, BlockRevertTopic}
	step := big.NewInt(Layer1FilterStep)
	lastL1Number := startL1Number
	for lastL1Number.Cmp(endL1Number) <= 0 {
		toL1Number := new(big.Int).Add(lastL1Number, step)
		if toL1Number.Cmp(endL1Number) > 0 {
			toL1Number = endL1Number
		}
		query := ethereum.FilterQuery{
			FromBlock: lastL1Number,
			ToBlock:   toL1Number,
			Addresses: []common.Address{m.l1Contract},
			Topics:    [][]common.Hash{topics},
		}

		logs, err := m.l1Provider.FilterLogs(context.Background(), query)
		if err != nil {
			return fmt.Errorf("filter logs failed: %v", err)
		}

		lastL1Number = new(big.Int).Add(toL1Number, big.NewInt(1))
		lastL1TxHash := common.Hash{}
		for _, log := range logs {
			switch log.Topics[0] {
			case BlockCommitTopic:
				// One layer1 transaction may include several layer2 blocks, skip duplicated.
				if log.TxHash == lastL1TxHash {
					continue
				}

				lastL1TxHash = log.TxHash
				commitBlocks, err := m.handleBlockCommitTopic(log)
				if err != nil {
					return err
				}
				for _, commitData := range commitBlocks {
					m.commitCh <- &commitData
				}

			case BlockVerifyTopic:
				_, err := m.handleBlockVerifyTopic(log)
				if err != nil {
					return err
				}

			case BlockRevertTopic:
				m.handleBlockRevertTopic(log)
			}
		}
	}

	return nil
}

func (m *RestoreManager) SyncVerifiedData() error {
	return nil
}

func (m *RestoreManager) SyncCommittedData() error {
	return nil
}
