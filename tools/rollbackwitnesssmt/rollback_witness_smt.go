package rollbackwitnesssmt

import (
	"fmt"
	"github.com/bnb-chain/zkbnb/service/witness/config"
	"github.com/bnb-chain/zkbnb/service/witness/witness"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"
)

func RollbackWitnessSmt(
	configFile string,
	height int64,
) error {
	var c config.Config
	conf.MustLoad(configFile, &c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()
	proc.AddShutdownListener(func() {
		logx.Close()
	})

	w, err := witness.NewWitness(c)
	if err != nil {
		return fmt.Errorf("failed to create witness instance, %v", err)
	}
	err = w.Rollback(height)
	if err != nil {
		return fmt.Errorf("failed to rollback smt, %v", err)
	}
	return nil
}
