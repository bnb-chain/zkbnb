package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/service/rpc/proverHub/internal/svc"
	"github.com/bnb-chain/zkbas/service/rpc/proverHub/proverHubProto"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetUnprovedBlockLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUnprovedBlockLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUnprovedBlockLogic {
	return &GetUnprovedBlockLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func packGetUnprovedBlockLogic(
	status int64,
	msg string,
	err string,
	result *proverHubProto.ResultGetUnprovedBlock,
) (res *proverHubProto.RespGetUnprovedBlock) {
	return &proverHubProto.RespGetUnprovedBlock{
		Status: status,
		Msg:    msg,
		Err:    err,
		Result: result,
	}
}

func (l *GetUnprovedBlockLogic) GetUnprovedBlock(in *proverHubProto.ReqGetUnprovedBlock) (*proverHubProto.RespGetUnprovedBlock, error) {
	var (
		result = &proverHubProto.ResultGetUnprovedBlock{}
	)
	// Lock
	// todo distributed lock

	var tryLockStatus = M.TryLock()
	fmt.Printf("TryLock: %t\n", tryLockStatus)
	if !tryLockStatus {
		return packGetUnprovedBlockLogic(util.FailStatus, util.FailMsg, "block is locking", result), nil
	}
	defer M.Unlock()

	// get crypto block with Mode
	cryptoBlockInfo := GetUnprovedCryptoBlock(in.Mode)
	if cryptoBlockInfo == nil {
		return packGetUnprovedBlockLogic(util.FailStatus, util.FailMsg, "no unproved block", result), nil
	}
	// change crypto block status
	cryptoBlockInfo.Status = RECEIVED

	cryptoBlockInfoBytes, err := json.Marshal(cryptoBlockInfo.BlockInfo)

	if err != nil {
		return packGetUnprovedBlockLogic(util.FailStatus, util.FailMsg, err.Error(), result), nil
	}

	// write cryptoBlock to result
	result.BlockInfo = string(cryptoBlockInfoBytes)

	return packGetUnprovedBlockLogic(util.SuccessStatus, util.SuccessMsg, util.NilErrorString, result), nil
}
