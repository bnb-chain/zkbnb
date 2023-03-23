package signature

import (
	"context"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/fetcher/address"
	"github.com/bnb-chain/zkbnb/service/apiserver/internal/svc"
	"github.com/bnb-chain/zkbnb/types"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

type VerifySignature struct {
	fetcher *address.Fetcher
}

func NewVerifySignature(ctx context.Context, svcCtx *svc.ServiceContext) *VerifySignature {
	fetcher := address.NewFetcher(ctx, svcCtx)
	return &VerifySignature{
		fetcher: fetcher,
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
	originAddressArray, err := v.fetcher.GetL1AddressByTx(TxType, TxInfo)
	if err != nil {
		return err
	}
	// For normal transactions other than atomic match transaction
	// There is only one signature address to be verified.
	// While for the atomic match transaction, there should be two
	// signature addresses to be verified, one is for the sell offer,
	// and other is for the buy offer.
	if len(originAddressArray) == 1 {
		originAddress := common.HexToAddress(originAddressArray[0])

		//Compare the original address and the public address to verify the identifier
		if publicAddress != originAddress {
			return types.AppErrTxSignatureError
		}
	} else if len(originAddressArray) == 2 {
		originAddress0 := common.HexToAddress(originAddressArray[0])
		originAddress1 := common.HexToAddress(originAddressArray[1])

		//Compare both the addresses for buy offer and sell offer to verify the identifier.
		if publicAddress != originAddress0 && publicAddress != originAddress1 {
			return types.AppErrTxSignatureError
		}
	}
	return nil
}
