package logic

import (
	"context"
	"encoding/json"
	"fmt"
	cryptoBlock "github.com/bnb-chain/zkbas-crypto/legend/circuit/bn254/block"
	"github.com/bnb-chain/zkbas/common/model/proofSender"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/service/rpc/proverHub/internal/svc"
	"github.com/bnb-chain/zkbas/service/rpc/proverHub/proverHubProto"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"
)

type SubmitProofLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSubmitProofLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitProofLogic {
	return &SubmitProofLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func packSubmitProofLogic(
	status int64,
	msg string,
	err string,
	result *proverHubProto.ResultSubmitProof,
) (res *proverHubProto.RespSubmitProof) {
	return &proverHubProto.RespSubmitProof{
		Status: status,
		Msg:    msg,
		Err:    err,
		Result: result,
	}
}

func (l *SubmitProofLogic) SubmitProof(in *proverHubProto.ReqSubmitProof) (*proverHubProto.RespSubmitProof, error) {

	// Read Lock
	M.Lock()
	defer M.Unlock()

	var (
		result = &proverHubProto.ResultSubmitProof{}
	)

	// Unmarshal cBlock
	var (
		cBlock *cryptoBlock.Block
	)
	err := json.Unmarshal([]byte(in.BlockInfo), &cBlock)
	if err != nil {
		SetUnprovedCryptoBlockStatus(cBlock.BlockNumber, PUBLISHED)
		logx.Error(fmt.Sprintf("Unmarshal Error: %s", err.Error()))
		return packSubmitProofLogic(util.FailStatus, util.FailMsg, err.Error(), result), nil
	}

	// Unmarshal proof
	var (
		proof *util.FormattedProof
	)
	err = json.Unmarshal([]byte(in.Proof), &proof)
	if err != nil {
		SetUnprovedCryptoBlockStatus(cBlock.BlockNumber, PUBLISHED)
		logx.Error(fmt.Sprintf("Unmarshal Error: %s", err.Error()))
		return packSubmitProofLogic(util.FailStatus, util.FailMsg, err.Error(), result), nil
	}

	// load vk
	fmt.Println("start reading verifying key")
	// TODO vk file path
	vk, err := util.LoadVerifyingKey(VerifyingKeyPath)
	if err != nil {
		SetUnprovedCryptoBlockStatus(cBlock.BlockNumber, PUBLISHED)
		logx.Error(fmt.Sprintf("LoadVerifyingKey Error: %s", err.Error()))
		return packSubmitProofLogic(util.FailStatus, util.FailMsg, err.Error(), result), nil
	}

	oProof, err := util.UnformatProof(proof)
	if err != nil {
		SetUnprovedCryptoBlockStatus(cBlock.BlockNumber, PUBLISHED)
		logx.Error(fmt.Sprintf("UnformatProof Error: %s", err.Error()))
		return packSubmitProofLogic(util.FailStatus, util.FailMsg, err.Error(), result), nil
	}

	// VerifyProof
	err = util.VerifyProof(oProof, vk, cBlock)
	if err != nil {
		SetUnprovedCryptoBlockStatus(cBlock.BlockNumber, PUBLISHED)
		logx.Error(fmt.Sprintf("Verify Proof Error: %s", err.Error()))
		return packSubmitProofLogic(util.FailStatus, util.FailMsg, err.Error(), result), nil
	}

	// Handle Proof
	// Store Proof and BlockInfo into database and modify the status of UnprovedBlockList

	// modify UnprovedBlockList
	var blockStatus = GetUnprovedCryptoBlockStatus(cBlock.BlockNumber)
	if blockStatus != RECEIVED {
		SetUnprovedCryptoBlockStatus(cBlock.BlockNumber, PUBLISHED)
		logx.Error(fmt.Sprintf("block status error: %d", blockStatus))
		return packSubmitProofLogic(util.FailStatus, util.FailMsg, fmt.Sprintf("block status error: %d", blockStatus), result), nil
	}

	// check param
	provedBlock := GetUnprovedCryptoBlockByBlockNumber(cBlock.BlockNumber)
	if provedBlock != nil {
		if common.Bytes2Hex(provedBlock.BlockInfo.NewStateRoot[:]) == common.Bytes2Hex(cBlock.NewStateRoot) &&
			common.Bytes2Hex(provedBlock.BlockInfo.BlockCommitment[:]) == common.Bytes2Hex(cBlock.BlockCommitment) &&
			provedBlock.BlockInfo.CreatedAt == cBlock.CreatedAt {
			var row = &proofSender.ProofSender{
				ProofInfo:   in.Proof,
				BlockNumber: cBlock.BlockNumber,
				Status:      proofSender.NotSent,
			}
			err = l.svcCtx.ProofSenderModel.CreateProof(row)
			if err != nil {
				// rollback UnprovedList
				SetUnprovedCryptoBlockStatus(cBlock.BlockNumber, PUBLISHED)
				logx.Error(fmt.Sprintf("CreateProof error"))
				return packSubmitProofLogic(util.FailStatus, util.FailMsg, err.Error(), result), nil
			}
			logx.Info(fmt.Sprintf("CreateProof Successfully!"))
		} else {
			logx.Error(fmt.Sprintf("data inconsistency error"))
			return packSubmitProofLogic(util.FailStatus, util.FailMsg, "data inconsistency", result), nil
		}
	} else {
		logx.Error(fmt.Sprintf("get provedBlock error, provedBlock is nil"))
		return packSubmitProofLogic(util.FailStatus, util.FailMsg, "get provedBlock error", result), nil
	}

	return packSubmitProofLogic(util.SuccessStatus, util.SuccessMsg, util.NilErrorString, result), nil
}
