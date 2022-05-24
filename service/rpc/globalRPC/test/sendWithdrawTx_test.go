package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	curve "github.com/zecrey-labs/zecrey-crypto/ecc/ztwistededwards/tebn254"
	"github.com/zecrey-labs/zecrey-crypto/wasm/zecrey-legend/legendTxTypes"
	"github.com/zecrey-labs/zecrey-legend/common/commonTx"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/globalRPCProto"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/config"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/server"
	"github.com/zecrey-labs/zecrey-legend/service/rpc/globalRPC/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"math/big"
	"testing"

	"github.com/zeromicro/go-zero/core/conf"
)

// /Users/gavin/Desktop/zecrey-v2
func TestSendWithdrawTx(t *testing.T) {
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
	txInfo := constructSendWithdrawTxInfo()
	resp, err := srv.SendTx(
		context.Background(),
		&globalRPCProto.ReqSendTx{
			TxType: commonTx.TxTypeWithdraw,
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

func constructSendWithdrawTxInfo() string {
	// from sher.legend to gavin.legend
	seed := "28e1a3762ff9944e9a4ad79477b756ef0aff3d2af76f0f40a0c3ec6ca76cf24b"
	key, err := curve.GenerateEddsaPrivateKey(seed)
	if err != nil {
		panic(err)
	}
	txInfo := &commonTx.WithdrawTxInfo{
		FromAccountIndex:  3,
		AssetId:           0,
		AssetAmount:       big.NewInt(10000),
		GasAccountIndex:   2,
		GasFeeAssetId:     0,
		GasFeeAssetAmount: big.NewInt(50),
		ToAddress:         "0x99AC8881834797ebC32f185ee27c2e96842e1a47",
		Nonce:             1,
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
