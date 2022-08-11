package transaction

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/errorcode"
	"github.com/bnb-chain/zkbas/service/api/app/internal/logic/transaction/sendrawtx"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
	"github.com/bnb-chain/zkbas/service/api/app/internal/types"
)

type SendTxLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendTxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendTxLogic {
	return &SendTxLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendTxLogic) SendTx(req *types.ReqSendTx) (resp *types.RespSendTx, err error) {
	resp = &types.RespSendTx{}
	switch req.TxType {
	case commonTx.TxTypeTransfer:
		resp.TxId, err = sendrawtx.SendTransferTx(l.ctx, l.svcCtx, l.svcCtx.StateFetcher, req.TxInfo)
		if err != nil {
			logx.Errorf("sendTransferTx err: %s", err.Error())
			return nil, err
		}
	case commonTx.TxTypeSwap:
		resp.TxId, err = sendrawtx.SendSwapTx(l.ctx, l.svcCtx, l.svcCtx.StateFetcher, req.TxInfo)
		if err != nil {
			logx.Errorf("sendSwapTx err: %s", err.Error())
			return nil, err
		}
	case commonTx.TxTypeAddLiquidity:
		resp.TxId, err = sendrawtx.SendAddLiquidityTx(l.ctx, l.svcCtx, l.svcCtx.StateFetcher, req.TxInfo)
		if err != nil {
			logx.Errorf("sendAddLiquidityTx err: %s", err.Error())
			return nil, err
		}
	case commonTx.TxTypeRemoveLiquidity:
		resp.TxId, err = sendrawtx.SendRemoveLiquidityTx(l.ctx, l.svcCtx, l.svcCtx.StateFetcher, req.TxInfo)
		if err != nil {
			logx.Errorf("sendRemoveLiquidityTx err: %s", err.Error())
			return nil, err
		}
	case commonTx.TxTypeWithdraw:
		resp.TxId, err = sendrawtx.SendWithdrawTx(l.ctx, l.svcCtx, l.svcCtx.StateFetcher, req.TxInfo)
		if err != nil {
			logx.Errorf("sendWithdrawTx err: %s", err.Error())
			return nil, err
		}
	case commonTx.TxTypeTransferNft:
		resp.TxId, err = sendrawtx.SendTransferNftTx(l.ctx, l.svcCtx, l.svcCtx.StateFetcher, req.TxInfo)
		if err != nil {
			logx.Errorf("sendTransferNftTX err: %s", err.Error())
			return nil, err
		}
	case commonTx.TxTypeAtomicMatch:
		resp.TxId, err = sendrawtx.SendAtomicMatchTx(l.ctx, l.svcCtx, l.svcCtx.StateFetcher, req.TxInfo)
		if err != nil {
			logx.Errorf("sendAtomicMatchTx err: %s", err.Error())
			return nil, err
		}
	case commonTx.TxTypeCancelOffer:
		resp.TxId, err = sendrawtx.SendCancelOfferTx(l.ctx, l.svcCtx, l.svcCtx.StateFetcher, req.TxInfo)
		if err != nil {
			logx.Errorf("sendCancelOfferTx err: %s", err.Error())
			return nil, err
		}
	case commonTx.TxTypeWithdrawNft:
		resp.TxId, err = sendrawtx.SendWithdrawNftTx(l.ctx, l.svcCtx, l.svcCtx.StateFetcher, req.TxInfo)
		if err != nil {
			logx.Errorf("sendWithdrawNftTx err: %s", err.Error())
			return nil, err
		}
	case commonTx.TxTypeCreateCollection:
		resp.TxId, err = sendrawtx.SendCreateCollectionTx(l.ctx, l.svcCtx, l.svcCtx.StateFetcher, req.TxInfo)
		if err != nil {
			logx.Errorf("sendCreateCollectionTx err: %s", err.Error())
			return nil, err
		}
	case commonTx.TxTypeMintNft:
		resp.TxId, err = sendrawtx.SendMintNftTx(l.ctx, l.svcCtx, l.svcCtx.StateFetcher, req.TxInfo)
		if err != nil {
			logx.Errorf("sendMintNftTx err: %s", err.Error())
			return nil, err
		}
	case commonTx.TxTypeOffer:
		logx.Errorf("invalid tx type: %d", req.TxType)
		return nil, errorcode.AppErrInvalidTxType
	default:
		logx.Errorf("invalid tx type: %d", req.TxType)
		return nil, errorcode.AppErrInvalidTxType
	}
	return resp, err
}
