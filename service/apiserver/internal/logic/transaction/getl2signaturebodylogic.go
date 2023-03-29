package transaction

import (
	"context"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/signature"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetL2SignatureBodyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetL2SignatureBodyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetL2SignatureBodyLogic {
	return &GetL2SignatureBodyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetL2SignatureBodyLogic) GetL2SignatureBody(req *types.ReqSendTx) (resp *types.SignBody, err error) {

	txType := req.TxType
	txInfo := req.TxInfo

	signatureBody, err := signature.GenerateSignatureBody(txType, txInfo)
	if err != nil {
		return nil, err
	}

	resp = &types.SignBody{
		SignBody: signatureBody.GetL1SignatureBody(),
	}
	return resp, nil
}
