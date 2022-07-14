package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zecrey-labs/zecrey-legend/common/model/blockForProof"
	"github.com/zecrey-labs/zecrey-legend/common/util"
	lockUtil "github.com/zecrey-labs/zecrey-legend/common/util/globalmapHandler"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/proverHub/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/proverHub/proverHubProto"
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

	lock := lockUtil.GetRedisLockByKey(l.svcCtx.RedisConn, RedisLockKey)
	err := lockUtil.TryAcquireLock(lock)
	if err != nil {
		return packGetUnprovedBlockLogic(util.FailStatus, util.FailMsg, "block is locking", result), nil
	}
	defer lock.Release()

	// get crypto block with Mode
	cryptoBlockInfo, err := l.svcCtx.BlockForProofModel.GetUnprovedCryptoBlockByMode(in.Mode)
	if err != nil {
		logx.Errorf("get unproved crypto block error, mode=%d, err=%s", in.Mode, err.Error())
	}
	if cryptoBlockInfo == nil {
		return packGetUnprovedBlockLogic(util.FailStatus, util.FailMsg, "no unproved block", result), nil
	}

	// change crypto block status
	err = l.svcCtx.BlockForProofModel.UpdateUnprovedCryptoBlockStatus(cryptoBlockInfo, blockForProof.StatusReceived)
	if err != nil {
		return packGetUnprovedBlockLogic(util.FailStatus, util.FailMsg, "update block status error", result), nil
	}

	// write cryptoBlock to result
	result.BlockInfo = cryptoBlockInfo.BlockData
	return packGetUnprovedBlockLogic(util.SuccessStatus, util.SuccessMsg, util.NilErrorString, result), nil
}
