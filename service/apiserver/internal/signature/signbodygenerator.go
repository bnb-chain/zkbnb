package signature

import (
	"github.com/bnb-chain/zkbnb-crypto/wasm/txtypes"
	"github.com/bnb-chain/zkbnb/types"
)

func GenerateSignatureBody(txType uint32, txInfo string) (txtypes.TxInfo, error) {
	switch txType {
	case types.TxTypeChangePubKey:
		return types.ParseChangePubKeyTxInfo(txInfo)
	case types.TxTypeTransfer:
		return types.ParseTransferTxInfo(txInfo)
	case types.TxTypeWithdraw:
		return types.ParseWithdrawTxInfo(txInfo)
	case types.TxTypeCreateCollection:
		return types.ParseCreateCollectionTxInfo(txInfo)
	case types.TxTypeMintNft:
		return types.ParseMintNftTxInfo(txInfo)
	case types.TxTypeTransferNft:
		return types.ParseTransferNftTxInfo(txInfo)
	case types.TxTypeOffer:
		return types.ParseOfferTxInfo(txInfo)
	case types.TxTypeCancelOffer:
		return types.ParseCancelOfferTxInfo(txInfo)
	case types.TxTypeWithdrawNft:
		return types.ParseWithdrawNftTxInfo(txInfo)
	case types.TxTypeUpdateNFT:
		return types.ParseUpdateNftTxInfo(txInfo)
	}
	return nil, types.AppErrUnsupportedTxType
}
