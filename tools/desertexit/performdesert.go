package desertexit

import (
	"encoding/json"
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

func Perform(configFile string, command string, amount string, nftIndexListStr string, owner string, privateKey string, proof string, token string) error {
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
	if owner != "" {
		c.Address = owner
	}
	if token != "" {
		c.Token = token
	}
	if nftIndexListStr != "" {
		var nftIndexList []int64
		err := json.Unmarshal([]byte(nftIndexListStr), &nftIndexList)
		if err != nil {
			return err
		}
		c.NftIndexList = nftIndexList
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
		err = m.CancelOutstandingDeposit(c.Address)
		if err != nil {
			logx.Severe(err)
			return err
		}
		break
	case CommandWithdrawNFT:
		err = m.WithdrawPendingNFTBalance(c.NftIndexList)
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
		err = m.WithdrawPendingBalance(common.HexToAddress(c.Address), common.HexToAddress(c.Token), bigIntAmount)
		if err != nil {
			logx.Severe(err)
			return err
		}
		break
	case CommandGetBalance:
		_, err := m.GetBalance(common.HexToAddress(c.Address), common.HexToAddress(c.Token))
		if err != nil {
			logx.Severe(err)
			return err
		}
		break
	case CommandGetPendingBalance:
		_, err := m.GetPendingBalance(common.HexToAddress(c.Address), common.HexToAddress(c.Token))
		if err != nil {
			logx.Severe(err)
			return err
		}
		break
	default:
		logx.Severef("no %s  command", command)
	}
	return nil
}
