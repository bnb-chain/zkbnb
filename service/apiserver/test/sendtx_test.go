package test

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	curve "github.com/bnb-chain/zkbas-crypto/ecc/ztwistededwards/tebn254"
	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	common2 "github.com/bnb-chain/zkbas/common"
	"github.com/bnb-chain/zkbas/common/chain"
	"github.com/bnb-chain/zkbas/service/apiserver/internal/types"
	"github.com/bnb-chain/zkbas/tree"
	types2 "github.com/bnb-chain/zkbas/types"
)

const accountName = "sher.legend"
const seed = "28e1a3762ff9944e9a4ad79477b756ef0aff3d2af76f0f40a0c3ec6ca76cf24b"

func (s *AppSuite) TestSendTx() {

	type args struct {
		txType int
		txInfo string
	}
	tests := []struct {
		name     string
		args     args
		httpCode int
	}{
		{"add liquidity", args{types2.TxTypeAddLiquidity, constructAddLiquidityTxInfo(s)}, 200},
		{"remove liquidity", args{types2.TxTypeRemoveLiquidity, constructRemoveLiquidityTxInfo(s)}, 200},
		{"create collection", args{types2.TxTypeCreateCollection, constructCreateCollectionTxInfo(s)}, 200},
		{"mint nft", args{types2.TxTypeMintNft, constructMintNftTxInfo(s)}, 200},
		{"transfer nft", args{types2.TxTypeTransferNft, constructTransferNftTxInfo(s)}, 200},
		{"withdraw nft", args{types2.TxTypeWithdraw, constructSendNftTxInfo(s)}, 200},
		{"cancel offer", args{types2.TxTypeCancelOffer, constructCancelOfferTxInfo(s)}, 200},
		{"atomic match", args{types2.TxTypeAtomicMatch, constructAtomicMatchTxInfo(s)}, 200},
		{"transfer", args{types2.TxTypeTransfer, constructTransferTxInfo(s)}, 200},
		{"swap", args{types2.TxTypeSwap, constructSwapTxInfo(s)}, 200},
		{"withdraw", args{types2.TxTypeWithdraw, constructWithdrawTxInfo(s)}, 200},
		{"offer", args{types2.TxTypeOffer, "invalid"}, 400},
		{"invalid tx type", args{100, "invalid"}, 400},
	}

	for _, tt := range tests {
		httpCode, result := SendTx(s, tt.args.txType, tt.args.txInfo)
		assert.Equal(s.T(), tt.httpCode, httpCode)
		if httpCode == http.StatusOK {
			assert.NotNil(s.T(), result.TxHash)
			fmt.Printf("result: %+v \n", result)
		}
	}

}

func SendTx(s *AppSuite, txType int, txInfo string) (int, *types.TxHash) {
	resp, err := http.PostForm(s.url+"/api/v1/sendTx",
		url.Values{"tx_type": {strconv.Itoa(txType)}, "tx_info": {txInfo}})
	assert.NoError(s.T(), err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	assert.NoError(s.T(), err)
	fmt.Printf("body: %s \n", string(body))

	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, nil
	}
	result := types.TxHash{}
	err = json.Unmarshal(body, &result)
	return resp.StatusCode, &result
}

func getAccountIndex(s *AppSuite, accountName string) int64 {
	httpCode, accountResp := GetAccount(s, "name", accountName)
	if httpCode != http.StatusOK {
		panic("cannot get account: " + accountName)
	}
	return accountResp.Index
}

func getNextNonce(s *AppSuite, accountName string) int64 {
	httpCode, nonceResp := GetNextNonce(s, int(getAccountIndex(s, accountName)))
	if httpCode != http.StatusOK {
		panic("cannot get nonce for account: " + accountName)
	}
	return int64(nonceResp.Nonce)
}

func getNextOfferId(s *AppSuite, accountName string) int64 {
	httpCode, nonceResp := GetMaxOfferId(s, int(getAccountIndex(s, accountName)))
	if httpCode != http.StatusOK {
		panic("cannot get nonce for account: " + accountName)
	}
	return int64(nonceResp.OfferId + 1)
}

