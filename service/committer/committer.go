package committer

import (
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/bnb-chain/zkbas/service/committer/committer"
)

func Run(configFile string) error {
	var config committer.Config
	conf.MustLoad(configFile, &config)

	committer, err := committer.NewCommitter(&config)
	if err != nil {
		logx.Error("new committer failed:", err)
		return err
	}

	committer.Run()
	return nil
}
