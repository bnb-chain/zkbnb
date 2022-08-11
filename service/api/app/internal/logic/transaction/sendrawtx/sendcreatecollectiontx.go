package sendrawtx

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/nft"
	"github.com/bnb-chain/zkbas/common/zcrypto/txVerification"
	"github.com/bnb-chain/zkbas/service/api/app/internal/fetcher/state"
	"github.com/bnb-chain/zkbas/service/api/app/internal/svc"
)

func SendCreateCollectionTx(ctx context.Context, svcCtx *svc.ServiceContext, stateFetcher state.Fetcher, rawTxInfo string) (txId string, err error) {
	txInfo, err := commonTx.ParseCreateCollectionTxInfo(rawTxInfo)
	if err != nil {
		logx.Errorf("cannot parse tx err: %s", err.Error())
		return "", errorcode.AppErrInvalidTx
	}

	if err := legendTxTypes.ValidateCreateCollectionTxInfo(txInfo); err != nil {
		logx.Errorf("cannot pass static check, err: %s", err.Error())
		return "", errorcode.AppErrInvalidTxField.RefineError(err)
	}

	if err := CheckGasAccountIndex(txInfo.GasAccountIndex, svcCtx.SysConfigModel); err != nil {
		return "", err
	}

	var (
		accountInfoMap = make(map[int64]*commonAsset.AccountInfo)
	)
	accountInfoMap[txInfo.AccountIndex], err = stateFetcher.GetLatestAccountInfo(ctx, txInfo.AccountIndex)
	if err != nil {
		if err == errorcode.DbErrNotFound {
			return "", errorcode.AppErrInvalidTxField.RefineError("invalid FromAccountIndex")
		}
		logx.Errorf("unable to get account info by index: %d, err: %s", txInfo.AccountIndex, err.Error())
		return "", errorcode.AppErrInternal
	}
	if accountInfoMap[txInfo.GasAccountIndex] == nil {
		accountInfoMap[txInfo.GasAccountIndex], err = stateFetcher.GetBasicAccountInfo(ctx, txInfo.GasAccountIndex)
		if err != nil {
			if err == errorcode.DbErrNotFound {
				return "", errorcode.AppErrInvalidTxField.RefineError("invalid GasAccountIndex")
			}
			logx.Errorf("unable to get account info by index: %d, err: %s", txInfo.GasAccountIndex, err.Error())
			return "", errorcode.AppErrInternal
		}
	}

	txInfo.CollectionId = accountInfoMap[txInfo.AccountIndex].CollectionNonce

	var (
		txDetails []*mempool.MempoolTxDetail
	)
	txDetails, err = txVerification.VerifyCreateCollectionTxInfo(accountInfoMap, txInfo)
	if err != nil {
		return "", errorcode.AppErrVerification.RefineError(err)
	}

	// write into mempool
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return "", errorcode.AppErrInternal
	}
	txId, mempoolTx := ConstructMempoolTx(
		commonTx.TxTypeCreateCollection,
		txInfo.GasFeeAssetId,
		txInfo.GasFeeAssetAmount.String(),
		commonConstant.NilTxNftIndex,
		commonConstant.NilPairIndex,
		commonConstant.NilAssetId,
		accountInfoMap[txInfo.AccountIndex].AccountName,
		commonConstant.NilL1Address,
		string(txInfoBytes),
		"",
		txInfo.AccountIndex,
		txInfo.Nonce,
		txInfo.ExpiredAt,
		txDetails,
	)
	// construct nft Collection info
	nftCollectionInfo := &nft.L2NftCollection{
		CollectionId: txInfo.CollectionId,
		AccountIndex: txInfo.AccountIndex,
		Name:         txInfo.Name,
		Introduction: txInfo.Introduction,
		Status:       nft.CollectionPending,
	}
	if err = createMempoolTxForCreateCollection(nftCollectionInfo, mempoolTx, svcCtx); err != nil {
		logx.Errorf("fail to create mempool tx: %v, err: %s", mempoolTx, err.Error())
		_ = CreateFailTx(svcCtx.FailTxModel, commonTx.TxTypeCreateCollection, txInfo, err)
		return "", errorcode.AppErrInternal
	}
	return txId, nil
}

func createMempoolTxForCreateCollection(
	nftCollectionInfo *nft.L2NftCollection,
	nMempoolTx *mempool.MempoolTx,
	svcCtx *svc.ServiceContext,
) (err error) {
	// check collectionId exist
	exist, err := svcCtx.CollectionModel.IfCollectionExistsByCollectionId(nftCollectionInfo.AccountIndex, nftCollectionInfo.CollectionId)
	if err != nil {
		return err
	}
	if exist {
		logx.Errorf("collectionId duplicate creation: %d", nftCollectionInfo.CollectionId)
		return errors.New("collectionId duplicate creation")
	}

	// write into mempool
	if err := svcCtx.MempoolModel.CreateMempoolTxAndL2CollectionAndNonce(nMempoolTx, nftCollectionInfo); err != nil {
		return err
	}
	return nil
}
