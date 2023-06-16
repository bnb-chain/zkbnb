package desertexit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/bnb-chain/zkbnb-crypto/circuit"
	"github.com/bnb-chain/zkbnb-crypto/circuit/desert"
	desertTypes "github.com/bnb-chain/zkbnb-crypto/circuit/desert/types"
	cryptoTypes "github.com/bnb-chain/zkbnb-crypto/circuit/types"
	bsmt "github.com/bnb-chain/zkbnb-smt"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/common/chain"
	"github.com/bnb-chain/zkbnb/common/prove"
	"github.com/bnb-chain/zkbnb/tree"
	"github.com/bnb-chain/zkbnb/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-redis/redis/v8"
	"github.com/zeromicro/go-zero/core/logx"
)

func (c *GenerateProof) generateWitness(blockHeight int64, accountIndex int64, nftIndex int64, assetTokenAddress string, txType uint8) (
	desertInfo *desert.Desert, pubData []byte, err error,
) {
	var (
		accountsInfo [desertTypes.NbAccountsPerTx]*desertTypes.Account

		// account asset merkle proof
		merkleProofsAccountAssets [desertTypes.NbAccountsPerTx][circuit.AssetMerkleLevels][]byte
		// account merkle proof
		merkleProofsAccounts [desertTypes.NbAccountsPerTx][circuit.AccountMerkleLevels][]byte
		// nft
		nft *cryptoTypes.Nft
		// nft tree merkle proof
		merkleProofsNft [circuit.NftMerkleLevels][]byte

		exitTxInfo    *desert.ExitTx
		exitNftTxInfo *desert.ExitNftTx
	)

	accountTree, accountAssetTrees, nftTree, err := c.initSmtTree(blockHeight)
	if err != nil {
		return desertInfo, pubData, err
	}

	accountInfo, err := c.bc.DB().AccountModel.GetAccountByIndex(accountIndex)
	if err != nil {
		logx.Errorf("get account failed: %s", err)
		return desertInfo, pubData, err
	}
	formatAccountInfo, err := chain.ToFormatAccountInfo(accountInfo)
	if err != nil {
		return desertInfo, pubData, err
	}
	pk, err := common2.ParsePubKey(formatAccountInfo.PublicKey)
	if err != nil {
		return desertInfo, pubData, err
	}
	accountsInfo[0] = &desertTypes.Account{
		AccountIndex:    accountIndex,
		L1Address:       common2.AddressStrToBytes(formatAccountInfo.L1Address),
		AccountPk:       pk,
		Nonce:           formatAccountInfo.Nonce,
		CollectionNonce: formatAccountInfo.CollectionNonce,
		AssetRoot:       accountAssetTrees.Get(accountIndex).Root(),
	}

	// get account before
	accountMerkleProofs, err := accountTree.GetProof(uint64(accountIndex))
	if err != nil {
		return desertInfo, pubData, err
	}

	// set account merkle proof
	merkleProofsAccounts[0], err = prove.SetFixedAccountArray(accountMerkleProofs)
	if err != nil {
		return desertInfo, pubData, err
	}

	if err != nil {
		logx.Errorf("get stored block info: %s", err.Error())
		return desertInfo, pubData, err
	}

	if txType == desertTypes.TxTypeExit {
		monitor, err := NewDesertExit(c.config)
		if err != nil {
			logx.Severe(err)
			return desertInfo, pubData, err
		}

		var assetId uint16
		if assetTokenAddress == types.BNBAddress {
			assetId = 0
		} else {
			assetId, err = monitor.ValidateAssetAddress(common.HexToAddress(assetTokenAddress))
			if err != nil {
				logx.Severe(err)
				return desertInfo, pubData, err
			}
		}

		accountsInfo[0].AssetsInfo = &cryptoTypes.AccountAsset{
			AssetId:                  formatAccountInfo.AssetInfo[int64(assetId)].AssetId,
			Balance:                  formatAccountInfo.AssetInfo[int64(assetId)].Balance,
			OfferCanceledOrFinalized: formatAccountInfo.AssetInfo[int64(assetId)].OfferCanceledOrFinalized,
		}

		assetMerkleProof, err := accountAssetTrees.Get(accountIndex).GetProof(uint64(assetId))
		if err != nil {
			return desertInfo, pubData, err
		}
		merkleProofsAccountAssets[0], err = prove.SetFixedAccountAssetArray(assetMerkleProof)
		if err != nil {
			return desertInfo, pubData, err
		}

		nft = cryptoTypes.EmptyNft(circuit.LastNftIndex)
		nftMerkleProofs, err := nftTree.GetProof(circuit.LastNftIndex)
		if err != nil {
			return desertInfo, pubData, err
		}
		merkleProofsNft, err = prove.SetFixedNftArray(nftMerkleProofs)
		if err != nil {
			return desertInfo, pubData, err
		}

		// padding empty account
		emptyAssetTree, err := tree.NewMemAccountAssetTree()
		if err != nil {
			return desertInfo, pubData, err
		}
		accountsInfo[1] = desertTypes.EmptyAccount(circuit.LastAccountIndex, tree.NilAccountAssetRoot)
		// get account
		accountMerkleProofs, err := accountTree.GetProof(circuit.LastAccountIndex)
		if err != nil {
			return desertInfo, pubData, err
		}
		// set account merkle proof
		merkleProofsAccounts[1], err = prove.SetFixedAccountArray(accountMerkleProofs)
		if err != nil {
			return desertInfo, pubData, err
		}
		assetMerkleProof, err = emptyAssetTree.GetProof(0)
		if err != nil {
			return desertInfo, pubData, err
		}
		merkleProofsAccountAssets[1], err = prove.SetFixedAccountAssetArray(assetMerkleProof)
		if err != nil {
			return desertInfo, pubData, err
		}

		exitTxInfo = &desert.ExitTx{
			AccountIndex: accountIndex,
			L1Address:    accountsInfo[0].L1Address,
			AssetId:      int64(assetId),
			AssetAmount:  formatAccountInfo.AssetInfo[int64(assetId)].Balance,
		}

		pubData = GenerateExitPubData(exitTxInfo)
	} else {
		nftInfo, err := c.bc.DB().L2NftModel.GetNft(nftIndex)
		if err != nil {
			logx.Errorf("get nft failed: %s", err)
			return desertInfo, pubData, err
		}
		nft = &cryptoTypes.Nft{
			NftIndex:            nftInfo.NftIndex,
			NftContentHash:      common.FromHex(nftInfo.NftContentHash),
			CreatorAccountIndex: nftInfo.CreatorAccountIndex,
			OwnerAccountIndex:   nftInfo.OwnerAccountIndex,
			RoyaltyRate:         nftInfo.RoyaltyRate,
			CollectionId:        nftInfo.CollectionId,
			NftContentType:      nftInfo.NftContentType,
		}
		if nftInfo.NftContentHash == "0" {
			logx.Errorf("nft %d is already withdrawed to L1", nft.NftIndex)
			return desertInfo, pubData, errors.New(fmt.Sprintf("Nft %d Already withdrawed to L1", nft.NftIndex))
		}
		nftMerkleProofs, err := nftTree.GetProof(uint64(nftIndex))
		if err != nil {
			return desertInfo, pubData, err
		}
		merkleProofsNft, err = prove.SetFixedNftArray(nftMerkleProofs)
		if err != nil {
			return desertInfo, pubData, err
		}

		accountsInfo[0].AssetsInfo = cryptoTypes.EmptyAccountAsset(circuit.LastAccountAssetId)
		assetMerkleProof, err := accountAssetTrees.Get(accountIndex).GetProof(circuit.LastAccountAssetId)
		if err != nil {
			return desertInfo, pubData, err
		}
		merkleProofsAccountAssets[0], err = prove.SetFixedAccountAssetArray(assetMerkleProof)
		if err != nil {
			return desertInfo, pubData, err
		}

		//padding creatorAccount
		creatorAccountInfo, err := c.bc.DB().AccountModel.GetAccountByIndex(nft.CreatorAccountIndex)
		if err != nil {
			logx.Errorf("get account failed: %s", err)
			return desertInfo, pubData, err
		}

		pk, err := common2.ParsePubKey(creatorAccountInfo.PublicKey)
		if err != nil {
			return desertInfo, pubData, err
		}
		accountsInfo[1] = &desertTypes.Account{
			AccountIndex:    creatorAccountInfo.AccountIndex,
			L1Address:       common2.AddressStrToBytes(creatorAccountInfo.L1Address),
			AccountPk:       pk,
			Nonce:           creatorAccountInfo.Nonce,
			CollectionNonce: creatorAccountInfo.CollectionNonce,
			AssetRoot:       accountAssetTrees.Get(creatorAccountInfo.AccountIndex).Root(),
		}

		accountsInfo[1].AssetsInfo = cryptoTypes.EmptyAccountAsset(circuit.LastAccountAssetId)

		// get account
		accountMerkleProofs, err := accountTree.GetProof(uint64(creatorAccountInfo.AccountIndex))
		if err != nil {
			return desertInfo, pubData, err
		}
		// set account merkle proof
		merkleProofsAccounts[1], err = prove.SetFixedAccountArray(accountMerkleProofs)
		if err != nil {
			return desertInfo, pubData, err
		}

		assetMerkleProof, err = accountAssetTrees.Get(creatorAccountInfo.AccountIndex).GetProof(circuit.LastAccountAssetId)
		if err != nil {
			return desertInfo, pubData, err
		}
		merkleProofsAccountAssets[1], err = prove.SetFixedAccountAssetArray(assetMerkleProof)
		if err != nil {
			return desertInfo, pubData, err
		}

		exitNftTxInfo = &desert.ExitNftTx{
			AccountIndex:        accountIndex,
			L1Address:           accountsInfo[0].L1Address,
			CreatorAccountIndex: nft.CreatorAccountIndex,
			CreatorL1Address:    common2.AddressStrToBytes(creatorAccountInfo.L1Address),
			RoyaltyRate:         nft.RoyaltyRate,
			NftIndex:            nft.NftIndex,
			CollectionId:        nft.CollectionId,
			NftContentHash:      nft.NftContentHash,
			NftContentType:      nft.NftContentType,
		}
		pubData = GenerateExitNftPubData(exitNftTxInfo)
	}

	cryptoTx := &desert.Tx{
		TxType:                    txType,
		ExitTxInfo:                exitTxInfo,
		ExitNftTxInfo:             exitNftTxInfo,
		AccountRoot:               accountTree.Root(),
		AccountsInfo:              accountsInfo,
		NftRoot:                   nftTree.Root(),
		Nft:                       nft,
		MerkleProofsAccountAssets: merkleProofsAccountAssets,
		MerkleProofsAccounts:      merkleProofsAccounts,
		MerkleProofsNft:           merkleProofsNft,
	}

	stateRoot := tree.ComputeStateRootHash(accountTree.Root(), nftTree.Root())
	desertInfo = &desert.Desert{
		StateRoot:  stateRoot,
		Commitment: common.Hex2Bytes(CreateCommitment(stateRoot, pubData)),
		Tx:         cryptoTx,
	}
	bz, err := json.Marshal(desertInfo)
	if err != nil {
		return desertInfo, pubData, err
	}
	logx.Infof("witness desertInfo=%s", bz)

	return desertInfo, pubData, nil
}

