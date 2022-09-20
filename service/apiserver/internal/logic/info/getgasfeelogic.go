package info

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbnb/types"
)

type GetGasFeeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetGasFeeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGasFeeLogic {
	return &GetGasFeeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetGasFeeLogic) GetGasFee(req *types.ReqGetGasFee) (*types.GasFee, error) {
	gasFeeConfig, err := l.svcCtx.MemCache.GetSysConfigWithFallback(types2.SysGasFee, func() (interface{}, error) {
		return l.svcCtx.SysConfigModel.GetSysConfigByName(types2.SysGasFee)
	})
	if err != nil {
		logx.Errorf("fail to get gas fee config, err: %s", err.Error())
		return nil, types2.AppErrInternal
	}
	m := make(map[uint32]map[int]int64)
	err = json.Unmarshal([]byte(gasFeeConfig.Value), &m)
	if err != nil {
		logx.Errorf("fail to unmarshal gas fee config, err: %s", err.Error())
		return nil, types2.AppErrInternal
	}

	gasAsset, ok := m[req.AssetId]
	if !ok {
		logx.Errorf("cannot find gas config for asset id: %d", req.AssetId)
		return nil, types2.AppErrInvalidGasAsset
	}
	gasFee, ok := gasAsset[int(req.TxType)]
	if !ok {
		return nil, types2.AppErrInvalidTxType
	}

	resp := &types.GasFee{
		GasFee: strconv.FormatInt(gasFee, 10),
	}
	return resp, nil
}
