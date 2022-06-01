package logic

import (
	"errors"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zecrey-labs/zecrey-eth-rpc/zecreyContract/core/zecrey/basic"
	"github.com/zecrey-labs/zecrey/common/model/block"
	"github.com/zecrey-labs/zecrey/common/utils"
)

func DefaultBlockHeader() StorageBlockHeader {
	return StorageBlockHeader{
		BlockNumber:    0,
		OnchainOpsRoot: basic.SetFixed32Bytes(common.FromHex("0x01ef55cdf3b9b0d65e6fb6317f79627534d971fd96c811281af618c0028d5e7a")),
		AccountRoot:    basic.SetFixed32Bytes(common.FromHex("0x01ef55cdf3b9b0d65e6fb6317f79627534d971fd96c811281af618c0028d5e7a")),
		Timestamp:      big.NewInt(0),
		Commitment:     basic.SetFixed32Bytes(common.FromHex("0x0000000000000000000000000000000000000000000000000000000000000000")),
	}
}

/*
	ConvertBlocksToCommitBlockInfos: helper function to convert blocks to commit block infos
*/
func ConvertBlocksToCommitBlockInfos(oBlocks []*Block, chainId int64) (commitBlocks []ZecreyCommitBlockInfo, err error) {
	for _, oBlock := range oBlocks {
		if chainId > int64(len(oBlock.BlockDetails)) {
			return commitBlocks, errors.New("[ConvertBlocksToCommitBlockInfos] invalid chain id")
		}
		// scan each block detail
		var detail *block.BlockDetail
		for _, blockDetail := range oBlock.BlockDetails {
			if blockDetail.ChainId == chainId {
				detail = blockDetail
			}
		}
		if detail == nil {
			log.Println("[ConvertBlocksToCommitBlockInfos] not valid block detail")
			return commitBlocks, errors.New("[ConvertBlocksToCommitBlockInfos] not valid block detail")
		}
		// TODO pub data computation
		OnchainOpsRoot := utils.StringToBytes(oBlock.OnChainOpsRoot)
		NewAccountRoot := utils.StringToBytes(oBlock.AccountRoot)
		Commitment := utils.StringToBytes(oBlock.BlockCommitment)
		OnchainOpsMerkleProof, err := utils.DeserializeOnChainOpsMerkleProofs(detail.OnChainOpsMerkleProof)
		if err != nil {
			log.Println("[ConvertBlocksToCommitBlockInfos] unable to deserialize merkle proofs:", err)
			return commitBlocks, err
		}
		commitBlock := ZecreyCommitBlockInfo{
			BlockNumber:           uint32(oBlock.BlockHeight),
			OnchainOpsRoot:        basic.SetFixed32Bytes(OnchainOpsRoot),
			NewAccountRoot:        basic.SetFixed32Bytes(NewAccountRoot),
			Timestamp:             big.NewInt(oBlock.CreatedAt.UnixMilli()),
			Commitment:            basic.SetFixed32Bytes(Commitment),
			OnchainOpsPubData:     detail.OnChainPublicData,
			OnchainOpsCount:       uint16(detail.OnChainOpsCount),
			OnchainOpsMerkleProof: OnchainOpsMerkleProof,
			PubData:               nil,
		}
		commitBlocks = append(commitBlocks, commitBlock)
	}
	return commitBlocks, nil
}

func ConstructStoredBlockHeader(oBlock *Block) StorageBlockHeader {
	return StorageBlockHeader{
		BlockNumber:    uint32(oBlock.BlockHeight),
		OnchainOpsRoot: basic.SetFixed32Bytes(common.FromHex(oBlock.OnChainOpsRoot)),
		AccountRoot:    basic.SetFixed32Bytes(common.FromHex(oBlock.AccountRoot)),
		Timestamp:      big.NewInt(oBlock.CreatedAt.UnixMilli()),
		Commitment:     basic.SetFixed32Bytes(common.FromHex(oBlock.BlockCommitment)),
	}
}

func ConvertBlocksToVerifyBlockInfos(oBlocks []*Block, blockForProverModel BlockForProverModel) (storedHeaders []StorageBlockHeader, proofs []*big.Int, err error) {
	//start := oBlocks[0].BlockHeight
	//end := oBlocks[len(oBlocks)].BlockHeight
	// get block for provers
	//rowsAffected, blockForProvers, err := blockForProverModel.GetBlockForProverBetween(start, end)
	//if err != nil {
	//	log.Println("[ConvertBlocksToVerifyBlockInfos] unable to get block for prover:", err)
	//	return nil, nil, err
	//}
	//if rowsAffected == 0 {
	//	log.Println("[ConvertBlocksToVerifyBlockInfos] no new proofs have been generated, please wait ... ")
	//	return nil, nil, errors.New("[ConvertBlocksToVerifyBlockInfos] no new proofs have not been generated, please wait ... ")
	//}
	// check if all blocks proofs have been generated
	//if rowsAffected != int64(len(oBlocks)) {
	//	log.Println("[ConvertBlocksToVerifyBlockInfos] some block proofs have not been generated, please wait")
	//	return nil, nil, errors.New("[ConvertBlocksToVerifyBlockInfos] some block proofs have not been generated, please wait")
	//}
	// scan each block
	for _, oBlock := range oBlocks {
		storedHeaders = append(storedHeaders, ConstructStoredBlockHeader(oBlock))
		// get proof from block for prover table
		//if blockForProvers[i].L2BlockHeight != oBlock.BlockHeight {
		//	log.Println("[ConvertBlocksToVerifyBlockInfos] invalid block height")
		//	return nil, nil, errors.New("[ConvertBlocksToVerifyBlockInfos] invalid block height")
		//}
		var proof []*big.Int
		//err := json.Unmarshal([]byte(blockForProvers[i].Proof), &proof)
		//if err != nil {
		//	log.Println("[ConvertBlocksToVerifyBlockInfos] unable to unmarshal proofs")
		//	return nil, nil, err
		//}
		proof = ConstructDefaultProof()
		proofs = append(proofs, proof...)
	}
	return storedHeaders, proofs, nil
}

func ConstructDefaultProof() (proof []*big.Int) {
	proof = make([]*big.Int, 8)
	proof[0], _ = new(big.Int).SetString("17078208247904226131286733154055458372525818460003097123106809680214660660082", 10)
	proof[1], _ = new(big.Int).SetString("19702613505590472224296506922758479786923858323185561694141199358113026328890", 10)
	proof[2], _ = new(big.Int).SetString("16548643058199238042076515865888386840834211765939991954677076386771660557182", 10)
	proof[3], _ = new(big.Int).SetString("12568835886964903255396934137419055326321609119935923357116305232266058601740", 10)
	proof[4], _ = new(big.Int).SetString("5781027727391213768252402520397578030602952471214020286050004425090329650393", 10)
	proof[5], _ = new(big.Int).SetString("12133572170098296220290503215565965720448491129367767241673216192139058662411", 10)
	proof[6], _ = new(big.Int).SetString("3010131470819733212728342353967163159559374591839707318025938739563202229653", 10)
	proof[7], _ = new(big.Int).SetString("17250058953677492602049969944836939226823990935912348269134967481198114877164", 10)
	return proof
}
