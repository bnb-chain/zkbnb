package sendrawtx

import (
	"encoding/json"
	"errors"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/common/commonAsset"
	"github.com/bnb-chain/zkbas/common/commonConstant"
	"github.com/bnb-chain/zkbas/common/errorcode"
	"github.com/bnb-chain/zkbas/common/model/mempool"
	"github.com/bnb-chain/zkbas/common/model/sysConfig"
	"github.com/bnb-chain/zkbas/common/model/tx"
	"github.com/bnb-chain/zkbas/common/sysConfigName"
	"github.com/bnb-chain/zkbas/service/api/app/internal/fetcher/state"
)

type GasChecker interface {
	CheckGas(account *commonAsset.AccountInfo, txGasAccountIndex int64, txGasAssetId int64, txGasAssetAmount *big.Int) error
}

type gasChecker struct {
	gasAccountIndex int64
}

func NewGasChecker(sysConfigModel sysConfig.SysConfigModel) *gasChecker {
	gasAccountIndexConfig, err := sysConfigModel.GetSysConfigByName(sysConfigName.GasAccountIndex)
	if err != nil {
		logx.Errorf("fail to get config: %s, err: %s", sysConfigName.GasAccountIndex, err.Error())
		panic("GasAccountIndex is not configured")
	}
	gasAccountIndex, err := strconv.ParseInt(gasAccountIndexConfig.Value, 10, 64)
	if err != nil {
		logx.Errorf("cannot parse int :%s, err: %s", gasAccountIndexConfig.Value, err.Error())
		panic("GasAccountIndex is not properly configured")
	}
	return &gasChecker{gasAccountIndex: gasAccountIndex}
}

func (c *gasChecker) CheckGas(account *commonAsset.AccountInfo, txGasAccountIndex int64, txGasAssetId int64, txGasAssetAmount *big.Int) error {
	if c.gasAccountIndex != txGasAccountIndex {
		logx.Errorf("invalid gas account index, expected: %d, actual: %d", c.gasAccountIndex, txGasAccountIndex)
		return errors.New("invalid GasAccountIndex")
	}
	asset, ok := account.AssetInfo[txGasAssetId]
	if !ok || asset.Balance.Cmp(txGasAssetAmount) < 0 {
		logx.Errorf("insufficient balance of gas asset")
		return errors.New("insufficient balance of gas asset")
	}
	return nil
}

type NonceChecker interface {
	CheckNonce(accountIndex int64, txNonce int64) error
}

type nonceChecker struct {
	stateFetcher state.Fetcher
}

func NewNonceChecker(fetcher state.Fetcher) *nonceChecker {
	return &nonceChecker{stateFetcher: fetcher}
}

func (c *nonceChecker) CheckNonce(accountIndex int64, txNonce int64) error {
	pendingNonce, err := c.stateFetcher.GetPendingNonce(accountIndex)
	if err != nil {
		return errors.New("cannot verify Nonce")
	}
	if pendingNonce != txNonce {
		return errors.New("invalid Nonce")
	}
	return nil
}

type txHasher func(txInfo interface{}) ([]byte, error)

type MempoolTxSender interface {
	SendMempoolTx(hasher txHasher, txInfo interface{}, mempoolTx *mempool.MempoolTx) (string, error)
}

type mempoolTxSender struct {
	mempoolTxModel mempool.MempoolModel
	failTxModel    tx.FailTxModel
}

func NewMempoolTxSender(mempoolTxModel mempool.MempoolModel,
	failTxModel tx.FailTxModel) *mempoolTxSender {
	return &mempoolTxSender{
		mempoolTxModel: mempoolTxModel,
		failTxModel:    failTxModel,
	}
}

func (s mempoolTxSender) SendMempoolTx(hasher txHasher, txInfo interface{}, mempoolTx *mempool.MempoolTx) (string, error) {
	// generate tx id
	hash, err := hasher(txInfo)
	if err != nil {
		return "", errorcode.AppErrInternal
	}
	txId := common.Bytes2Hex(hash)
	mempoolTx.TxHash = txId

	// write into mempool
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		logx.Errorf("unable to marshal tx, err: %s", err.Error())
		return "", errorcode.AppErrInternal
	}
	mempoolTx.TxInfo = string(txInfoBytes)

	mempoolTx.L2BlockHeight = commonConstant.NilBlockHeight
	mempoolTx.Status = mempool.PendingTxStatus

	if err := s.mempoolTxModel.CreateBatchedMempoolTxs([]*mempool.MempoolTx{mempoolTx}); err != nil {
		logx.Errorf("fail to create mempool tx: %v, err: %s", mempoolTx, err.Error())

		failTx := &tx.FailTx{
			TxHash:    txId,
			TxType:    mempoolTx.TxType,
			TxStatus:  tx.StatusFail,
			AssetAId:  commonConstant.NilAssetId,
			AssetBId:  commonConstant.NilAssetId,
			TxAmount:  commonConstant.NilAssetAmountStr,
			TxInfo:    string(txInfoBytes),
			ExtraInfo: err.Error(),
			Memo:      "",
		}
		_ = s.failTxModel.CreateFailTx(failTx)
		return "", errorcode.AppErrInternal
	}
	return txId, nil
}

type TxSender interface {
	SendTx(rawTx string) (string, error)
}
