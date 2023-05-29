package desertexit

import (
	"fmt"
	desertTypes "github.com/bnb-chain/zkbnb-crypto/circuit/desert/types"
	"github.com/bnb-chain/zkbnb/tools/desertexit/config"
	"github.com/bnb-chain/zkbnb/tools/desertexit/desertexit"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"
	"strconv"
	"time"
)

const GracefulShutdownTimeout = 10 * time.Second
const CommandRunGenerateProof = "run"
const CommandContinueGenerateProof = "continue"

func Run(configFile string, address string, token string, nftIndex string, proofFolder string) error {
	var c config.Config
	conf.MustLoad(configFile, &c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()

	if address != "" {
		c.Address = address
	}
	if token != "" {
		c.Token = token
	}
	if proofFolder != "" {
		c.ProofFolder = proofFolder
	}
	if nftIndex != "" {
		var err error
		c.NftIndex, err = strconv.ParseInt(nftIndex, 10, 64)
		if err != nil {
			logx.Severe(err)
			return err
		}
	}

	if c.Address == "" || (c.Token == "" && c.NftIndex <= 0) || (c.Token != "" && c.NftIndex > 0) {
		logx.Severe("invalid parameter")
		return fmt.Errorf("invalid parameter")
	}

	var txType uint8
	if c.Token != "" {
		txType = desertTypes.TxTypeExit
	} else {
		txType = desertTypes.TxTypeExitNft
	}

	m, err := desertexit.NewDesertExit(&c)
	if err != nil {
		logx.Severe(err)
		return err
	}

	proc.SetTimeToForceQuit(GracefulShutdownTimeout)
	proc.AddShutdownListener(func() {
		logx.Info("start to shutdown desertexit......")
		m.Shutdown()
		_ = logx.Close()
	})

	go func() {
		for {
			err := m.MonitorGenericBlocks()
			if err != nil {
				logx.Severef("monitor blocks error, %v", err)
			} else {
				break
			}
		}
	}()

	go func() {
		for {
			err := m.CleanHistoryBlocks()
			if err != nil {
				logx.Severef("clear history blocks error, %v", err)
			}
			time.Sleep(1 * time.Minute)
		}
	}()

	desertExit, err := desertexit.NewGenerateProof(&c)
	if err != nil {
		logx.Severe(err)
		return err
	}

	blockHeight, err := desertExit.Run()
	if err != nil {
		logx.Severe(err)
		return err
	}
	logx.Info("execute all the l2 blocks successfully")

	err = desertExit.GenerateProof(blockHeight, txType)
	if err != nil {
		logx.Severe(err)
		return err
	}
	logx.Info("generate proof successfully")
	return nil
}
