package logic

import (
	"context"
	"errors"
	"fmt"
	blockModel "github.com/zecrey-labs/zecrey/common/model/block"
	"reflect"

	"github.com/zecrey-labs/zecrey/common/commonAccount"
	"github.com/zecrey-labs/zecrey/common/commonAsset"
	"github.com/zecrey-labs/zecrey/common/model/mempool"
	"github.com/zecrey-labs/zecrey/common/model/tx"
	"github.com/zecrey-labs/zecrey/common/utils"
	"github.com/zecrey-labs/zecrey/service/rpc/globalRPC/globalRPCProto"
	"github.com/zecrey-labs/zecrey/service/rpc/globalRPC/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLatestTxsListByAccountIndexLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLatestTxsListByAccountIndexLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLatestTxsListByAccountIndexLogic {
	return &GetLatestTxsListByAccountIndexLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func packGetLatestTxsListByAccountIndex(
	status int64,
	msg string,
	err string,
	result []*globalRPCProto.ResultGetLatestTxsListByAccountIndex,
) (res *globalRPCProto.RespGetLatestTxsListByAccountIndex) {
	res = &globalRPCProto.RespGetLatestTxsListByAccountIndex{
		Status: status,
		Msg:    msg,
		Err:    err,
		Result: result,
	}
	return res
}

func (l *GetLatestTxsListByAccountIndexLogic) GetLatestTxsListByAccountIndex(in *globalRPCProto.ReqGetLatestTxsListByAccountIndex) (*globalRPCProto.RespGetLatestTxsListByAccountIndex, error) {

	var (
		respResult      []*globalRPCProto.ResultGetLatestTxsListByAccountIndex
		mempoolTxs      []*mempool.MempoolTx
		mempoolTxAmount int
		txs             []*tx.Tx
	)

	err := utils.CheckRequestParam(utils.TypeAccountIndex, reflect.ValueOf(in.AccountIndex))
	if err != nil {
		errInfo := fmt.Sprintf("[logic.GetLatestTxsListByAccountIndex] %s", err)
		logx.Error(errInfo)
		return packGetLatestTxsListByAccountIndex(FailStatus, FailMsg, errInfo, respResult), nil
	}

	err = utils.CheckRequestParam(utils.TypeLimit, reflect.ValueOf(in.Limit))
	if err != nil {
		errInfo := fmt.Sprintf("[logic.GetLatestTxsListByAccountIndex] %s", err)
		logx.Error(errInfo)
		return packGetLatestTxsListByAccountIndex(FailStatus, FailMsg, errInfo, respResult), nil
	}

	mempoolTxs, err = l.svcCtx.MempoolModel.GetMempoolTxsListByAccountIndex(int64(in.AccountIndex), int64(in.Limit), 0)
	if err != nil {
		if err != mempool.ErrNotFound {
			errInfo := fmt.Sprintf("[logic.GetLatestTxsListByAccountIndex] => [MempoolModel.GetMempoolTxsListByAccountIndex]: %s. Invalid AccountIndex: %v",
				err.Error(), in.AccountIndex)
			logx.Error(errInfo)
			return packGetLatestTxsListByAccountIndex(FailStatus, FailMsg, errInfo, respResult), errors.New(errInfo)
		}
	}
	mempoolTxAmount = len(mempoolTxs)

	if uint64(mempoolTxAmount) < in.Limit {
		txs, err = l.svcCtx.TxModel.GetTxsListByAccountIndex(int64(in.AccountIndex), int(in.Limit)-mempoolTxAmount, 0)
		if err != nil {
			if err != tx.ErrNotFound {
				errInfo := fmt.Sprintf("[logic.GetLatestTxsListByAccountIndex] => [MempoolModel.GetMempoolTxsListByAccountIndex]: %s. Invalid AccountIndex: %v",
					err.Error(), in.AccountIndex)
				logx.Error(errInfo)
				return packGetLatestTxsListByAccountIndex(FailStatus, FailMsg, errInfo, respResult), errors.New(errInfo)
			}
		}
	}

	poolAccount, err := l.svcCtx.AccountHistoryModel.GetAccountByAccountIndex(commonAccount.PoolAccountIndex)
	if err != nil {
		errInfo := fmt.Sprintf("[logic.GetLatestTxsListByAccountIndex] => [AccountModel.GetAccountByAccountIndex] err: %s",
			err.Error())
		logx.Error(errInfo)
		return packGetLatestTxsListByAccountIndex(FailStatus, FailMsg, errInfo, respResult), errors.New(errInfo)
	}

	for _, v := range mempoolTxs {

		var (
			nTxDetails []*globalRPCProto.TxDetail
		)

		for _, w := range v.MempoolDetails {
			if w.AssetType == commonAsset.LiquidityAssetType && w.AccountIndex == commonAccount.PoolAccountIndex {

				poolDetails, err := utils.SplitPoolTx(utils.PoolTxDetail{
					AssetId:      w.AssetId,
					AssetType:    w.AssetType,
					AccountIndex: w.AccountIndex,
					AccountName:  w.AccountName,
					PublicKey:    poolAccount.PublicKey,
					BalanceEnc:   w.BalanceEnc,
					BalanceDelta: w.BalanceDelta,
				})
				if err != nil {
					errInfo := fmt.Sprintf("[logic.GetLatestTxsListByAccountIndex]<=>[utils.SplitPoolTx] %s", err.Error())
					logx.Error(errInfo)
					return packGetLatestTxsListByAccountIndex(FailStatus, "fail", errInfo, respResult), nil
				}

				for _, poolDetail := range poolDetails {
					nTxDetails = append(nTxDetails, &globalRPCProto.TxDetail{
						AssetId:           w.AssetId,
						AssetType:         w.AssetType,
						AccountIndex:      w.AccountIndex,
						AccountName:       w.AccountName,
						AccountBalanceEnc: poolDetail.BalanceEnc,
						AccountDeltaEnc:   poolDetail.BalanceDelta,
					})
				}
			} else {
				nTxDetails = append(nTxDetails, &globalRPCProto.TxDetail{
					AssetId:           w.AssetId,
					AssetType:         w.AssetType,
					AccountIndex:      w.AccountIndex,
					AccountName:       w.AccountName,
					AccountBalanceEnc: w.BalanceEnc,
					AccountDeltaEnc:   w.BalanceDelta,
				})
			}
		}
		respResult = append(respResult, &globalRPCProto.ResultGetLatestTxsListByAccountIndex{
			TxHash:        v.TxHash,
			TxType:        v.TxType,
			TxStatus:      tx.StatusPending,
			TxAssetAId:    v.AssetAId,
			TxAssetBId:    v.AssetBId,
			TxDetails:     nTxDetails,
			TxAmount:      v.TxAmount,
			NativeAddress: v.NativeAddress,
			GasFeeAssetId: v.GasFeeAssetId,
			GasFee:        v.GasFee,
			BlockStatus:   0,
			BlockHeight:   0,
			BlockId:       0,
			ChainId:       v.ChainId,
			Memo:          v.Memo,
			CreateAt:      v.CreatedAt.Unix(),
		})
	}

	for _, v := range txs {
		var (
			nTxDetails []*globalRPCProto.TxDetail
		)
		for _, w := range v.TxDetails {
			if w.AssetType == commonAsset.LiquidityAssetType && w.AccountIndex == commonAccount.PoolAccountIndex {

				poolDetails, err := utils.SplitPoolTx(utils.PoolTxDetail{
					AssetId:      w.AssetId,
					AssetType:    w.AssetType,
					AccountIndex: w.AccountIndex,
					AccountName:  w.AccountName,
					PublicKey:    poolAccount.PublicKey,
					BalanceEnc:   w.BalanceEnc,
					BalanceDelta: w.BalanceDelta,
				})
				if err != nil {
					errInfo := fmt.Sprintf("[logic.GetLatestTxsListByAccountIndex]<=>[utils.SplitPoolTx] %s", err.Error())
					logx.Error(errInfo)
					return packGetLatestTxsListByAccountIndex(FailStatus, "fail", errInfo, respResult), nil
				}

				for _, poolDetail := range poolDetails {
					nTxDetails = append(nTxDetails, &globalRPCProto.TxDetail{
						AssetId:           w.AssetId,
						AssetType:         w.AssetType,
						AccountIndex:      w.AccountIndex,
						AccountName:       w.AccountName,
						AccountBalanceEnc: poolDetail.BalanceEnc,
						AccountDeltaEnc:   poolDetail.BalanceDelta,
					})
				}
			} else {
				nTxDetails = append(nTxDetails, &globalRPCProto.TxDetail{
					AssetId:           w.AssetId,
					AssetType:         w.AssetType,
					AccountIndex:      w.AccountIndex,
					AccountName:       w.AccountName,
					AccountBalanceEnc: w.BalanceEnc,
					AccountDeltaEnc:   w.BalanceDelta,
				})
			}
		}

		var blockStatus int
		blockStatusInfo, err := l.svcCtx.BlockModel.GetBlockStatusCacheByBlockHeight(v.BlockHeight)
		if err == nil {
			// In Cache
			blockStatus = int(blockStatusInfo.BlockStatus)
		} else {
			// Not In Cache
			block, err := l.svcCtx.BlockModel.GetBlockByBlockHeight(v.BlockHeight)
			if err != nil {
				errInfo := fmt.Sprintf("[logic.GetLatestTxsListByAccountIndex]<=>[BlockModel.GetBlockByBlockHeight] %s", err.Error())
				logx.Error(errInfo)
				return packGetLatestTxsListByAccountIndex(FailStatus, "fail", errInfo, respResult), nil
			}
			blockStatusInfo = &blockModel.BlockStatusInfo{
				BlockStatus: block.BlockStatus,
				CommittedAt: block.CommittedAt,
				VerifiedAt:  block.VerifiedAt,
				ExecutedAt:  block.ExecutedAt,
			}

			err = l.svcCtx.BlockModel.UpdateBlockStatusCacheByBlockHeight(v.BlockHeight, blockStatusInfo)
			if err != nil {
				errInfo := fmt.Sprintf("[logic.GetLatestTxsListByAccountIndex]<=>[BlockModel.UpdateBlockStatusCacheByBlockHeight] %s", err.Error())
				logx.Error(errInfo)
				return packGetLatestTxsListByAccountIndex(FailStatus, "fail", errInfo, respResult), nil
			}

			blockStatus = int(block.BlockStatus)
		}

		respResult = append(respResult, &globalRPCProto.ResultGetLatestTxsListByAccountIndex{
			TxHash:        v.TxHash,
			TxType:        v.TxType,
			TxStatus:      tx.StatusPending,
			TxAssetAId:    v.AssetAId,
			TxAssetBId:    v.AssetBId,
			TxDetails:     nTxDetails,
			TxAmount:      v.TxAmount,
			NativeAddress: v.NativeAddress,
			GasFeeAssetId: v.GasFeeAssetId,
			GasFee:        v.GasFee,
			BlockStatus:   int64(blockStatus),
			BlockHeight:   v.BlockHeight,
			BlockId:       v.BlockId,
			Memo:          v.Memo,
			ChainId:       v.ChainId,
			CreateAt:      v.CreatedAt.Unix(),
		})
	}

	return packGetLatestTxsListByAccountIndex(SuccessStatus, SuccessMsg, "", respResult), nil
}
