package generateproof

import (
	"github.com/bnb-chain/zkbnb/tools/desertexit/generateproof/config"
	"github.com/bnb-chain/zkbnb/tools/desertexit/generateproof/generateproof"
	"github.com/goccy/go-json"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"
	"time"
)

const CommandRunGenerateProof = "run"
const CommandContinueGenerateProof = "continue"

func Run(configFile string, address string, token string, nftIndexListStr string, proofFolder string) error {
	var c config.Config
	conf.MustLoad(configFile, &c)
	logx.MustSetup(c.LogConf)
	logx.DisableStat()
	proc.AddShutdownListener(func() {
		logx.Close()
	})
	if address != "" {
		c.Address = address
	}
	if token != "" {
		c.Token = token
	}
	if proofFolder != "" {
		c.ProofFolder = proofFolder
	}
	if nftIndexListStr != "" {
		var nftIndexList []int64
		err := json.Unmarshal([]byte(nftIndexListStr), &nftIndexList)
		if err != nil {
			return err
		}
		c.NftIndexList = nftIndexList
	}

	m, err := generateproof.NewMonitor(&c)
	if err != nil {
		logx.Severe(err)
		return err
	}

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

	desertExit, err := generateproof.NewDesertExit(&c)
	if err != nil {
		return err
	}
	err = desertExit.Run()
	if err != nil {
		logx.Severe(err)
		return err
	}
	return nil
}