func CreateCommitment(
	stateRoot []byte,
	pubData []byte,
) string {
	var buf bytes.Buffer
	buf.Write(chain.CleanAndPaddingByteByModulus(stateRoot))
	buf.Write(pubData)
	commitment := common2.Sha56Hash(buf.Bytes())
	return common.Bytes2Hex(commitment)
}

func GenerateExitPubData(exitTxInfo *desert.ExitTx) []byte {
	var buf bytes.Buffer
	buf.WriteByte(uint8(desertTypes.TxTypeExit))
	buf.Write(common2.Uint32ToBytes(uint32(exitTxInfo.AccountIndex)))
	buf.Write(common2.Uint16ToBytes(uint16(exitTxInfo.AssetId)))
	buf.Write(common2.Uint128ToBytes(exitTxInfo.AssetAmount))
	buf.Write(exitTxInfo.L1Address)

	pubData := common2.SuffixPaddingBuToPubdataSize(buf.Bytes())
	return pubData
}

func GenerateExitNftPubData(exitNftTxInfo *desert.ExitNftTx) []byte {
	var buf bytes.Buffer
	buf.WriteByte(uint8(desertTypes.TxTypeExitNft))
	buf.Write(common2.Uint32ToBytes(uint32(exitNftTxInfo.AccountIndex)))
	buf.Write(common2.Uint32ToBytes(uint32(exitNftTxInfo.CreatorAccountIndex)))
	buf.Write(common2.Uint16ToBytes(uint16(exitNftTxInfo.RoyaltyRate)))
	buf.Write(common2.Uint40ToBytes(exitNftTxInfo.NftIndex))
	buf.Write(common2.Uint16ToBytes(uint16(exitNftTxInfo.CollectionId)))
	buf.Write(exitNftTxInfo.L1Address)
	buf.Write(exitNftTxInfo.CreatorL1Address)
	buf.Write(common2.PrefixPaddingBufToChunkSize(exitNftTxInfo.NftContentHash))
	buf.WriteByte(uint8(exitNftTxInfo.NftContentType))
	pubData := common2.SuffixPaddingBuToPubdataSize(buf.Bytes())
	return pubData
}

