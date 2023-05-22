package rollbackwitnesssmt

import (
	"fmt"
	"github.com/bnb-chain/zkbnb/service/witness/config"
	"github.com/bnb-chain/zkbnb/service/witness/witness"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"
)

func RollbackWitnessSmt(
	configFile string,
	height int64,
) error {
	c := config.Config{}
	if err := config.InitSystemConfiguration(&c, configFile); err != nil {
		logx.Severef("failed to initiate system configuration, %v", err)
		panic("failed to initiate system configuration, err:" + err.Error())
	}
	logx.MustSetup(c.LogConf)
	logx.DisableStat()
	proc.AddShutdownListener(func() {
		logx.Close()
	})

	if !c.EnableRollback {
		return fmt.Errorf("rollback switch not turned on")
	}

	w, err := witness.NewWitness(c, false)
	if err != nil {
		return fmt.Errorf("failed to create witness instance, %v", err)
	}

	toHeight, err := w.BlockModel.GetCurrentBlockHeight()
	if err != nil {
		return fmt.Errorf("get current block height failed: %s", err.Error())
	}
	logx.Infof("get current block height: %d", toHeight)

	err = w.Rollback(height, toHeight)
	if err != nil {
		return fmt.Errorf("failed to rollback smt, %v", err)
	}
	logx.Infof("rollback smt success,the new smt version is %d now", height-1)
	return nil
}
