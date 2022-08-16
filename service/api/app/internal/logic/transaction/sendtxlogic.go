package transaction

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/transaction/sendrawtx"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type SendTxLogic struct {
	logx.Logger
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	txSenders map[int]sendrawtx.TxSender
}

func NewSendTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendTxLogic {
	gasChecker := sendrawtx.NewGasChecker(svcCtx.SysConfigModel)
	nonceChecker := sendrawtx.NewNonceChecker()
	mempoolTxSender := sendrawtx.NewMempoolTxSender(svcCtx.MempoolModel, svcCtx.FailTxModel)

	txSenders := make(map[int]sendrawtx.TxSender)
	txSenders[commonTx.TxTypeAddLiquidity] = sendrawtx.NewAddLiquidityTxSender(ctx, svcCtx, gasChecker, nonceChecker, mempoolTxSender)
	txSenders[commonTx.TxTypeAtomicMatch] = sendrawtx.NewAtomicMatchTxSender(ctx, svcCtx, gasChecker, nonceChecker, mempoolTxSender)
	txSenders[commonTx.TxTypeCancelOffer] = sendrawtx.NewCancelTxSender(ctx, svcCtx, gasChecker, nonceChecker, mempoolTxSender)
	txSenders[commonTx.TxTypeCreateCollection] = sendrawtx.NewCreateCollectionTxSender(ctx, svcCtx, gasChecker, nonceChecker, mempoolTxSender)
	txSenders[commonTx.TxTypeMintNft] = sendrawtx.NewMintNftTxSender(ctx, svcCtx, gasChecker, nonceChecker, mempoolTxSender)
	txSenders[commonTx.TxTypeRemoveLiquidity] = sendrawtx.NewRemoveLiquidityTxSender(ctx, svcCtx, gasChecker, nonceChecker, mempoolTxSender)
	txSenders[commonTx.TxTypeSwap] = sendrawtx.NewSwapTxSender(ctx, svcCtx, gasChecker, nonceChecker, mempoolTxSender)
	txSenders[commonTx.TxTypeTransferNft] = sendrawtx.NewTransferNftTxSender(ctx, svcCtx, gasChecker, nonceChecker, mempoolTxSender)
	txSenders[commonTx.TxTypeTransfer] = sendrawtx.NewTransferTxSender(ctx, svcCtx, gasChecker, nonceChecker, mempoolTxSender)
	txSenders[commonTx.TxTypeWithdrawNft] = sendrawtx.NewWithdrawNftTxSender(ctx, svcCtx, gasChecker, nonceChecker, mempoolTxSender)
	txSenders[commonTx.TxTypeWithdraw] = sendrawtx.NewWithdrawTxSender(ctx, svcCtx, gasChecker, nonceChecker, mempoolTxSender)

	return &SendTxLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		txSenders: txSenders,
	}
}

func (l *SendTxLogic) SendTx(req *types.ReqSendTx) (resp *types.RespSendTx, err error) {
	resp = &types.RespSendTx{}
	sender, ok := l.txSenders[int(req.TxType)]
	if !ok {
		logx.Errorf("invalid tx type: %d", req.TxType)
		return nil, errorcode.AppErrInvalidTxType
	}
	resp.TxId, err = sender.SendTx(req.TxInfo)
	if err != nil {
		return nil, err
	}
	return resp, err
}