func (c *GenerateProof) initSmtTree(blockHeight int64) (accountTree bsmt.SparseMerkleTree, accountAssetTrees *tree.AssetTreeCache, nftTree bsmt.SparseMerkleTree, err error) {
	storedBlockInfo, err := c.getStoredBlockInfo()
	if err != nil {
		return accountTree, accountAssetTrees, nftTree, err
	}

	if c.config.TreeDB.Driver == tree.MemoryDB {
		accountTree, accountAssetTrees, nftTree, err = c.doInitSmtTree(blockHeight, true)
		if err != nil {
			return accountTree, accountAssetTrees, nftTree, err
		}
	} else {
		accountTree, accountAssetTrees, nftTree, err = c.doInitSmtTree(blockHeight, false)
		if err != nil {
			return accountTree, accountAssetTrees, nftTree, err
		}

		stateRoot := tree.ComputeStateRootHash(accountTree.Root(), nftTree.Root())
		if common.Bytes2Hex(stateRoot) != storedBlockInfo.StateRoot {
			nodeClient := redis.NewClient(&redis.Options{Addr: c.config.TreeDB.RedisDBOption.Addr, Password: c.config.TreeDB.RedisDBOption.Password})
			nodeClient.FlushAll(context.Background())
			accountTree, accountAssetTrees, nftTree, err = c.doInitSmtTree(blockHeight, true)
			if err != nil {
				return accountTree, accountAssetTrees, nftTree, err
			}
		}
	}
	stateRoot := tree.ComputeStateRootHash(accountTree.Root(), nftTree.Root())
	if err != nil {
		return accountTree, accountAssetTrees, nftTree, err
	}
	if common.Bytes2Hex(stateRoot) != storedBlockInfo.StateRoot {
		return accountTree, accountAssetTrees, nftTree, fmt.Errorf("stateRoot is not equal storedBlockInfo.StateRoot")
	}

	return accountTree, accountAssetTrees, nftTree, nil
}

func (c *GenerateProof) doInitSmtTree(blockHeight int64, reload bool) (accountTree bsmt.SparseMerkleTree, accountAssetTrees *tree.AssetTreeCache, nftTree bsmt.SparseMerkleTree, err error) {
	treeCtx, err := tree.NewContext("desertexit", c.config.TreeDB.Driver, reload, true, c.config.TreeDB.RoutinePoolSize, &c.config.TreeDB.LevelDBOption, &c.config.TreeDB.RedisDBOption, c.config.TreeDB.AssetTreeCacheSize, false, 200)
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
		treeCtx)
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