func constructAddLiquidityTxInfo(s *AppSuite) string {
	key, err := curve.GenerateEddsaPrivateKey(seed)
	if err != nil {
		panic(err)
	}
	assetAAmount := big.NewInt(100000)
	assetBAmount := big.NewInt(100000)
	lpAmount, err := chain.ComputeEmptyLpAmount(assetAAmount, assetBAmount)
	if err != nil {
		panic(err)
	}
	expiredAt := time.Now().Add(time.Hour * 2).UnixMilli()
	txInfo := &types2.AddLiquidityTxInfo{
		FromAccountIndex:  getAccountIndex(s, accountName),
		PairIndex:         0,
		AssetAId:          0,
		AssetAAmount:      assetAAmount,
		AssetBId:          2,
		AssetBAmount:      assetBAmount,
		LpAmount:          lpAmount,
		GasAccountIndex:   1,
		GasFeeAssetId:     2,
		GasFeeAssetAmount: big.NewInt(5000),
		ExpiredAt:         expiredAt,
		Nonce:             getNextNonce(s, accountName),
	}
	hFunc := mimc.NewMiMC()
	msgHash, err := legendTxTypes.ComputeAddLiquidityMsgHash(txInfo, hFunc)
	if err != nil {
		panic(err)
	}
	hFunc.Reset()
	signature, err := key.Sign(msgHash, hFunc)
	if err != nil {
		panic(err)
	}
	txInfo.Sig = signature
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		panic(err)
	}
	return string(txInfoBytes)
}

func constructAtomicMatchTxInfo(s *AppSuite) string {
	sherSeed := seed
	sherKey, err := curve.GenerateEddsaPrivateKey(sherSeed)
	if err != nil {
		panic(err)
	}
	gavinSeed := "17673b9a9fdec6dc90c7cc1eb1c939134dfb659d2f08edbe071e5c45f343d008"
	gavinKey, err := curve.GenerateEddsaPrivateKey(gavinSeed)
	if err != nil {
		panic(err)
	}
	listedAt := time.Now().UnixMilli()
	expiredAt := time.Now().Add(time.Hour * 2).UnixMilli()
	buyOffer := &types2.OfferTxInfo{
		Type:         types2.BuyOfferType,
		OfferId:      getNextOfferId(s, "gavin.legend"),
		AccountIndex: getAccountIndex(s, "gavin.legend"),
		NftIndex:     1,
		AssetId:      0,
		AssetAmount:  big.NewInt(10000),
		ListedAt:     listedAt,
		ExpiredAt:    expiredAt,
		TreasuryRate: 200,
		Sig:          nil,
	}
	hFunc := mimc.NewMiMC()
	buyHash, err := legendTxTypes.ComputeOfferMsgHash(buyOffer, hFunc)
	if err != nil {
		panic(err)
	}
	hFunc.Reset()
	buySig, err := gavinKey.Sign(buyHash, hFunc)
	if err != nil {
		panic(err)
	}
	buyOffer.Sig = buySig
	sellOffer := &types2.OfferTxInfo{
		Type:         types2.SellOfferType,
		OfferId:      getNextOfferId(s, accountName),
		AccountIndex: getAccountIndex(s, accountName),
		NftIndex:     1,
		AssetId:      0,
		AssetAmount:  big.NewInt(10000),
		ListedAt:     listedAt,
		ExpiredAt:    expiredAt,
		TreasuryRate: 200,
		Sig:          nil,
	}
	hFunc.Reset()
	sellHash, err := legendTxTypes.ComputeOfferMsgHash(sellOffer, hFunc)
	if err != nil {
		panic(err)
	}
	hFunc.Reset()
	sellSig, err := sherKey.Sign(sellHash, hFunc)
	if err != nil {
		panic(err)
	}
	sellOffer.Sig = sellSig
	txInfo := &types2.AtomicMatchTxInfo{
		AccountIndex:      getAccountIndex(s, accountName),
		BuyOffer:          buyOffer,
		SellOffer:         sellOffer,
		GasAccountIndex:   1,
		GasFeeAssetId:     0,
		GasFeeAssetAmount: big.NewInt(5000),
		Nonce:             getNextNonce(s, accountName),
		ExpiredAt:         expiredAt,
		Sig:               nil,
	}
	hFunc.Reset()
	msgHash, err := legendTxTypes.ComputeAtomicMatchMsgHash(txInfo, hFunc)
	if err != nil {
		panic(err)
	}
	hFunc.Reset()
	signature, err := sherKey.Sign(msgHash, hFunc)
	if err != nil {
		panic(err)
	}
	txInfo.Sig = signature
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		panic(err)
	}
	return string(txInfoBytes)
}

