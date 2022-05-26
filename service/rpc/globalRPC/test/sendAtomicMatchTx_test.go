package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	curve "github.com/zecrey-labs/zecrey-crypto/ecc/ztwistededwards/tebn254"
	"github.com/zecrey-labs/zecrey-crypto/wasm/zecrey-legend/legendTxTypes"
	"github.com/zecrey-labs/zecrey-legend/common/commonAsset"
	"github.com/zecrey-labs/zecrey-legend/common/commonTx"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/globalRPCProto"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/config"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/server"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"math/big"
	"testing"
	"time"

	"github.com/zeromicro/go-zero/core/conf"
)

// /Users/gavin/Desktop/zecrey-v2
func TestSendAtomicMatchTx(t *testing.T) {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	logx.MustSetup(c.LogConf)
	ctx := svc.NewServiceContext(c)

	/*
		err := globalmapHandler.ReloadGlobalMap(ctx)
		if err != nil {
			logx.Error("[main] %s", err.Error())
			return
		}
	*/

	srv := server.NewGlobalRPCServer(ctx)
	txInfo := constructSendAtomicMatchTxInfo()
	resp, err := srv.SendTx(
		context.Background(),
		&globalRPCProto.ReqSendTx{
			TxType: commonTx.TxTypeAtomicMatch,
			TxInfo: txInfo,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	respBytes, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(respBytes))
	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
}

func constructSendAtomicMatchTxInfo() string {
	// from sher.legend to gavin.legend
	sherSeed := "28e1a3762ff9944e9a4ad79477b756ef0aff3d2af76f0f40a0c3ec6ca76cf24b"
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
	buyOffer := &commonTx.OfferTxInfo{
		Type:         commonAsset.BuyOfferType,
		OfferId:      0,
		AccountIndex: 3,
		NftIndex:     0,
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
	sellOffer := &commonTx.OfferTxInfo{
		Type:         commonAsset.SellOfferType,
		OfferId:      0,
		AccountIndex: 2,
		NftIndex:     0,
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
	txInfo := &commonTx.AtomicMatchTxInfo{
		AccountIndex:      2,
		BuyOffer:          buyOffer,
		SellOffer:         sellOffer,
		GasAccountIndex:   1,
		GasFeeAssetId:     0,
		GasFeeAssetAmount: big.NewInt(5000),
		Nonce:             8,
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
