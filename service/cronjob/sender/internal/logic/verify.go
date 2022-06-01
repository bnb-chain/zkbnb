package logic

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zecrey-labs/zecrey/common/model/l1TxSender"
	"github.com/zecrey-labs/zecrey/common/model/proofSender"
	"github.com/zecrey-labs/zecrey/common/utils"
	"io"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/zecrey-labs/zecrey-eth-rpc/zecreyContract/core/zecrey/basic"
)

func SendVerifiedBlocks(
	params *SenderParam,
	l1TxSenderModel L1TxSenderModel,
	proofSenderModel ProofSenderModel,

) (err error) {

	var storedHeaders []StorageBlockHeader
	var proofs []*proofSender.ProofSender
	var proofsBigInt []*big.Int

	// scan l1 tx sender table for handled verified height
	lastHandledBlock, err := l1TxSenderModel.GetLatestHandledBlock(params.ChainId, VerifyTxType)
	if err != nil {
		if err != l1TxSender.ErrNotFound {
			log.Println("[SendVerifiedBlocks] unable to get latest handled block:", err)
			return err
		}
	}

	// check if the block has been verified
	var mainChainLastHandledHeight int64
	if params.Mode == MainChain {
		// if this chain is not main chain, check if main chain has been verified
		if params.ChainId != params.MainChainId {
			mainChainLastHandledBlock, err := l1TxSenderModel.GetLatestHandledBlock(params.MainChainId, VerifyTxType)
			if err != nil {
				log.Println("[SendVerifiedBlocks] unable to get latest handled block:", err)
				return err
			}
			if mainChainLastHandledBlock == nil {
				log.Println("[SendVerifiedBlocks] main chain has not been verified, should wait")
				return nil
			}
			// check main chain verified height
			mainChainLastHandledHeight = mainChainLastHandledBlock.L2BlockHeight
		}
	}

	// if lastHandledBlock == nil, means we haven't verified any blocks, just start from 0
	if lastHandledBlock == nil {
		// scan l1 tx sender table for pending verified height that higher than the latest handled height
		rowsAffected, pendingSenders, err := l1TxSenderModel.GetLatestPendingBlocks(params.ChainId, VerifyTxType)
		if err != nil {
			log.Println("[SendVerifiedBlocks] unable to get latest pending blocks:", err)
			return err
		}
		// if rowsAffected == 0, means we haven't verified any new blocks in any chain
		if rowsAffected == 0 {
			// get blocks from block table
			if params.Mode == MainChain && (params.ChainId != params.MainChainId) {
				// catch up main chain progress

				proofs, err = proofSenderModel.GetProofsByBlockRange(1, mainChainLastHandledHeight, params.MaxBlockCount)
				if err != nil {
					log.Println("[SendVerifiedBlocks] unable to get sender new blocks:", err)
					return err
				}
			} else { // StandAlone Mode || Not MainChainId
				// send all blocks satisfied requirement
				proofs, err = proofSenderModel.GetProofsByBlockRange(1, int64(1+params.MaxBlockCount), params.MaxBlockCount)
				if err != nil {
					log.Println("[SendVerifiedBlocks] unable to get sender new blocks:", err)
					return err
				}
			}

			storedHeaders, proofsBigInt, err = utils.ConvertToStorageBlockHeaderAndProof(proofs)
			if err != nil {
				log.Println("[ConvertToStorageBlockHeaderAndProof] ConvertToStorageBlockHeaderAndProof error:", err)
				return err
			}

		} else {
			// table panic
			if rowsAffected != 1 {
				log.Println("[SendVerifiedBlocks] a terrible error, need to check!!! ")
				return errors.New("[SendVerifiedBlocks] a terrible error, need to check!!! ")
			}
			// if rowsAffected != 0, means there is no handled verified block in this chain but there is pending block.
			// check if the tx is pending or fail, if fail re-send it.
			pendingSender := pendingSenders[0]
			_, isPending, err := params.Cli.GetTransactionByHash(pendingSender.L1TxHash)
			// if err != nil, means we cannot get this tx by hash
			if err != nil {
				// if we cannot get it from rpc and the time over 1 min
				lastUpdatedAt := pendingSender.UpdatedAt.UnixMilli()
				now := time.Now().UnixMilli()
				if now-lastUpdatedAt > params.MaxWaitingTime {
					// drop the record
					err := l1TxSenderModel.DeleteL1TxSender(pendingSender)
					if err != nil {
						log.Println("[SendVerifiedBlocks] unable to delete l1 tx sender:", err)
						return err
					}
					return nil
				} else {
					log.Println("[SendVerifiedBlocks] tx cannot be found, but not exceed time limit", pendingSender.L1TxHash)
					return nil
				}
			}
			// if it is pending, still waiting
			if isPending {
				log.Println("[SendVerifiedBlocks] tx is still pending, no need to work for anything tx hash:", pendingSender.L1TxHash)
				return nil
			}
		}
	} else {
		// if lastHandledBlock != nil, means there is handled verified blocks
		// scan l1 tx sender table for pending verified height that higher than the latest handled height
		rowsAffected, pendingSenders, err := l1TxSenderModel.GetLatestPendingBlocks(params.ChainId, VerifyTxType)
		if err != nil {
			log.Println("[SendVerifiedBlocks] unable to get latest pending blocks:", err)
			return err
		}
		// if rowsAffected == 0, means there is no pending blocks to wait. Just sending it!
		if rowsAffected == 0 {
			// get proofs
			if params.Mode == MainChain && (params.ChainId != params.MainChainId) {
				// catch up main chain progress

				proofs, err = proofSenderModel.GetProofsByBlockRange(lastHandledBlock.L2BlockHeight + 1, mainChainLastHandledHeight, params.MaxBlockCount)
				if err != nil {
					log.Println("[SendVerifiedBlocks] unable to get sender new blocks:", err)
					return err
				}
			} else {
				// send all blocks satisfied requirement
				proofs, err = proofSenderModel.GetProofsByBlockRange(lastHandledBlock.L2BlockHeight + 1, lastHandledBlock.L2BlockHeight+1+int64(params.MaxBlockCount), params.MaxBlockCount)
				if err != nil {
					log.Println("[SendVerifiedBlocks] unable to get sender new blocks:", err)
					return err
				}
			}

			storedHeaders, proofsBigInt, err = utils.ConvertToStorageBlockHeaderAndProof(proofs)
			if err != nil {
				log.Println("[ConvertToStorageBlockHeaderAndProof] ConvertToStorageBlockHeaderAndProof error:", err)
				return err
			}

		} else {
			// means there is pending blocks. Need to check it!
			if rowsAffected != 1 {
				log.Println("[SendVerifiedBlocks] a terrible error, need to check!!! ")
				return errors.New("[SendVerifiedBlocks] a terrible error, need to check!!! ")
			}
			// if rowsAffected != 0, check if the tx is pending or fail, if fail re-send it.
			pendingSender := pendingSenders[0]
			isSuccess, err := params.Cli.WaitingTransactionStatus(pendingSender.L1TxHash)
			// if err != nil, means we cannot get this tx by hash
			if err != nil {
				// if we cannot get it from rpc and the time over 1 min
				lastUpdatedAt := pendingSender.UpdatedAt.UnixMilli()
				now := time.Now().UnixMilli()
				if now-lastUpdatedAt > params.MaxWaitingTime {
					// drop the record
					err := l1TxSenderModel.DeleteL1TxSender(pendingSender)
					if err != nil {
						log.Println("[SendVerifiedBlocks] unable to delete l1 tx sender:", err)
						return err
					}
					return nil
				} else {
					log.Println("[SendVerifiedBlocks] tx cannot be found, but not exceed time limit", pendingSender.L1TxHash)
					return nil
				}
			}
			// if it is pending, still waiting
			if !isSuccess {
				log.Println("[SendVerifiedBlocks] tx is still pending, no need to work for anything tx hash:", pendingSender.L1TxHash)
				return nil
			}
		}
	}

	// commit blocks on-chain
	if len(storedHeaders) != 0 {
		if params.DebugParams != nil {
			// file output start
			var (
				f *os.File
				filename = params.DebugParams.FilePrefix + "/verify.txt"
			)
			if utils.CheckFileIsExist(filename) {
				f, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
				if err != nil {
					panic("OpenFile error")
				}
				fmt.Println("file exists")
			} else {
				f, err = os.Create(filename)
				fmt.Println("file doesn't exists")
				if err != nil {
					panic("Create file error")
				}
			}
			defer f.Close()

			for i, v:= range storedHeaders{
				_, err = io.WriteString(
					f,
					fmt.Sprintf("storedHeaders Input %d\n" +
						"BlockNumber : %d\n" +
						"Timestamp : %d\n" +
						"NewAccountRoot : %s\n" +
						"Commitment : %s\n" +
						"OnchainOpsRoot : %s\n",
						i,
						v.BlockNumber,
						v.Timestamp,
						common.Bytes2Hex(v.AccountRoot[:]),
						common.Bytes2Hex(v.Commitment[:]),
						common.Bytes2Hex(v.OnchainOpsRoot[:])),
				)
			}
			// file output end
		}

		txHash, err := basic.ZecreyVerifyBlocks(
			params.Cli, params.AuthCli, params.ZecreyInstance, storedHeaders, proofsBigInt, params.GasPrice, params.GasLimit,
		)
		if err != nil {
			log.Println("[SendVerifiedBlocks] unable to verify blocks:", err)
			return err
		}
		// update l1 tx sender table records
		newSender := &L1TxSender{
			ChainId:       uint8(params.ChainId),
			L1TxHash:      txHash,
			TxStatus:      PendingStatus,
			TxType:        VerifyTxType,
			L2BlockHeight: int64(storedHeaders[len(storedHeaders)-1].BlockNumber),
		}
		isValid, err := l1TxSenderModel.CreateL1TxSender(newSender) // todo add proofSender status modification
		if err != nil {
			log.Println("[SendCommittedBlocks] unable to create l1 tx sender")
			return err
		}
		if !isValid {
			log.Println("[SendVerifiedBlocks] cannot create new senders")
			return errors.New("[SendVerifiedBlocks] cannot create new senders")
		}
		log.Println("[SendVerifiedBlocks] new blocks have been verified(height):", newSender.L2BlockHeight)
		return nil
	} else {
		log.Println("[SendVerifiedBlocks] no new blocks need to verify")
		return nil
	}
}