func constructCancelOfferTxInfo(s *AppSuite) string {
	key, err := curve.GenerateEddsaPrivateKey(seed)
	if err != nil {
		panic(err)
	}
	expiredAt := time.Now().Add(time.Hour * 2).UnixMilli()
	txInfo := &types2.CancelOfferTxInfo{
		AccountIndex:      getAccountIndex(s, accountName),
		OfferId:           getNextOfferId(s, accountName),
		GasAccountIndex:   1,
		GasFeeAssetId:     2,
		GasFeeAssetAmount: big.NewInt(5000),
		ExpiredAt:         expiredAt,
		Nonce:             getNextNonce(s, accountName),
		Sig:               nil,
	}
	hFunc := mimc.NewMiMC()
	msgHash, err := legendTxTypes.ComputeCancelOfferMsgHash(txInfo, hFunc)
	if err != nil {
		panic(err)
	}
	hFunc.Reset()
	signature, err := key.Sign(msgHash, hFunc)
	if err != nil {
		panic(err)
	}
	txInfo.Sig = signature
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		panic(err)
	}
	return string(txInfoBytes)
}

func constructCreateCollectionTxInfo(s *AppSuite) string {
	key, err := curve.GenerateEddsaPrivateKey(seed)
	if err != nil {
		panic(err)
	}
	expiredAt := time.Now().Add(time.Hour * 2).UnixMilli()
	txInfo := &types2.CreateCollectionTxInfo{
		AccountIndex:      getAccountIndex(s, accountName),
		CollectionId:      0,
		Name:              "Zkbas Collection",
		Introduction:      "Wonderful zkbas!",
		GasAccountIndex:   1,
		GasFeeAssetId:     2,
		GasFeeAssetAmount: big.NewInt(5000),
		ExpiredAt:         expiredAt,
		Nonce:             getNextNonce(s, accountName),
		Sig:               nil,
	}
	hFunc := mimc.NewMiMC()
	msgHash, err := legendTxTypes.ComputeCreateCollectionMsgHash(txInfo, hFunc)
	if err != nil {
		panic(err)
	}
	hFunc.Reset()
	signature, err := key.Sign(msgHash, hFunc)
	if err != nil {
		panic(err)
	}
	txInfo.Sig = signature
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		panic(err)
	}
	return string(txInfoBytes)
}

func constructMintNftTxInfo(s *AppSuite) string {
	key, err := curve.GenerateEddsaPrivateKey(seed)
	if err != nil {
		panic(err)
	}
	nameHash, err := common2.AccountNameHash("gavin.legend")
	if err != nil {
		panic(err)
	}
	hFunc := mimc.NewMiMC()
	hFunc.Write([]byte(common2.RandomUUID()))
	contentHash := hFunc.Sum(nil)
	expiredAt := time.Now().Add(time.Hour * 2).UnixMilli()
	txInfo := &types2.MintNftTxInfo{
		CreatorAccountIndex: getAccountIndex(s, accountName),
		ToAccountIndex:      3,
		ToAccountNameHash:   nameHash,
		NftContentHash:      common.Bytes2Hex(contentHash),
		NftCollectionId:     1,
		CreatorTreasuryRate: 0,
		GasAccountIndex:     1,
		GasFeeAssetId:       2,
		GasFeeAssetAmount:   big.NewInt(5000),
		ExpiredAt:           expiredAt,
		Nonce:               getNextNonce(s, accountName),
		Sig:                 nil,
	}
	hFunc.Reset()
	msgHash, err := legendTxTypes.ComputeMintNftMsgHash(txInfo, hFunc)
	if err != nil {
		panic(err)
	}
	hFunc.Reset()
	signature, err := key.Sign(msgHash, hFunc)
	if err != nil {
		panic(err)
	}
	txInfo.Sig = signature
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		panic(err)
	}
	return string(txInfoBytes)
}

