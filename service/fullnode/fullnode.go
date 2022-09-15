package fullnode

import (
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbnb/service/fullnode/fullnode"
)

func Run(configFile string) error {
	var config fullnode.Config
	conf.MustLoad(configFile, &config)

	fullnode, err := fullnode.Newfullnode(&config)
	if err != nil {
		logx.Error("new fullnode failed:", err)
		return err
	}

	fullnode.Run()
	return nil
}
