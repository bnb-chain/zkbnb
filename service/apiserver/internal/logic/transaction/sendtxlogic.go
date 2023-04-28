package transaction

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bnb-chain/zkbnb-crypto/wasm/txtypes"
	common2 "github.com/bnb-chain/zkbnb/common"
	"github.com/bnb-chain/zkbnb/common/redislock"
	"github.com/bnb-chain/zkbnb/core/executor"
	"github.com/bnb-chain/zkbnb/dao/dbcache"
	"github.com/bnb-chain/zkbnb/dao/nft"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/permctrl"
	"gorm.io/gorm"
	"net/http"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb/core"
	"github.com/bnb-chain/zkbnb/dao/tx"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/types"
	types2 "github.com/bnb-chain/zkbnb/types"
)

type SendTxLogic struct {
	logx.Logger
	ctx               context.Context
	svcCtx            *svc.ServiceContext
	permissionControl *permctrl.PermissionControl
	header            http.Header
}

func NewSendTxLogic(ctx context.Context, svcCtx *svc.ServiceContext, header http.Header) *SendTxLogic {
	permissionControl := permctrl.NewPermissionControl(svcCtx)
	return &SendTxLogic{
		Logger:            logx.WithContext(ctx),
		ctx:               ctx,
		svcCtx:            svcCtx,
		permissionControl: permissionControl,
		header:            header,
	}
}