func constructRemoveLiquidityTxInfo(s *AppSuite) string {
	key, err := curve.GenerateEddsaPrivateKey(seed)
	if err != nil {
		panic(err)
	}
	assetAMinAmount := big.NewInt(98)
	assetBMinAmount := big.NewInt(99)
	lpAmount := big.NewInt(100)
	expiredAt := time.Now().Add(time.Hour * 2).UnixMilli()
	txInfo := &types2.RemoveLiquidityTxInfo{
		FromAccountIndex:  getAccountIndex(s, accountName),
		PairIndex:         0,
		AssetAId:          0,
		AssetAMinAmount:   assetAMinAmount,
		AssetBId:          2,
		AssetBMinAmount:   assetBMinAmount,
		LpAmount:          lpAmount,
		AssetAAmountDelta: nil,
		AssetBAmountDelta: nil,
		GasAccountIndex:   1,
		GasFeeAssetId:     2,
		GasFeeAssetAmount: big.NewInt(5000),
		ExpiredAt:         expiredAt,
		Nonce:             getNextNonce(s, accountName),
		Sig:               nil,
	}
	hFunc := mimc.NewMiMC()
	msgHash, err := legendTxTypes.ComputeRemoveLiquidityMsgHash(txInfo, hFunc)
	if err != nil {
		panic(err)
	}
	hFunc.Reset()
	signature, err := key.Sign(msgHash, hFunc)
	if err != nil {
		panic(err)
	}
	txInfo.Sig = signature
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		panic(err)
	}
	return string(txInfoBytes)
}

func constructSwapTxInfo(s *AppSuite) string {
	key, err := curve.GenerateEddsaPrivateKey(seed)
	if err != nil {
		panic(err)
	}
	assetAAmount := big.NewInt(100)
	assetBAmount := big.NewInt(98)
	expiredAt := time.Now().Add(time.Hour * 2).UnixMilli()
	txInfo := &types2.SwapTxInfo{
		FromAccountIndex:  getAccountIndex(s, accountName),
		PairIndex:         0,
		AssetAId:          2,
		AssetAAmount:      assetAAmount,
		AssetBId:          0,
		AssetBMinAmount:   assetBAmount,
		AssetBAmountDelta: nil,
		GasAccountIndex:   1,
		GasFeeAssetId:     0,
		GasFeeAssetAmount: big.NewInt(5000),
		ExpiredAt:         expiredAt,
		Nonce:             getNextNonce(s, accountName),
		Sig:               nil,
	}
	hFunc := mimc.NewMiMC()
	msgHash, err := legendTxTypes.ComputeSwapMsgHash(txInfo, hFunc)
	if err != nil {
		panic(err)
	}
	hFunc.Reset()
	signature, err := key.Sign(msgHash, hFunc)
	if err != nil {
		panic(err)
	}
	txInfo.Sig = signature
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		panic(err)
	}
	return string(txInfoBytes)
}

func constructTransferNftTxInfo(s *AppSuite) string {
	seed := "17673b9a9fdec6dc90c7cc1eb1c939134dfb659d2f08edbe071e5c45f343d008"
	key, err := curve.GenerateEddsaPrivateKey(seed)
	if err != nil {
		panic(err)
	}
	nameHash, err := common2.AccountNameHash("sher.legend")
	if err != nil {
		panic(err)
	}
	expiredAt := time.Now().Add(time.Hour * 2).UnixMilli()
	txInfo := &types2.TransferNftTxInfo{
		FromAccountIndex:  getAccountIndex(s, "gavin.legend"),
		ToAccountIndex:    getAccountIndex(s, accountName),
		ToAccountNameHash: nameHash,
		NftIndex:          1,
		GasAccountIndex:   1,
		GasFeeAssetId:     0,
		GasFeeAssetAmount: big.NewInt(5000),
		CallData:          "",
		CallDataHash:      nil,
		ExpiredAt:         expiredAt,
		Nonce:             getNextNonce(s, "gavin.legend"),
		Sig:               nil,
	}
	hFunc := mimc.NewMiMC()
	hFunc.Write([]byte(txInfo.CallData))
	callDataHash := hFunc.Sum(nil)
	txInfo.CallDataHash = callDataHash
	hFunc.Reset()
	msgHash, err := legendTxTypes.ComputeTransferNftMsgHash(txInfo, hFunc)
	if err != nil {
		panic(err)
	}
	hFunc.Reset()
	signature, err := key.Sign(msgHash, hFunc)
	if err != nil {
		panic(err)
	}
	txInfo.Sig = signature
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		panic(err)
	}
	return string(txInfoBytes)
}

