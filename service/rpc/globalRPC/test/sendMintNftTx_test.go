package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	curve "github.com/bnb-chain/zkbas-crypto/ecc/ztwistededwards/tebn254"
	"github.com/bnb-chain/zkbas-crypto/wasm/legend/legendTxTypes"
	"github.com/bnb-chain/zkbas/common/commonTx"
	"github.com/bnb-chain/zkbas/common/util"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/globalRPCProto"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/config"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/server"
	"github.com/bnb-chain/zkbas/service/rpc/globalRPC/internal/svc"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/logx"
	"math/big"
	"testing"
	"time"

	"github.com/zeromicro/go-zero/core/conf"
)

func TestSendMintNftTx(t *testing.T) {
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
	txInfo := constructSendMintNftTxInfo()
	resp, err := srv.SendTx(
		context.Background(),
		&globalRPCProto.ReqSendTx{
			TxType: commonTx.TxTypeMintNft,
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

func constructSendMintNftTxInfo() string {
	// from sher.legend to gavin.legend
	seed := "28e1a3762ff9944e9a4ad79477b756ef0aff3d2af76f0f40a0c3ec6ca76cf24b"
	key, err := curve.GenerateEddsaPrivateKey(seed)
	if err != nil {
		panic(err)
	}
	nameHash, err := util.AccountNameHash("gavin.legend")
	if err != nil {
		panic(err)
	}
	hFunc := mimc.NewMiMC()
	hFunc.Write([]byte(util.RandomUUID()))
	contentHash := hFunc.Sum(nil)
	expiredAt := time.Now().Add(time.Hour * 2).UnixMilli()
	txInfo := &commonTx.MintNftTxInfo{
		CreatorAccountIndex: 2,
		ToAccountIndex:      3,
		ToAccountNameHash:   nameHash,
		NftContentHash:      common.Bytes2Hex(contentHash),
		NftCollectionId:     1,
		CreatorTreasuryRate: 0,
		GasAccountIndex:     1,
		GasFeeAssetId:       2,
		GasFeeAssetAmount:   big.NewInt(5000),
		ExpiredAt:           expiredAt,
		Nonce:               7,
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
