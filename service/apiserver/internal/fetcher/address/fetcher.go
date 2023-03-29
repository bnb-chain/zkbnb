package address

import (
	"context"
	"fmt"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/types"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type Fetcher struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFetcher(ctx context.Context, svcCtx *svc.ServiceContext) *Fetcher {
	return &Fetcher{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (f *Fetcher) GetL1AddressByTx(TxType uint32, TxInfo string) ([]string, error) {

	var l1Address0 = ""
	var l1Address1 = ""
	var err error
	if types.TxTypeWithdraw == TxType {
		l1Address0, err = f.fetcherForWithdrawal(TxInfo)
	} else if types.TxTypeTransfer == TxType {
		l1Address0, err = f.fetcherForTransfer(TxInfo)
	} else if types.TxTypeCreateCollection == TxType {
		l1Address0, err = f.fetcherForCreateCollection(TxInfo)
	} else if types.TxTypeMintNft == TxType {
		l1Address0, err = f.fetcherForMintNft(TxInfo)
	} else if types.TxTypeTransferNft == TxType {
		l1Address0, err = f.fetcherForTransferNft(TxInfo)
	} else if types.TxTypeWithdrawNft == TxType {
		l1Address0, err = f.fetcherForWithdrawalNft(TxInfo)
	} else if types.TxTypeCancelOffer == TxType {
		l1Address0, err = f.fetcherForCancelOffer(TxInfo)
	} else if types.TxTypeAtomicMatch == TxType {
		l1Address0, l1Address1, err = f.fetcherForAtomicMatch(TxInfo)
	} else if types.TxTypeUpdateNFT == TxType {
		l1Address0, err = f.fetcherForUpdateNft(TxInfo)
	} else if types.TxTypeChangePubKey == TxType {
		l1Address0, err = f.fetcherForChangePubKey(TxInfo)
	} else {
		return nil, errors.New(fmt.Sprintf("Can not find Fetcher Function for TxType:%d", TxType))
	}

	if err != nil {
		return nil, err
	}

	if len(l1Address1) > 0 {
		return []string{l1Address0, l1Address1}, nil
	} else {
		return []string{l1Address0}, nil
	}
}

func (f *Fetcher) fetcherForWithdrawal(txInfo string) (string, error) {
	transaction, err := types.ParseWithdrawTxInfo(txInfo)
	if err != nil {
		logx.Errorf("parse withdrawal tx failed: %s", err.Error())
		return "", types.AppErrInvalidTxInfo
	}
	return f.fetchL1AddressByAccountIndex(transaction.FromAccountIndex)
}

func (f *Fetcher) fetcherForTransfer(txInfo string) (string, error) {
	transaction, err := types.ParseTransferTxInfo(txInfo)
	if err != nil {
		logx.Errorf("parse transfer tx failed: %s", err.Error())
		return "", types.AppErrInvalidTxInfo
	}
	return f.fetchL1AddressByAccountIndex(transaction.FromAccountIndex)
}

func (f *Fetcher) fetcherForCreateCollection(txInfo string) (string, error) {
	transaction, err := types.ParseCreateCollectionTxInfo(txInfo)
	if err != nil {
		logx.Errorf("parse create collection tx failed: %s", err.Error())
		return "", types.AppErrInvalidTxInfo
	}
	return f.fetchL1AddressByAccountIndex(transaction.AccountIndex)
}

func (f *Fetcher) fetcherForMintNft(txInfo string) (string, error) {
	transaction, err := types.ParseMintNftTxInfo(txInfo)
	if err != nil {
		logx.Errorf("parse mint nft tx failed: %s", err.Error())
		return "", types.AppErrInvalidTxInfo
	}
	return f.fetchL1AddressByAccountIndex(transaction.CreatorAccountIndex)
}

func (f *Fetcher) fetcherForTransferNft(txInfo string) (string, error) {
	transaction, err := types.ParseTransferNftTxInfo(txInfo)
	if err != nil {
		logx.Errorf("parse cancel offer tx failed: %s", err.Error())
		return "", types.AppErrInvalidTxInfo
	}
	return f.fetchL1AddressByAccountIndex(transaction.FromAccountIndex)
}

func (f *Fetcher) fetcherForWithdrawalNft(txInfo string) (string, error) {
	transaction, err := types.ParseWithdrawNftTxInfo(txInfo)
	if err != nil {
		logx.Errorf("parse withdrawal nft tx failed: %s", err.Error())
		return "", types.AppErrInvalidTxInfo
	}
	return f.fetchL1AddressByAccountIndex(transaction.AccountIndex)
}

func (f *Fetcher) fetcherForCancelOffer(txInfo string) (string, error) {
	transaction, err := types.ParseCancelOfferTxInfo(txInfo)
	if err != nil {
		logx.Errorf("parse cancel offer tx failed: %s", err.Error())
		return "", types.AppErrInvalidTxInfo
	}
	return f.fetchL1AddressByAccountIndex(transaction.AccountIndex)
}

func (f *Fetcher) fetcherForAtomicMatch(txInfo string) (string, string, error) {
	transaction, err := types.ParseAtomicMatchTxInfo(txInfo)
	if err != nil {
		logx.Errorf("parse atomic match tx failed: %s", err.Error())
		return "", "", types.AppErrInvalidTxInfo
	}
	sellOfferL1Address, err := f.fetchL1AddressByAccountIndex(transaction.SellOffer.AccountIndex)
	if err != nil {
		return "", "", err
	}
	buyOfferL1Address, err := f.fetchL1AddressByAccountIndex(transaction.BuyOffer.AccountIndex)
	if err != nil {
		return "", "", err
	}
	return sellOfferL1Address, buyOfferL1Address, nil
}

func (f *Fetcher) fetcherForUpdateNft(txInfo string) (string, error) {
	tx, err := types.ParseUpdateNftTxInfo(txInfo)
	if err != nil {
		return "", err
	}
	return f.fetchL1AddressByAccountIndex(tx.AccountIndex)
}

func (f *Fetcher) fetcherForChangePubKey(txInfo string) (string, error) {
	transaction, err := types.ParseChangePubKeyTxInfo(txInfo)
	if err != nil {
		logx.Errorf("parse ChangePubKey tx failed: %s", err.Error())
		return "", types.AppErrInvalidTxInfo
	}
	return f.fetchL1AddressByAccountIndex(transaction.AccountIndex)
}

func (f *Fetcher) fetchL1AddressByAccountIndex(accountIndex int64) (string, error) {
	account, err := f.svcCtx.MemCache.GetAccountWithFallback(accountIndex, func() (interface{}, error) {
		return f.svcCtx.AccountModel.GetAccountByIndex(accountIndex)
	})
	if err != nil {
		if err == types.DbErrNotFound {
			return "", types.AppErrAccountNotFound
		}
		return "", err
	}
	return account.L1Address, nil
}
