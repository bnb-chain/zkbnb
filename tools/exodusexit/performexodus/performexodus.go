package performexodus

import (
	"github.com/bnb-chain/zkbnb/tools/exodusexit/performexodus/config"
	"github.com/bnb-chain/zkbnb/tools/exodusexit/performexodus/performexodus"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

func Run(configFile string) error {
	var c config.Config
	conf.MustLoad(configFile, &c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()

	m, err := performexodus.NewPerformExodus(c)
	if err != nil {
		logx.Severe(err)
		panic(err)
	}
	//var performDesertAsset generateproof.PerformDesertAssetData
	//conf.MustLoad("./tools/exodusexit/proofdata/performDesertAsset.json", &performDesertAsset)
	//err = m.PerformDesert(performDesertAsset)
	//if err != nil {
	//	return err
	//}

	//var performDesertNftData generateproof.PerformDesertNftData
	//conf.MustLoad("./tools/exodusexit/proofdata/performDesertNft.json", &performDesertNftData)
	//err = m.PerformDesertNft(performDesertNftData)
	//if err != nil {
	//	return err
	//}

	//err = m.ActivateDesertMode()
	//if err != nil {
	//	return err
	//}

	//var nftIndex *big.Int
	//err = m.WithdrawPendingNFTBalance(nftIndex)
	//if err != nil {
	//	return err
	//}
	//
	//var owner common.Address
	//var token common.Address
	//var amount *big.Int
	//err = m.WithdrawPendingBalance(owner, token, amount)
	//if err != nil {
	//	return err
	//}

	err = m.CancelOutstandingDeposit()
	if err != nil {
		return err
	}
	return nil
}
