package sendrawtx

import (
	"encoding/base64"
	"encoding/json"
	"strconv"

	"github.com/zeromicro/go-zero/core/stores/redis"

	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/common/util"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"

	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/model/mempool"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/model/sysconfig"
	"github.com/bnb-chain/zkbas/common/sysconfigName"
	"github.com/bnb-chain/zkbas/errorcode"
)

func CheckGasAccountIndex(txGasAccountIndex int64, sysConfigModel sysconfig.SysconfigModel) error {
	gasAccountIndexConfig, err := sysConfigModel.GetSysconfigByName(sysconfigName.GasAccountIndex)
	if err != nil {
		logx.Errorf("fail to get config: %s, err: %s", sysconfigName.GasAccountIndex, err.Error())
		return errorcode.RpcErrInternal
	}
	gasAccountIndex, err := strconv.ParseInt(gasAccountIndexConfig.Value, 10, 64)
	if err != nil {
		logx.Errorf("cannot parse int :%s, err: %s", gasAccountIndexConfig.Value, err.Error())
		return errorcode.RpcErrInternal
	}
	if gasAccountIndex != txGasAccountIndex {
		logx.Errorf("invalid gas account index, expected: %d, actual: %d", gasAccountIndex, txGasAccountIndex)
		return errorcode.RpcErrInvalidTxField.RefineError("invalid GasAccountIndex")
	}
	return nil
}

func ComputeL2TxTxHash(txInfo string) string {
	hFunc := mimc.NewMiMC()
	hFunc.Write([]byte(txInfo))
	return base64.StdEncoding.EncodeToString(hFunc.Sum(nil))
}

func ConstructMempoolTx(
	txType int64,
	gasFeeAssetId int64,
	gasFeeAssetAmount string,
	nftIndex int64,
	pairIndex int64,
	assetId int64,
	txAmount string,
	toAddress string,
	txInfo string,
	memo string,
	accountIndex int64,
	nonce int64,
	expiredAt int64,
	txDetails []*mempool.MempoolTxDetail,
) (txId string, mempoolTx *mempool.MempoolTx) {
	txId = ComputeL2TxTxHash(txInfo)
	return txId, &mempool.MempoolTx{
		TxHash:         txId,
		TxType:         txType,
		GasFeeAssetId:  gasFeeAssetId,
		GasFee:         gasFeeAssetAmount,
		NftIndex:       nftIndex,
		PairIndex:      pairIndex,
		AssetId:        assetId,
		TxAmount:       txAmount,
		NativeAddress:  toAddress,
		MempoolDetails: txDetails,
		TxInfo:         txInfo,
		ExtraInfo:      "",
		Memo:           memo,
		AccountIndex:   accountIndex,
		Nonce:          nonce,
		ExpiredAt:      expiredAt,
		L2BlockHeight:  commonConstant.NilBlockHeight,
		Status:         mempool.PendingTxStatus,
	}
}

func CreateMempoolTx(
	nMempoolTx *mempool.MempoolTx,
	redisConnection *redis.Redis,
	mempoolModel mempool.MempoolModel,
) (err error) {
	var keys []string
	for _, mempoolTxDetail := range nMempoolTx.MempoolDetails {
		keys = append(keys, util.GetAccountKey(mempoolTxDetail.AccountIndex))
	}
	_, err = redisConnection.Del(keys...)
	if err != nil {
		logx.Errorf("fail to delete keys from redis: %s", err.Error())
		return err
	}
	// write into mempool
	err = mempoolModel.CreateBatchedMempoolTxs([]*mempool.MempoolTx{nMempoolTx})
	return err
}

func CreateFailTx(failTxModel tx.FailTxModel, txType int, txInfo interface{}, error error) error {
	txHash := util.RandomUUID()
	nativeAddress := "0x00"
	txMarshaled, err := json.Marshal(txInfo)
	if err != nil {
		logx.Errorf("unable to marshal, error: %s", err.Error())
		return err
	}
	// write into fail tx
	failTx := &tx.FailTx{
		// transaction id, is primary key
		TxHash: txHash,
		// transaction type
		TxType: int64(txType),
		// tx status, 1 - success(default), 2 - failure
		TxStatus: tx.StatusFail,
		// l1asset id
		AssetAId: commonConstant.NilAssetId,
		// AssetBId
		AssetBId: commonConstant.NilAssetId,
		// tx amount
		TxAmount: commonConstant.NilAssetAmountStr,
		// layer1 address
		NativeAddress: nativeAddress,
		// tx proof
		TxInfo: string(txMarshaled),
		// extra info, if tx fails, show the error info
		ExtraInfo: error.Error(),
		// native memo info
		Memo: "",
	}

	err = failTxModel.CreateFailTx(failTx)
	return err
}