func constructTransferTxInfo(s *AppSuite) string {
	key, err := curve.GenerateEddsaPrivateKey(seed)
	if err != nil {
		panic(err)
	}
	nameHash, err := common2.AccountNameHash("gavin.legend")
	if err != nil {
		panic(err)
	}
	expiredAt := time.Now().Add(time.Hour * 2).UnixMilli()
	txInfo := &types2.TransferTxInfo{
		FromAccountIndex:  getAccountIndex(s, accountName),
		ToAccountIndex:    getAccountIndex(s, "gavin.legend"),
		ToAccountNameHash: nameHash,
		AssetId:           0,
		AssetAmount:       big.NewInt(100000),
		GasAccountIndex:   1,
		GasFeeAssetId:     2,
		GasFeeAssetAmount: big.NewInt(5000),
		Memo:              "transfer",
		CallData:          "",
		CallDataHash:      tree.NilHash,
		Nonce:             getNextNonce(s, accountName),
		ExpiredAt:         expiredAt,
		Sig:               nil,
	}
	hFunc := mimc.NewMiMC()
	hFunc.Write([]byte(txInfo.CallData))
	callDataHash := hFunc.Sum(nil)
	txInfo.CallDataHash = callDataHash
	hFunc.Reset()
	msgHash, err := legendTxTypes.ComputeTransferMsgHash(txInfo, hFunc)
	if err != nil {
		panic(err)
	}
	hFunc.Reset()
	signature, err := key.Sign(msgHash, hFunc)
	if err != nil {
		panic(err)
	}
	txInfo.Sig = signature
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		panic(err)
	}
	return string(txInfoBytes)
}

func constructSendNftTxInfo(s *AppSuite) string {
	key, err := curve.GenerateEddsaPrivateKey(seed)
	if err != nil {
		panic(err)
	}
	expiredAt := time.Now().Add(time.Hour * 2).UnixMilli()
	txInfo := &types2.WithdrawNftTxInfo{
		AccountIndex:      getAccountIndex(s, accountName),
		NftIndex:          1,
		ToAddress:         "0xd5Aa3B56a2E2139DB315CdFE3b34149c8ed09171",
		GasAccountIndex:   1,
		GasFeeAssetId:     0,
		GasFeeAssetAmount: big.NewInt(5000),
		ExpiredAt:         expiredAt,
		Nonce:             getNextNonce(s, accountName),
		Sig:               nil,
	}
	hFunc := mimc.NewMiMC()
	msgHash, err := legendTxTypes.ComputeWithdrawNftMsgHash(txInfo, hFunc)
	if err != nil {
		panic(err)
	}
	hFunc.Reset()
	signature, err := key.Sign(msgHash, hFunc)
	if err != nil {
		panic(err)
	}
	txInfo.Sig = signature
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		panic(err)
	}
	return string(txInfoBytes)
}

func constructWithdrawTxInfo(s *AppSuite) string {
	key, err := curve.GenerateEddsaPrivateKey(seed)
	if err != nil {
		panic(err)
	}
	expiredAt := time.Now().Add(time.Hour * 2).UnixMilli()
	txInfo := &types2.WithdrawTxInfo{
		FromAccountIndex:  getAccountIndex(s, accountName),
		AssetId:           0,
		AssetAmount:       big.NewInt(10000000),
		GasAccountIndex:   1,
		GasFeeAssetId:     2,
		GasFeeAssetAmount: big.NewInt(5000),
		ToAddress:         "0x99AC8881834797ebC32f185ee27c2e96842e1a47",
		Nonce:             getNextNonce(s, accountName),
		ExpiredAt:         expiredAt,
		Sig:               nil,
	}
	hFunc := mimc.NewMiMC()
	msgHash, err := legendTxTypes.ComputeWithdrawMsgHash(txInfo, hFunc)
	if err != nil {
		panic(err)
	}
	hFunc.Reset()
	signature, err := key.Sign(msgHash, hFunc)
	if err != nil {
		panic(err)
	}
	txInfo.Sig = signature
	txInfoBytes, err := json.Marshal(txInfo)
	if err != nil {
		panic(err)
	}
	return string(txInfoBytes)
}
