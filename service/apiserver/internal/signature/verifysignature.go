package signature

import (
	"context"
	"fmt"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/types"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
	"strconv"
)

type VerifySignature struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVerifySignature(ctx context.Context, svcCtx *svc.ServiceContext) *VerifySignature {
	return &VerifySignature{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (v *VerifySignature) VerifySignatureInfo(TxType uint32, TxInfo, Signature string) error {

	// For compatibility consideration, if signature string is empty, directly ignore the validation
	if len(Signature) == 0 {
		return nil
	}

	//Generate the signature body data from the transaction type and transaction info
	signatureBody, err := GenerateSignatureBody(TxType, TxInfo)
	if err != nil {
		return err
	}
	message := accounts.TextHash([]byte(signatureBody))

	//Decode from signature string to get the signature byte array
	signatureContent, err := hexutil.Decode(Signature)
	if err != nil {
		return err
	}
	signatureContent[64] -= 27 // Transform yellow paper V from 27/28 to 0/1

	//Calculate the public key from the signature and source string
	signaturePublicKey, err := crypto.SigToPub(message, signatureContent)
	if err != nil {
		return err
	}

	//Calculate the address from the public key
	publicAddress := crypto.PubkeyToAddress(*signaturePublicKey)

	//Query the origin address from the database
	originAddressStr, err := v.GetL1AddressByTx(TxType, TxInfo)
	if err != nil {
		return err
	}
	originAddress := common.HexToAddress(originAddressStr)

	//Compare the original address and the public address to verify the identifier
	if publicAddress != originAddress {
		return errors.New("Tx Signature Error")
	}
	return nil
}

func (v *VerifySignature) GetL1AddressByTx(TxType uint32, TxInfo string) (string, error) {

	var l1Address string
	var err error

	if types.TxTypeWithdraw == TxType {
		l1Address, err = v.fetcherForWithdrawal(TxInfo)
	} else if types.TxTypeTransfer == TxType {
		l1Address, err = v.fetcherForTransfer(TxInfo)
	} else if types.TxTypeCreateCollection == TxType {
		l1Address, err = v.fetcherForCreateCollection(TxInfo)
	} else if types.TxTypeMintNft == TxType {
		l1Address, err = v.fetcherForMintNft(TxInfo)
	} else if types.TxTypeTransferNft == TxType {
		l1Address, err = v.fetcherForTransferNft(TxInfo)
	} else if types.TxTypeWithdrawNft == TxType {
		l1Address, err = v.fetcherForWithdrawalNft(TxInfo)
	} else if types.TxTypeCancelOffer == TxType {
		l1Address, err = v.fetcherForCancelOffer(TxInfo)
	} else if types.TxTypeAtomicMatch == TxType {
		l1Address, err = v.fetcherForAtomicMatch(TxInfo)
	} else if types.TxTypeEmpty == TxType {
		l1Address, err = v.fetcherForAccount(TxInfo)
	} else {
		return "", errors.New(fmt.Sprintf("Can not find Fetcher Function for TxType:%d", TxType))
	}

	if err != nil {
		return "", err
	}
	return l1Address, nil
}

func (v *VerifySignature) fetcherForWithdrawal(txInfo string) (string, error) {
	transaction, err := types.ParseWithdrawTxInfo(txInfo)
	if err != nil {
		logx.Errorf("parse withdrawal tx failed: %s", err.Error())
		return "", errors.New("invalid tx info")
	}
	return v.fetchL1AddressByAccountIndex(transaction.FromAccountIndex)
}

func (v *VerifySignature) fetcherForTransfer(txInfo string) (string, error) {
	transaction, err := types.ParseTransferTxInfo(txInfo)
	if err != nil {
		logx.Errorf("parse transfer tx failed: %s", err.Error())
		return "", errors.New("invalid tx info")
	}
	return v.fetchL1AddressByAccountIndex(transaction.FromAccountIndex)
}

func (v *VerifySignature) fetcherForCreateCollection(txInfo string) (string, error) {
	transaction, err := types.ParseCreateCollectionTxInfo(txInfo)
	if err != nil {
		logx.Errorf("parse create collection tx failed: %s", err.Error())
		return "", errors.New("invalid tx info")
	}
	return v.fetchL1AddressByAccountIndex(transaction.AccountIndex)
}

func (v *VerifySignature) fetcherForMintNft(txInfo string) (string, error) {
	transaction, err := types.ParseMintNftTxInfo(txInfo)
	if err != nil {
		logx.Errorf("parse mint nft tx failed: %s", err.Error())
		return "", errors.New("invalid tx info")
	}
	return v.fetchL1AddressByAccountIndex(transaction.CreatorAccountIndex)
}

func (v *VerifySignature) fetcherForTransferNft(txInfo string) (string, error) {
	transaction, err := types.ParseTransferNftTxInfo(txInfo)
	if err != nil {
		logx.Errorf("parse cancel offer tx failed: %s", err.Error())
		return "", errors.New("invalid tx info")
	}
	return v.fetchL1AddressByAccountIndex(transaction.FromAccountIndex)
}

func (v *VerifySignature) fetcherForWithdrawalNft(txInfo string) (string, error) {
	transaction, err := types.ParseWithdrawNftTxInfo(txInfo)
	if err != nil {
		logx.Errorf("parse withdrawal nft tx failed: %s", err.Error())
		return "", errors.New("invalid tx info")
	}
	return v.fetchL1AddressByAccountIndex(transaction.AccountIndex)
}

func (v *VerifySignature) fetcherForCancelOffer(txInfo string) (string, error) {
	transaction, err := types.ParseCancelOfferTxInfo(txInfo)
	if err != nil {
		logx.Errorf("parse cancel offer tx failed: %s", err.Error())
		return "", errors.New("invalid tx info")
	}
	return v.fetchL1AddressByAccountIndex(transaction.AccountIndex)
}

func (v *VerifySignature) fetcherForAtomicMatch(txInfo string) (string, error) {
	transaction, err := types.ParseAtomicMatchTxInfo(txInfo)
	if err != nil {
		logx.Errorf("parse atomic match tx failed: %s", err.Error())
		return "", errors.New("invalid tx info")
	}
	return v.fetchL1AddressByAccountIndex(transaction.AccountIndex)
}

func (v *VerifySignature) fetcherForAccount(txInfo string) (string, error) {
	accountIndex, err := strconv.ParseInt(txInfo, 10, 64)
	if err != nil {
		logx.Errorf("parse atomic match int64 failed: %s", err.Error())
		return "", errors.New("invalid tx info")
	}
	return v.fetchL1AddressByAccountIndex(accountIndex)
}

func (v *VerifySignature) fetchL1AddressByAccountIndex(accountIndex int64) (string, error) {
	account, err := v.svcCtx.MemCache.GetAccountWithFallback(accountIndex, func() (interface{}, error) {
		return v.svcCtx.AccountModel.GetAccountByIndex(accountIndex)
	})
	if err != nil {
		return "", err
	}
	return account.L1Address, nil
}
