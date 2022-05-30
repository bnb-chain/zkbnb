package transaction

import (
	"context"
	"fmt"
	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/repo/mempool"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/svc"
	"github.com/zecrey-labs/zecrey-legend/service/api/app/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"strconv"
)

type GetMempoolTxsLogic struct {
	logx.Logger
	ctx     context.Context
	svcCtx  *svc.ServiceContext
	mempool mempool.Mempool
}

func NewGetMempoolTxsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMempoolTxsLogic {
	return &GetMempoolTxsLogic{
		Logger:  logx.WithContext(ctx),
		ctx:     ctx,
		svcCtx:  svcCtx,
		mempool: mempool.New(svcCtx.Config),
	}
}
func packGetMempoolTxsResp(
	offset uint16,
	limit uint16,
	total uint32,
	mempoolTxs []*types.Tx,
) (res *types.RespGetMempoolTxs) {
	res = &types.RespGetMempoolTxs{
		Offset:     offset,
		Limit:      limit,
		Total:      total,
		MempoolTxs: mempoolTxs,
	}
	return res
}
func (l *GetMempoolTxsLogic) GetMempoolTxs(req *types.ReqGetMempoolTxs) (resp *types.RespGetMempoolTxs, err error) {
	//	err := utils.CheckRequestParam(utils.TypeLimit, reflect.ValueOf(req.Limit))
	//	err = utils.CheckRequestParam(utils.TypeLimit, reflect.ValueOf(req.Limit))
	mempoolTxs, err := l.mempool.GetMempoolTxs(int64(req.Limit), int64(req.Offset))
	logx.Info(req.Limit, req.Offset)
	if err != nil {
		errInfo := fmt.Sprintf("[appService.transaction.GetMempoolTxsList]<=>[MempoolModel.GetMempoolTxsList] %s", err.Error())
		logx.Error(errInfo)
		return packGetMempoolTxsResp(req.Offset, req.Limit, 0, nil), nil
	}

	// Todo: why not do total=len(mempoolTxs)
	total, err := l.mempool.GetMempoolTxsTotalCount()
	if err != nil {
		errInfo := fmt.Sprintf("[appService.transaction.GetMempoolTxsList]<=>[MempoolModel.GetMempoolTxsTotalCount] %s", err.Error())
		logx.Error(errInfo)
		return packGetMempoolTxsResp(req.Offset, req.Limit, 0, nil), nil
	}

	data := make([]*types.Tx, 0)
	for _, mempoolTx := range mempoolTxs {
		txDetails := make([]*types.TxDetail, 0)
		for _, txDetail := range mempoolTx.MempoolDetails {

			if txDetail.AssetType == commonAsset.LiquidityAssetType {
				//Todo: add json string of liquidity transaction to the list
			} else {
				txDetails = append(txDetails, &types.TxDetail{
					//Todo: verify if accountBalance is still needed, since its no longer a field of table TxDetail
					//Todo: int64 or int?
					//Todo: need balance or not?  no need
					AssetId:      int(txDetail.AssetId),
					AssetType:    int(txDetail.AssetType),
					AccountIndex: int32(txDetail.AccountIndex),
					AccountName:  txDetail.AccountName,
					AccountDelta: txDetail.BalanceDelta,
				})
			}
		}
		//Todo: int64 or int?
		txAmount, err := strconv.Atoi(mempoolTx.TxAmount)
		if err != nil {
			errInfo := fmt.Sprintf("[appService.transaction.GetMempoolTxsList]<=>[MempoolModel.GetMempoolTxsTotalCount] %s", err.Error())
			logx.Error(errInfo)
			return packGetMempoolTxsResp(req.Offset, req.Limit, 0, nil), nil
		}
		// Todo: why is the field in db string?
		gasFee, err := strconv.Atoi(mempoolTx.GasFee)
		data = append(data, &types.Tx{
			TxHash:        mempoolTx.TxHash,
			TxType:        uint8(mempoolTx.TxType),
			AssetAId:      int(mempoolTx.AssetAId),
			AssetBId:      int(mempoolTx.AssetBId),
			TxDetails:     txDetails,
			TxAmount:      txAmount,
			NativeAddress: mempoolTx.NativeAddress,
			TxStatus:      1, //pending
			GasFeeAssetId: uint16(mempoolTx.GasFeeAssetId),
			GasFee:        int64(gasFee),
			CreatedAt:     mempoolTx.CreatedAt.Unix(),
			Memo:          mempoolTx.Memo,
		})
	}
	resp = &types.RespGetMempoolTxs{
		Offset:     req.Offset,
		Limit:      req.Limit,
		Total:      uint32(total),
		MempoolTxs: data,
	}
	return resp, nil
}
