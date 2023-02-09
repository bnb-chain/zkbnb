package signature

import (
	"context"
	"fmt"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/types"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type L1AddressFetcher struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewL1AddressFetcher(ctx context.Context, svcCtx *svc.ServiceContext) *L1AddressFetcher {
	return &L1AddressFetcher{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (f *L1AddressFetcher) GetL1AddressByTx(TxType uint32, TxInfo string) (string, error) {

	var l1Address string
	var err error

	if types.TxTypeWithdraw == TxType {
		l1Address, err = f.fetcherForWithdrawal(TxInfo)
	} else if types.TxTypeTransfer == TxType {
		l1Address, err = f.fetcherForTransfer(TxInfo)
	} else if types.TxTypeCreateCollection == TxType {
		l1Address, err = f.fetcherForCreateCollection(TxInfo)
	} else if types.TxTypeMintNft == TxType {
		l1Address, err = f.fetcherForMintNft(TxInfo)
	} else if types.TxTypeTransferNft == TxType {
		l1Address, err = f.fetcherForTransferNft(TxInfo)
	} else if types.TxTypeWithdrawNft == TxType {
		l1Address, err = f.fetcherForWithdrawalNft(TxInfo)
	} else if types.TxTypeCancelOffer == TxType {
		l1Address, err = f.fetcherForCancelOffer(TxInfo)
	} else if types.TxTypeAtomicMatch == TxType {
		l1Address, err = f.fetcherForAtomicMatch(TxInfo)
	} else {
		return "", errors.New(fmt.Sprintf("Can not find Fetcher Function for TxType:%d", TxType))
	}

	if err != nil {
		return "", err
	}
	return l1Address, nil
}

func (f *L1AddressFetcher) fetcherForWithdrawal(txInfo string) (string, error) {
	transaction, err := types.ParseWithdrawTxInfo(txInfo)
	if err != nil {
		logx.Errorf("parse withdrawal tx failed: %s", err.Error())
		return "", errors.New("invalid tx info")
	}
	return f.fetchL1AddressByAccountIndex(transaction.FromAccountIndex)
}

func (f *L1AddressFetcher) fetcherForTransfer(txInfo string) (string, error) {
	transaction, err := types.ParseTransferTxInfo(txInfo)
	if err != nil {
		logx.Errorf("parse transfer tx failed: %s", err.Error())
		return "", errors.New("invalid tx info")
	}
	return f.fetchL1AddressByAccountIndex(transaction.FromAccountIndex)
}

func (f *L1AddressFetcher) fetcherForCreateCollection(txInfo string) (string, error) {
	transaction, err := types.ParseCreateCollectionTxInfo(txInfo)
	if err != nil {
		logx.Errorf("parse create collection tx failed: %s", err.Error())
		return "", errors.New("invalid tx info")
	}
	return f.fetchL1AddressByAccountIndex(transaction.AccountIndex)
}

func (f *L1AddressFetcher) fetcherForMintNft(txInfo string) (string, error) {
	transaction, err := types.ParseMintNftTxInfo(txInfo)
	if err != nil {
		logx.Errorf("parse mint nft tx failed: %s", err.Error())
		return "", errors.New("invalid tx info")
	}
	return f.fetchL1AddressByAccountIndex(transaction.CreatorAccountIndex)
}

func (f *L1AddressFetcher) fetcherForTransferNft(txInfo string) (string, error) {
	transaction, err := types.ParseTransferNftTxInfo(txInfo)
	if err != nil {
		logx.Errorf("parse cancel offer tx failed: %s", err.Error())
		return "", errors.New("invalid tx info")
	}
	return f.fetchL1AddressByAccountIndex(transaction.FromAccountIndex)
}

func (f *L1AddressFetcher) fetcherForWithdrawalNft(txInfo string) (string, error) {
	transaction, err := types.ParseWithdrawNftTxInfo(txInfo)
	if err != nil {
		logx.Errorf("parse withdrawal nft tx failed: %s", err.Error())
		return "", errors.New("invalid tx info")
	}
	return f.fetchL1AddressByAccountIndex(transaction.AccountIndex)
}

func (f *L1AddressFetcher) fetcherForCancelOffer(txInfo string) (string, error) {
	transaction, err := types.ParseCancelOfferTxInfo(txInfo)
	if err != nil {
		logx.Errorf("parse cancel offer tx failed: %s", err.Error())
		return "", errors.New("invalid tx info")
	}
	return f.fetchL1AddressByAccountIndex(transaction.AccountIndex)
}

func (f *L1AddressFetcher) fetcherForAtomicMatch(txInfo string) (string, error) {
	transaction, err := types.ParseAtomicMatchTxInfo(txInfo)
	if err != nil {
		logx.Errorf("parse atomic match tx failed: %s", err.Error())
		return "", errors.New("invalid tx info")
	}
	return f.fetchL1AddressByAccountIndex(transaction.AccountIndex)
}

func (f *L1AddressFetcher) fetchL1AddressByAccountIndex(accountIndex int64) (string, error) {
	account, err := f.svcCtx.MemCache.GetAccountWithFallback(accountIndex, func() (interface{}, error) {
		return f.svcCtx.AccountModel.GetAccountByIndex(accountIndex)
	})
	if err != nil {
		return "", err
	}
	return account.L1Address, nil
}
