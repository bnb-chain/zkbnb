package performexodus

import (
	"github.com/bnb-chain/zkbnb/tools/exodusexit/generateproof/generateproof"
	"github.com/bnb-chain/zkbnb/tools/exodusexit/performexodus/config"
	"github.com/bnb-chain/zkbnb/tools/exodusexit/performexodus/performexodus"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"math/big"
)

func Run(configFile string, command string, amount string, nftIndex string, owner string, privateKey string, proof string, token string) error {
	var c config.Config
	conf.MustLoad(configFile, &c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()
	if privateKey != "" {
		c.ChainConfig.PrivateKey = privateKey
	}
	m, err := performexodus.NewPerformExodus(c)
	if err != nil {
		logx.Severe(err)
		return err
	}

	if command == "activateDesert" {
		err = m.ActivateDesertMode()
		if err != nil {
			return err
		}
	}
	if command == "performAsset" {
		var performDesertAsset generateproof.PerformDesertAssetData
		conf.MustLoad(proof, &performDesertAsset)
		err = m.PerformDesert(performDesertAsset)
		if err != nil {
			return err
		}
	}
	if command == "performNft" {
		var performDesertNftData generateproof.PerformDesertNftData
		conf.MustLoad(proof, &performDesertNftData)
		err = m.PerformDesertNft(performDesertNftData)
		if err != nil {
			return err
		}
	}
	if command == "cancelOutstandingDeposit" {
		err = m.CancelOutstandingDeposit()
		if err != nil {
			return err
		}
	}
	if command == "withdrawNFT" {
		bigIntNftIndex, success := new(big.Int).SetString(nftIndex, 10)
		if !success {
			logx.Severe("failed to transfer big int")
			return nil
		}
		err = m.WithdrawPendingNFTBalance(bigIntNftIndex)
		if err != nil {
			return err
		}
	}
	if command == "withdrawAsset" {
		bigIntAmount, success := new(big.Int).SetString(amount, 10)
		if !success {
			logx.Severe("failed to transfer big int")
			return nil
		}
		err = m.WithdrawPendingBalance(common.HexToAddress(owner), common.HexToAddress(token), bigIntAmount)
		if err != nil {
			return err
		}
	}
	return nil
}
