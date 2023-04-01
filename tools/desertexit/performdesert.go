package desertexit

import (
	"github.com/bnb-chain/zkbnb/tools/desertexit/config"
	"github.com/bnb-chain/zkbnb/tools/desertexit/desertexit"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"
	"math/big"
)

const CommandActivateDesert = "activateDesert"
const CommandPerformAsset = "performAsset"
const CommandPerformNft = "performNft"
const CommandCancelOutstandingDeposit = "cancelOutstandingDeposit"
const CommandWithdrawNFT = "withdrawNFT"
const CommandWithdrawAsset = "withdrawAsset"
const CommandGetBalance = "getBalance"
const CommandGetPendingBalance = "getPendingBalance"

func Perform(configFile string, command string, amount string, nftIndex string, owner string, privateKey string, proof string, token string) error {
	var c config.Config
	conf.MustLoad(configFile, &c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()
	proc.AddShutdownListener(func() {
		logx.Close()
	})
	if privateKey != "" {
		c.ChainConfig.PrivateKey = privateKey
	}
	m, err := desertexit.NewPerformDesert(c)
	if err != nil {
		logx.Severe(err)
		return err
	}
	switch command {
	case CommandActivateDesert:
		err = m.ActivateDesertMode()
		if err != nil {
			logx.Severe(err)
			return err
		}
		break
	case CommandPerformAsset:
		var performDesertAsset desertexit.PerformDesertAssetData
		conf.MustLoad(proof, &performDesertAsset)
		err = m.PerformDesert(performDesertAsset)
		if err != nil {
			logx.Severe(err)
			return err
		}
		break
	case CommandPerformNft:
		var performDesertNftData desertexit.PerformDesertNftData
		conf.MustLoad(proof, &performDesertNftData)
		err = m.PerformDesertNft(performDesertNftData)
		if err != nil {
			logx.Severe(err)
			return err
		}
		break
	case CommandCancelOutstandingDeposit:
		err = m.CancelOutstandingDeposit("")
		if err != nil {
			logx.Severe(err)
			return err
		}
		break
	case CommandWithdrawNFT:
		bigIntNftIndex, success := new(big.Int).SetString(nftIndex, 10)
		if !success {
			logx.Severe("failed to transfer big int")
			return nil
		}
		err = m.WithdrawPendingNFTBalance(bigIntNftIndex)
		if err != nil {
			logx.Severe(err)
			return err
		}
		break
	case CommandWithdrawAsset:
		bigIntAmount, success := new(big.Int).SetString(amount, 10)
		if !success {
			logx.Severe("failed to transfer big int")
			return nil
		}
		err = m.WithdrawPendingBalance(common.HexToAddress(owner), common.HexToAddress(token), bigIntAmount)
		if err != nil {
			logx.Severe(err)
			return err
		}
		break
	case CommandGetBalance:
		_, err := m.GetBalance(common.HexToAddress(owner), common.HexToAddress(token))
		if err != nil {
			logx.Severe(err)
			return err
		}
		break
	case CommandGetPendingBalance:
		_, err := m.GetPendingBalance(common.HexToAddress(owner), common.HexToAddress(token))
		if err != nil {
			logx.Severe(err)
			return err
		}
		break
	}
	return nil
}
