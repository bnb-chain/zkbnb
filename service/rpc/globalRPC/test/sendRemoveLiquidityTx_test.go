package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	curve "github.com/bnb-chain/zkbas-crypto/ecc/ztwistededwards/tebn254"
	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/globalRPCProto"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/config"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/server"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/svc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/zeromicro/go-zero/core/logx"
	"math/big"
	"testing"
	"time"

	"github.com/zeromicro/go-zero/core/conf"
)

func TestSendRemoveLiquidityTx(t *testing.T) {
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
	txInfo := constructSendRemoveLiquidityTxInfo()
	resp, err := srv.SendTx(
		context.Background(),
		&globalRPCProto.ReqSendTx{
			TxType: commonTx.TxTypeRemoveLiquidity,
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

func constructSendRemoveLiquidityTxInfo() string {
	// from sher.legend to gavin.legend
	seed := "28e1a3762ff9944e9a4ad79477b756ef0aff3d2af76f0f40a0c3ec6ca76cf24b"
	key, err := curve.GenerateEddsaPrivateKey(seed)
	if err != nil {
		panic(err)
	}
	assetAMinAmount := big.NewInt(98)
	assetBMinAmount := big.NewInt(99)
	lpAmount := big.NewInt(100)
	expiredAt := time.Now().Add(time.Hour * 2).UnixMilli()
	txInfo := &commonTx.RemoveLiquidityTxInfo{
		FromAccountIndex:  2,
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
		Nonce:             5,
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
