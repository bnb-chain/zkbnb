package performexodus

import (
	"github.com/bnb-chain/zkbnb/tools/exodusexit/generateproof/generateproof"
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
	var performDesertAsset generateproof.PerformDesertData
	conf.MustLoad("./tools/exodusexit/proofdata/performDesertAsset.json", &performDesertAsset)
	err = m.PerformDesert(performDesertAsset.NftRoot, performDesertAsset.ExitData, performDesertAsset.AssetMerkleProof, performDesertAsset.AccountMerkleProof)
	if err != nil {
		return err
	}
	return nil
}