func (s *SendTxLogic) SendTx(req *types.ReqSendTx) (resp *types.TxHash, err error) {
	pendingTxCount, err := s.svcCtx.MemCache.GetTxPendingCountKeyPrefix(func() (interface{}, error) {
		txStatuses := []int64{tx.StatusPending}
		return s.svcCtx.TxPoolModel.GetTxsTotalCount(tx.GetTxWithStatuses(txStatuses))
	})
	if err != nil {
		return nil, types2.AppErrInternal
	}

	channelName := s.header.Get("Channel-Name")
	if len(channelName) > 64 {
		return nil, types2.AppErrChannelNameTooHigh.RefineError(64)
	}
	if s.svcCtx.Config.TxPool.MaxPendingTxCount > 0 && pendingTxCount >= int64(s.svcCtx.Config.TxPool.MaxPendingTxCount) {
		return nil, types2.AppErrTooManyTxs
	}
	// Control the permission list with the whitelist or blacklist
	err = s.permissionControl.Control(req.TxType, req.TxInfo)
	if err != nil {
		return nil, err
	}

	resp = &types.TxHash{}
	bc, err := core.NewBlockChainForDryRun(s.svcCtx.AccountModel, s.svcCtx.NftModel, s.svcCtx.TxPoolModel,
		s.svcCtx.AssetModel, s.svcCtx.SysConfigModel, s.svcCtx.RedisCache, s.svcCtx.MemCache.GetCache())
	if err != nil {
		logx.Error("fail to init blockchain runner:", err)
		return nil, types2.AppErrInternal
	}
	newPoolTx := tx.BaseTx{
		TxHash: types2.EmptyTxHash, // Would be computed in prepare method of executors.
		TxType: int64(req.TxType),
		TxInfo: req.TxInfo,

		GasFeeAssetId: types2.NilAssetId,
		GasFee:        types2.NilAssetAmount,
		NftIndex:      types2.NilNftIndex,
		CollectionId:  types2.NilCollectionNonce,
		AssetId:       types2.NilAssetId,
		TxAmount:      types2.NilAssetAmount,
		NativeAddress: types2.EmptyL1Address,

		ChannelName: channelName,
		BlockHeight: types2.NilBlockHeight,
		TxStatus:    tx.StatusPending,
	}
	newTx := &tx.Tx{BaseTx: newPoolTx}
	executor, err := executor.NewTxExecutor(bc, newTx)
	if err != nil {
		return resp, err
	}
	accountIndex := executor.GetTxInfo().GetAccountIndex()
	s.svcCtx.ChannelTxMetric.WithLabelValues(channelName, strconv.Itoa(executor.GetTxInfo().GetTxType())).Inc()
	nonce := executor.GetTxInfo().GetNonce()
	lock := redislock.GetRedisLock(s.svcCtx.RedisConn, "apiserver:senttx:"+strconv.FormatInt(accountIndex, 10)+"_"+strconv.FormatInt(nonce, 10), 30)
	ok, err := lock.Acquire()
	if err != nil {
		return resp, err
	}
	if !ok {
		logx.Infof(" the apiserversenttx lock has been used, re-try later, accountIndex=%d,nonce=%d", accountIndex, nonce)
		return resp, types2.AppErrInvalidNonce
	}
	defer lock.Release()

	err = bc.ApplyTransaction(newTx)
	if err != nil {
		return resp, err
	}
	newTx.BaseTx.TxType = int64(req.TxType)
	newTx.BaseTx.TxInfo = req.TxInfo
	newTx.BaseTx.BlockHeight = types2.NilBlockHeight
	newTx.BaseTx.TxStatus = tx.StatusPending
	if newTx.BaseTx.TxType == types2.TxTypeMintNft {
		newTx.BaseTx.NftIndex = types2.NilNftIndex
	}
	if newTx.BaseTx.TxType == types2.TxTypeCreateCollection {
		newTx.BaseTx.CollectionId = types2.NilCollectionNonce
	}
	if newTx.BaseTx.TxType == types2.TxTypeMintNft {
		txInfo, _ := types2.ParseMintNftTxInfo(req.TxInfo)
		cid, err := sendToIpfs(txInfo, newTx.BaseTx.TxHash)
		if err != nil {
			return resp, err
		}
		history := &nft.L2NftMetadataHistory{
			Nonce:    0,
			TxHash:   newTx.BaseTx.TxHash,
			NftIndex: types2.NilNftIndex,
			IpfsCid:  cid,
			IpnsName: txInfo.IpnsName,
			IpnsId:   txInfo.IpnsId,
			Mutable:  txInfo.MutableAttributes,
			Metadata: txInfo.MetaData,
			Status:   nft.StatusPending,
		}
		b, err := json.Marshal(txInfo)
		if err != nil {
			return resp, err
		}
		newTx.BaseTx.TxInfo = string(b)
		err = s.svcCtx.DB.Transaction(func(db *gorm.DB) error {
			err = s.svcCtx.NftMetadataHistoryModel.CreateL2NftMetadataHistoryInTransact(db, history)
			if err != nil {
				return err
			}
			err = s.svcCtx.TxPoolModel.CreateTxsInTransact(db, []*tx.PoolTx{{BaseTx: newTx.BaseTx}})
			return err
		})
		if err != nil {
			logx.Errorf("fail to create pool tx: %v, err: %s", newTx, err.Error())
			return resp, types2.AppErrInternal
		}
	} else {
		if err := s.svcCtx.TxPoolModel.CreateTxs([]*tx.PoolTx{{BaseTx: newTx.BaseTx}}); err != nil {
			logx.Errorf("fail to create pool tx: %v, err: %s", newTx, err.Error())
			return resp, types2.AppErrInternal
		}
	}
	s.svcCtx.RedisCache.Set(context.Background(), dbcache.AccountNonceKeyByIndex(newTx.AccountIndex), newTx.Nonce)
	resp.TxHash = newTx.TxHash
	return resp, nil
}

func sendToIpfs(txInfo *txtypes.MintNftTxInfo, txHash string) (string, error) {
	ipnsId, err := common2.Ipfs.GenerateIPNS(txHash)
	if err != nil {
		return "", err
	}
	cid, err := uploadIpfs(txInfo.MetaData, fmt.Sprintf("%s%s", "ipns://", ipnsId.Id))
	if err != nil {
		return "", err
	}
	hash, err := common2.Ipfs.GenerateHash(cid)
	if err != nil {
		return "", err
	}
	txInfo.NftContentHash = hash
	txInfo.IpnsName = txHash
	txInfo.IpnsId = ipnsId.Id
	return cid, nil
}

func uploadIpfs(metaData string, mutableAttributes string) (string, error) {
	meta := ""
	if len(metaData) > 0 {
		content := metaData[len(metaData)-1:]
		if content == "}" {
			metaData = metaData[:len(metaData)-1]
		}
		meta = fmt.Sprintf("%s,\"%s\":\"%s\"}", metaData, "mutable_attributes", mutableAttributes)
	} else {
		meta = fmt.Sprintf("{\"%s\":\"%s\"}", "mutable_attributes", mutableAttributes)
	}
	cid, err := common2.Ipfs.Upload(meta)
	if err != nil {
		return "", err
	}
	return cid, nil
}
